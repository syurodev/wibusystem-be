package database

// File này chứa các ví dụ sử dụng PostgresDB
// KHÔNG build file này - chỉ dùng để tham khảo

/*
import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// Example 1: Query đơn giản
func ExampleQueryRows(db *PostgresDB, logger *zap.Logger) error {
	ctx := context.Background()

	// Sử dụng schema từ config
	query := `SELECT id, email, name FROM identify.users WHERE status = $1 LIMIT 10`

	rows, err := db.Pool.Query(ctx, query, "active")
	if err != nil {
		return err
	}
	defer rows.Close()

	// Scan results
	for rows.Next() {
		var id int64
		var email, name string

		if err := rows.Scan(&id, &email, &name); err != nil {
			logger.Error("Failed to scan row", zap.Error(err))
			continue
		}

		logger.Info("User found",
			zap.Int64("id", id),
			zap.String("email", email),
			zap.String("name", name))
	}

	return rows.Err()
}

// Example 2: Query single row
func ExampleQueryRow(db *PostgresDB) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var email string
	query := `SELECT email FROM identify.users WHERE id = $1`

	err := db.Pool.QueryRow(ctx, query, 123).Scan(&email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil // Không tìm thấy user
		}
		return "", err
	}

	return email, nil
}

// Example 3: Insert/Update/Delete với Exec
func ExampleExec(db *PostgresDB) error {
	ctx := context.Background()

	query := `
		INSERT INTO catalog.products (name, price, category)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var productID int64
	err := db.Pool.QueryRow(ctx, query, "Product Name", 99.99, "Electronics").Scan(&productID)
	if err != nil {
		return err
	}

	// productID giờ chứa ID của record vừa insert
	return nil
}

// Example 4: Sử dụng Transaction
func ExampleTransaction(db *PostgresDB) error {
	ctx := context.Background()

	// WithTransaction tự động commit/rollback
	return db.WithTransaction(ctx, func(tx pgx.Tx) error {
		// Insert user
		var userID int64
		err := tx.QueryRow(ctx,
			`INSERT INTO identify.users (email, name) VALUES ($1, $2) RETURNING id`,
			"user@example.com", "John Doe",
		).Scan(&userID)
		if err != nil {
			return err // Tự động rollback
		}

		// Insert user profile
		_, err = tx.Exec(ctx,
			`INSERT INTO identify.user_profiles (user_id, bio) VALUES ($1, $2)`,
			userID, "Hello world",
		)
		if err != nil {
			return err // Tự động rollback
		}

		// Nếu không có error, transaction sẽ tự động commit
		return nil
	})
}

// Example 5: Batch Insert
func ExampleBatch(db *PostgresDB) error {
	ctx := context.Background()

	batch := &pgx.Batch{}

	// Thêm nhiều queries vào batch
	for i := 0; i < 100; i++ {
		batch.Queue(
			`INSERT INTO catalog.products (name, price) VALUES ($1, $2)`,
			fmt.Sprintf("Product %d", i),
			float64(i)*10,
		)
	}

	// Execute batch
	results := db.Pool.SendBatch(ctx, batch)
	defer results.Close()

	// Process results nếu cần
	for i := 0; i < 100; i++ {
		_, err := results.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

// Example 6: Query với dynamic schema
func ExampleDynamicSchema(db *PostgresDB, domain string) error {
	ctx := context.Background()

	// Lấy schema name theo domain
	schema := db.GetSchemaName(domain)

	// Build query với schema động
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s.users`, schema)

	var count int64
	err := db.Pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return err
	}

	return nil
}

// Example 7: Sử dụng connection pool stats cho monitoring
func ExampleMonitoring(db *PostgresDB, logger *zap.Logger) {
	stats := db.Stats()

	logger.Info("Database pool statistics",
		zap.Int32("total_conns", stats.TotalConns()),
		zap.Int32("acquired_conns", stats.AcquiredConns()),
		zap.Int32("idle_conns", stats.IdleConns()),
		zap.Int64("empty_acquire", stats.EmptyAcquireCount()),
		zap.Int64("max_idle_destroy", stats.MaxIdleDestroyCount()),
		zap.Int64("max_lifetime_destroy", stats.MaxLifetimeDestroyCount()),
	)
}

// Example 8: Health check endpoint (dùng trong HTTP handler)
func ExampleHealthCheck(db *PostgresDB) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return db.Health(ctx)
}

// Example 9: Prepared Statement (reusable query)
func ExamplePreparedStatement(db *PostgresDB) error {
	ctx := context.Background()

	// Acquire connection từ pool
	conn, err := db.Pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	// Tên prepared statement
	stmtName := "get_user_by_email"

	// Prepare statement (cache trên server)
	_, err = conn.Conn().Prepare(ctx, stmtName, `
		SELECT id, name, email FROM identify.users WHERE email = $1
	`)
	if err != nil {
		return err
	}

	// Sử dụng prepared statement nhiều lần
	emails := []string{"user1@test.com", "user2@test.com", "user3@test.com"}
	for _, email := range emails {
		var id int64
		var name, userEmail string

		err := conn.QueryRow(ctx, stmtName, email).Scan(&id, &name, &userEmail)
		if err != nil {
			if err == pgx.ErrNoRows {
				continue
			}
			return err
		}

		// Process user...
	}

	return nil
}
*/
