# Database Package

Package này cung cấp PostgreSQL database connection sử dụng `pgx/v5` với connection pooling.

## Features

- ✅ Connection pooling với `pgxpool`
- ✅ Multi-schema support (identify, catalog, community, payment)
- ✅ Query logging (optional)
- ✅ Health check
- ✅ Graceful shutdown
- ✅ Transaction helper
- ✅ Context-based operations
- ✅ Connection statistics

## Khởi tạo

```go
import (
    "context"
    "time"

    "system/configs"
    "system/internal/platform/database"
)

// Load config
cfg, _ := configs.LoadConfig(".env")

// Khởi tạo database
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

db, err := database.NewPostgresDB(ctx, &cfg.DB, cfg.Log.DBLogQueries, logger)
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

## Configuration

Database configuration trong `.env`:

```env
DB_HOST=localhost
DB_PORT=5432
DB_NAME=system_dev
DB_USER=system_dev
DB_PASSWORD=system_dev
DB_SSL_MODE=disable
DB_MAX_OPEN_CONNS=25        # Số connection tối đa
DB_MAX_IDLE_CONNS=5         # Số idle connection tối thiểu
DB_CONN_MAX_LIFETIME=5m     # Lifetime của mỗi connection

# Schemas
DB_IDENTIFY_SCHEMA=identify
DB_CATALOG_SCHEMA=catalog
DB_COMMUNITY_SCHEMA=community
DB_PAYMENT_SCHEMA=payment

# Logging
DB_LOG_QUERIES=false        # Bật/tắt query logging
```

## Connection Pool Parameters

- **MaxConns**: Số lượng connection tối đa trong pool (default: 25)
- **MinConns**: Số lượng idle connection tối thiểu được giữ (default: 5)
- **MaxConnLifetime**: Thời gian tối đa một connection tồn tại trước khi refresh (default: 5m)
- **MaxConnIdleTime**: Thời gian tối đa một connection idle trước khi đóng (default: 30m)
- **HealthCheckPeriod**: Interval check health của idle connections (default: 1m)

## Sử dụng

### 1. Query Rows

```go
ctx := context.Background()
query := `SELECT id, email, name FROM identify.users WHERE status = $1`

rows, err := db.Pool.Query(ctx, query, "active")
if err != nil {
    return err
}
defer rows.Close()

for rows.Next() {
    var id int64
    var email, name string

    if err := rows.Scan(&id, &email, &name); err != nil {
        continue
    }

    // Process data...
}
```

### 2. Query Single Row

```go
var email string
query := `SELECT email FROM identify.users WHERE id = $1`

err := db.Pool.QueryRow(ctx, query, 123).Scan(&email)
if err == pgx.ErrNoRows {
    // Không tìm thấy
}
```

### 3. Insert/Update/Delete

```go
var productID int64
query := `
    INSERT INTO catalog.products (name, price)
    VALUES ($1, $2)
    RETURNING id
`

err := db.Pool.QueryRow(ctx, query, "Product", 99.99).Scan(&productID)
```

### 4. Transaction

```go
err := db.WithTransaction(ctx, func(tx pgx.Tx) error {
    // Insert user
    var userID int64
    err := tx.QueryRow(ctx,
        `INSERT INTO identify.users (email, name) VALUES ($1, $2) RETURNING id`,
        "user@example.com", "John Doe",
    ).Scan(&userID)
    if err != nil {
        return err // Tự động rollback
    }

    // Insert profile
    _, err = tx.Exec(ctx,
        `INSERT INTO identify.user_profiles (user_id, bio) VALUES ($1, $2)`,
        userID, "Bio",
    )
    if err != nil {
        return err // Tự động rollback
    }

    return nil // Tự động commit
})
```

### 5. Dynamic Schema

```go
// Lấy schema name theo domain
schema := db.GetSchemaName("identify") // Returns "identify"

// Sử dụng trong query
query := fmt.Sprintf(`SELECT * FROM %s.users`, schema)
```

### 6. Health Check

```go
healthInfo, err := db.Health(ctx)
if err != nil {
    log.Error(err)
}

// healthInfo contains:
// - status: "healthy"
// - max_connections: 25
// - open_connections: 6
// - in_use: 0
// - idle: 6
// - wait_count: 0
```

### 7. Pool Statistics

```go
stats := db.Stats()

log.Info("Pool stats",
    "total", stats.TotalConns(),
    "acquired", stats.AcquiredConns(),
    "idle", stats.IdleConns(),
)
```

## Query Logging

Bật query logging trong `.env`:

```env
DB_LOG_QUERIES=true
```

Khi bật, tất cả queries sẽ được log ra với level INFO:

```
2025-10-26T14:18:21.299+0700  INFO  Query  {"sql": "SELECT * FROM users WHERE id = $1", "args": [123], "time": "2.5ms"}
```

## Best Practices

### 1. Luôn dùng Context với Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

rows, err := db.Pool.Query(ctx, query, args...)
```

### 2. Đóng Rows sau khi Query

```go
rows, err := db.Pool.Query(ctx, query)
if err != nil {
    return err
}
defer rows.Close() // QUAN TRỌNG!

for rows.Next() {
    // ...
}
```

### 3. Dùng Parameterized Queries (tránh SQL Injection)

```go
// ✅ ĐÚNG
query := `SELECT * FROM users WHERE email = $1`
rows, _ := db.Pool.Query(ctx, query, userEmail)

// ❌ SAI - SQL Injection risk!
query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", userEmail)
```

### 4. Dùng Transaction cho Multiple Operations

```go
// Nếu cần insert/update nhiều bảng, dùng transaction
err := db.WithTransaction(ctx, func(tx pgx.Tx) error {
    // Multiple operations here
    return nil
})
```

### 5. Monitor Connection Pool

```go
// Định kỳ check pool stats
ticker := time.NewTicker(1 * time.Minute)
go func() {
    for range ticker.C {
        stats := db.Stats()
        if stats.TotalConns() >= stats.MaxConns() {
            log.Warn("Connection pool exhausted!")
        }
    }
}()
```

## Graceful Shutdown

```go
func main() {
    db, _ := database.NewPostgresDB(ctx, &cfg.DB, false, logger)
    defer db.Close() // Đóng pool khi shutdown

    // Setup signal handling
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    <-quit
    log.Info("Shutting down...")
    // db.Close() sẽ được gọi bởi defer
}
```

## Troubleshooting

### Connection Pool Exhausted

Nếu thấy error "connection pool exhausted":

1. Tăng `DB_MAX_OPEN_CONNS` trong `.env`
2. Kiểm tra xem có leak connections không (rows không được close)
3. Check pool stats để monitor

### Slow Queries

Bật query logging để debug:

```env
DB_LOG_QUERIES=true
```

### Connection Timeout

Tăng timeout khi khởi tạo database:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
```

## Examples

Xem thêm examples trong file `example_usage.go`.
