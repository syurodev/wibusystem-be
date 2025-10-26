# Redis Package

Package này cung cấp Redis client sử dụng `go-redis/v9` với connection pooling và helper functions.

## Features

- ✅ Connection pooling với go-redis/v9
- ✅ Health check
- ✅ Connection statistics
- ✅ Helper functions cho common operations
- ✅ Graceful shutdown
- ✅ Context-based operations
- ✅ Retry strategy
- ✅ Auto-reconnect

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

// Khởi tạo Redis
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

rdb, err := database.NewRedisClient(ctx, &cfg.Redis, logger)
if err != nil {
    log.Fatal(err)
}
defer rdb.Close()
```

## Configuration

Redis configuration trong `.env`:

```env
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=your_password_here
REDIS_DB=0                  # Database number (0-15)
REDIS_MAX_RETRIES=3         # Số lần retry khi command fail
REDIS_POOL_SIZE=10          # Số connections tối đa trong pool
REDIS_MIN_IDLE_CONNS=5      # Số idle connections tối thiểu
```

## Connection Pool Parameters

- **PoolSize**: Số lượng connections tối đa trong pool (default: 10)
- **MinIdleConns**: Số lượng idle connections tối thiểu (default: 5)
- **DialTimeout**: Timeout khi tạo connection mới (5s)
- **ReadTimeout**: Timeout khi đọc data (3s)
- **WriteTimeout**: Timeout khi ghi data (3s)
- **PoolTimeout**: Timeout khi chờ connection từ pool (4s)
- **ConnMaxIdleTime**: Đóng idle connections sau 5 phút
- **ConnMaxLifetime**: Đóng connections sau 30 phút
- **MaxRetries**: Số lần retry tối đa (default: 3)

## Basic Operations

### 1. String Operations

```go
ctx := context.Background()

// Set key-value với expiration
err := rdb.Set(ctx, "user:email:123", "user@example.com", 1*time.Hour)

// Get value
email, err := rdb.Get(ctx, "user:email:123")
if err == redis.Nil {
    // Key không tồn tại
} else if err != nil {
    // Error khác
}

// Delete key
err = rdb.Del(ctx, "user:email:123")
```

### 2. Hash Operations

```go
// Set multiple fields
err := rdb.HSet(ctx, "user:profile:123",
    "name", "John Doe",
    "email", "john@example.com",
    "age", "30",
)

// Get single field
name, err := rdb.HGet(ctx, "user:profile:123", "name")

// Get all fields
profile, err := rdb.HGetAll(ctx, "user:profile:123")
// Returns: map[string]string{"name": "John Doe", "email": "...", "age": "30"}
```

### 3. Counter Operations

```go
// Increment
views, err := rdb.Incr(ctx, "page:views:homepage")
// Returns: 1, 2, 3, ...

// Decrement
stock, err := rdb.Decr(ctx, "product:stock:456")
```

### 4. Key Expiration

```go
// Check if key exists
exists, err := rdb.Exists(ctx, "session:abc")

// Set expiration
err = rdb.Expire(ctx, "session:abc", 30*time.Minute)

// Get TTL
ttl, err := rdb.TTL(ctx, "session:abc")
// Returns: duration (e.g., 29m59s)
```

## Common Use Cases

### 1. Caching với JSON

```go
type User struct {
    ID    int64  `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// Cache user data
user := User{ID: 123, Name: "John", Email: "john@example.com"}
userJSON, _ := json.Marshal(user)
err := rdb.Set(ctx, "user:cache:123", userJSON, 30*time.Minute)

// Retrieve from cache
cachedJSON, err := rdb.Get(ctx, "user:cache:123")
var cachedUser User
json.Unmarshal([]byte(cachedJSON), &cachedUser)
```

### 2. Session Management

```go
type Session struct {
    UserID    int64     `json:"user_id"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

sessionID := "sess_abc123"
session := Session{UserID: 123, Email: "user@example.com", CreatedAt: time.Now()}

// Store session (expires in 24h)
sessionJSON, _ := json.Marshal(session)
err := rdb.Set(ctx, "session:"+sessionID, sessionJSON, 24*time.Hour)

// Retrieve session
sessionData, err := rdb.Get(ctx, "session:"+sessionID)
if err == redis.Nil {
    // Session expired
}

// Logout (invalidate session)
err = rdb.Del(ctx, "session:"+sessionID)
```

### 3. Rate Limiting

```go
func CheckRateLimit(rdb *RedisClient, userID string) (bool, error) {
    ctx := context.Background()
    key := fmt.Sprintf("rate_limit:%s", userID)

    // Increment request count
    count, err := rdb.Incr(ctx, key)
    if err != nil {
        return false, err
    }

    // Set expiration on first request
    if count == 1 {
        rdb.Expire(ctx, key, 1*time.Minute)
    }

    // Check limit (100 requests per minute)
    if count > 100 {
        return false, fmt.Errorf("rate limit exceeded")
    }

    return true, nil
}
```

### 4. Cache-Aside Pattern

```go
func GetUser(rdb *RedisClient, userID int64) (*User, error) {
    ctx := context.Background()
    cacheKey := fmt.Sprintf("user:%d", userID)

    // 1. Try cache first
    cachedData, err := rdb.Get(ctx, cacheKey)
    if err == nil {
        // Cache hit
        var user User
        json.Unmarshal([]byte(cachedData), &user)
        return &user, nil
    }

    // 2. Cache miss - fetch from DB
    user, err := fetchFromDB(userID)
    if err != nil {
        return nil, err
    }

    // 3. Store in cache
    userJSON, _ := json.Marshal(user)
    rdb.Set(ctx, cacheKey, userJSON, 1*time.Hour)

    return user, nil
}
```

### 5. Distributed Lock (using direct client)

```go
func AcquireLock(rdb *RedisClient, lockKey string, ttl time.Duration) (bool, error) {
    ctx := context.Background()

    // Try to acquire lock
    acquired, err := rdb.Client.SetNX(ctx, lockKey, "locked", ttl).Result()
    if err != nil {
        return false, err
    }

    return acquired, nil
}

func ReleaseLock(rdb *RedisClient, lockKey string) error {
    return rdb.Del(context.Background(), lockKey)
}
```

## Advanced Operations (Direct Client Access)

### List Operations

```go
// Push to list
err := rdb.Client.RPush(ctx, "queue:jobs", "job1", "job2").Err()

// Pop from list
job, err := rdb.Client.LPop(ctx, "queue:jobs").Result()
```

### Set Operations

```go
// Add to set
err := rdb.Client.SAdd(ctx, "tags:123", "electronics", "laptop").Err()

// Check membership
isMember, err := rdb.Client.SIsMember(ctx, "tags:123", "laptop").Result()

// Get all members
members, err := rdb.Client.SMembers(ctx, "tags:123").Result()
```

### Sorted Set Operations

```go
// Add to sorted set
err := rdb.Client.ZAdd(ctx, "leaderboard", redis.Z{Score: 100, Member: "player1"}).Err()

// Get rank
rank, err := rdb.Client.ZRank(ctx, "leaderboard", "player1").Result()

// Get top N
top, err := rdb.Client.ZRevRangeWithScores(ctx, "leaderboard", 0, 9).Result()
```

### Pipeline (Batch Operations)

```go
pipe := rdb.Client.Pipeline()

pipe.Set(ctx, "key1", "value1", 0)
pipe.Set(ctx, "key2", "value2", 0)
pipe.Incr(ctx, "counter")

// Execute all at once
_, err := pipe.Exec(ctx)
```

### Pub/Sub

```go
// Publish
err := rdb.Client.Publish(ctx, "notifications", "New order").Err()

// Subscribe (usually in goroutine)
pubsub := rdb.Client.Subscribe(ctx, "notifications")
defer pubsub.Close()

ch := pubsub.Channel()
for msg := range ch {
    fmt.Printf("Received: %s\n", msg.Payload)
}
```

## Health Check

```go
healthInfo, err := rdb.Health(ctx)
if err != nil {
    log.Error(err)
}

// healthInfo contains:
// - status: "healthy"
// - latency: "2.5ms"
// - pool: {total_conns, idle_conns, hits, misses, timeouts}
// - config: {db, pool_size, min_idle_conns}
```

## Connection Statistics

```go
stats := rdb.Stats()

log.Info("Redis pool stats",
    "total_conns", stats.TotalConns,
    "idle_conns", stats.IdleConns,
    "stale_conns", stats.StaleConns,
    "hits", stats.Hits,
    "misses", stats.Misses,
    "timeouts", stats.Timeouts,
)
```

## Error Handling

```go
import "github.com/redis/go-redis/v9"

value, err := rdb.Get(ctx, "key")
if err == redis.Nil {
    // Key không tồn tại (not an error)
    log.Info("Key not found")
} else if err != nil {
    // Error thực sự
    log.Error("Redis error", zap.Error(err))
} else {
    // Success
    log.Info("Value:", value)
}
```

## Best Practices

### 1. Luôn dùng Context với Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

value, err := rdb.Get(ctx, "key")
```

### 2. Set expiration cho mọi keys (tránh memory leak)

```go
// ✅ ĐÚNG - có expiration
rdb.Set(ctx, "session:abc", data, 24*time.Hour)

// ❌ SAI - no expiration (có thể gây memory leak)
rdb.Set(ctx, "session:abc", data, 0)
```

### 3. Sử dụng key naming convention

```go
// Good key naming patterns:
"user:profile:123"
"session:abc123xyz"
"cache:product:456"
"rate_limit:user:789"
"lock:payment:order:123"
```

### 4. Handle redis.Nil error properly

```go
value, err := rdb.Get(ctx, "key")
if err == redis.Nil {
    // Đây KHÔNG phải error - key simply không tồn tại
    return defaultValue
} else if err != nil {
    // Đây là error thực sự
    return err
}
```

### 5. Sử dụng Pipeline cho batch operations

```go
// ✅ ĐÚNG - dùng pipeline
pipe := rdb.Client.Pipeline()
for i := 0; i < 100; i++ {
    pipe.Set(ctx, fmt.Sprintf("key:%d", i), i, 0)
}
pipe.Exec(ctx)

// ❌ SAI - multiple round trips
for i := 0; i < 100; i++ {
    rdb.Set(ctx, fmt.Sprintf("key:%d", i), i, 0)
}
```

### 6. Monitor pool health

```go
// Định kỳ check pool stats
ticker := time.NewTicker(1 * time.Minute)
go func() {
    for range ticker.C {
        stats := rdb.Stats()
        if stats.Timeouts > 100 {
            log.Warn("Redis pool timeouts detected!", "count", stats.Timeouts)
        }
    }
}()
```

## Graceful Shutdown

```go
func main() {
    rdb, _ := database.NewRedisClient(ctx, &cfg.Redis, logger)
    defer rdb.Close() // Đóng client khi shutdown

    // Setup signal handling
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    <-quit
    log.Info("Shutting down...")
    // rdb.Close() sẽ được gọi bởi defer
}
```

## Testing

### FlushDB (CẢNH BÁO: chỉ dùng cho testing!)

```go
// ⚠️ Xóa tất cả keys trong DB hiện tại
err := rdb.FlushDB(ctx)
```

### Mock Redis trong tests

```go
import "github.com/go-redis/redismock/v9"

func TestWithMockRedis(t *testing.T) {
    client, mock := redismock.NewClientMock()

    mock.ExpectSet("key", "value", 0).SetVal("OK")
    mock.ExpectGet("key").SetVal("value")

    // Test your code...
}
```

## Troubleshooting

### Connection Timeout

**Nguyên nhân:** Redis server không phản hồi trong thời gian quy định.

**Giải pháp:**
1. Check Redis server có đang chạy: `docker ps | grep redis`
2. Check network connectivity: `ping redis_host`
3. Tăng DialTimeout trong config
4. Check firewall/security groups

### Pool Exhausted

**Nguyên nhân:** Tất cả connections trong pool đang được sử dụng.

**Giải pháp:**
1. Tăng `REDIS_POOL_SIZE` trong `.env`
2. Check connection leaks (connections không được release)
3. Monitor pool stats để identify bottleneck

### High Latency

**Nguyên nhân:** Redis operations chậm.

**Giải pháo:**
1. Check Redis server load: `redis-cli INFO stats`
2. Use pipeline cho batch operations
3. Check network latency
4. Consider Redis cluster/sharding

## Security Notes

### 1. Luôn dùng password trong production

```env
REDIS_PASSWORD=your_strong_password_here
```

### 2. Sử dụng separate databases cho different purposes

```go
// DB 0: Cache
// DB 1: Sessions
// DB 2: Rate limiting
// etc.
```

### 3. Không lưu sensitive data trực tiếp

```go
// ✅ ĐÚNG - encrypt trước khi lưu
encryptedData := encrypt(sensitiveData)
rdb.Set(ctx, "user:ssn:123", encryptedData, ttl)

// ❌ SAI - lưu plaintext
rdb.Set(ctx, "user:ssn:123", "123-45-6789", ttl)
```

## Examples

Xem thêm examples chi tiết trong file `redis_example_usage.go`.
