package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"system/configs"
)

// PostgresDB là wrapper cho pgxpool.Pool với các thông tin bổ sung.
// Struct này cung cấp interface để làm việc với PostgreSQL database.
type PostgresDB struct {
	Pool *pgxpool.Pool // Connection pool chính

	// Schemas - Các schema được cấu hình cho từng domain
	IdentifySchema  string
	CatalogSchema   string
	CommunitySchema string
	PaymentSchema   string

	logger *zap.Logger
	config *configs.DatabaseConfig
}

// NewPostgresDB khởi tạo kết nối PostgreSQL với pgxpool.
// Hàm này sẽ:
//   - Tạo connection pool với các tham số từ config
//   - Thiết lập query logging (nếu được bật)
//   - Ping để kiểm tra kết nối
//   - Trả về PostgresDB instance hoặc error
//
// Parameters:
//   - ctx: Context cho việc khởi tạo (thường dùng context.Background())
//   - cfg: Database configuration từ configs package
//   - enableQueryLog: Bật/tắt query logging
//   - logger: Zap logger instance để log queries và errors
//
// Returns:
//   - *PostgresDB: Database instance đã kết nối thành công
//   - error: Lỗi nếu không thể kết nối
func NewPostgresDB(ctx context.Context, cfg *configs.DatabaseConfig, enableQueryLog bool, logger *zap.Logger) (*PostgresDB, error) {
	// 1. Xây dựng connection string (DSN - Data Source Name)
	// Format: postgres://username:password@host:port/database?sslmode=disable
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
	)

	// 2. Tạo connection pool config từ DSN
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// 3. Cấu hình connection pool parameters
	// MaxConns: Số lượng connection tối đa trong pool
	poolConfig.MaxConns = int32(cfg.MaxOpenConns)

	// MinConns: Số lượng connection tối thiểu được giữ trong pool (idle connections)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)

	// MaxConnLifetime: Thời gian tối đa một connection có thể tồn tại
	// Sau thời gian này, connection sẽ bị đóng và tạo mới (tránh connection cũ bị stale)
	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime

	// MaxConnIdleTime: Thời gian tối đa một connection có thể idle trước khi bị đóng
	// Default là 30 phút - giúp giải phóng resources khi không sử dụng
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	// HealthCheckPeriod: Interval để check health của idle connections
	// Default là 1 phút - đảm bảo connections trong pool vẫn khỏe mạnh
	poolConfig.HealthCheckPeriod = 1 * time.Minute

	// 4. Thiết lập query logging nếu được bật trong config
	if enableQueryLog {
		// Tracer sẽ log tất cả queries với level INFO
		poolConfig.ConnConfig.Tracer = &tracelog.TraceLog{
			Logger:   newPgxZapLogger(logger),
			LogLevel: tracelog.LogLevelInfo,
		}
	}

	// 5. Thiết lập default query execution mode
	// QueryExecModeSimpleProtocol: Sử dụng simple protocol thay vì extended protocol
	// Phù hợp cho các queries đơn giản, giảm overhead
	poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec

	// 6. Tạo connection pool với config đã setup
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// 7. Ping database để verify kết nối
	// Timeout 5 giây cho ping operation
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close() // Đóng pool nếu ping failed
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("PostgreSQL connection pool established successfully",
		zap.String("host", cfg.Host),
		zap.String("port", cfg.Port),
		zap.String("database", cfg.Name),
		zap.Int("max_conns", cfg.MaxOpenConns),
		zap.Int("min_conns", cfg.MaxIdleConns),
	)

	// 8. Khởi tạo PostgresDB instance
	return &PostgresDB{
		Pool:            pool,
		IdentifySchema:  cfg.IdentifySchema,
		CatalogSchema:   cfg.CatalogSchema,
		CommunitySchema: cfg.CommunitySchema,
		PaymentSchema:   cfg.PaymentSchema,
		logger:          logger,
		config:          cfg,
	}, nil
}

// Close đóng connection pool một cách graceful.
// Nên được gọi khi application shutdown để cleanup resources.
func (db *PostgresDB) Close() {
	db.logger.Info("Closing PostgreSQL connection pool...")
	db.Pool.Close()
	db.logger.Info("PostgreSQL connection pool closed successfully")
}

// Ping kiểm tra database connection có còn sống không.
// Useful cho health checks.
//
// Parameters:
//   - ctx: Context với timeout để tránh hang
//
// Returns:
//   - error: nil nếu ping thành công, error nếu thất bại
func (db *PostgresDB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

// Stats trả về thống kê của connection pool.
// Hữu ích cho monitoring và debugging.
//
// Returns:
//   - *pgxpool.Stat: Thống kê gồm số connections đang acquire, idle, total, etc.
func (db *PostgresDB) Stats() *pgxpool.Stat {
	return db.Pool.Stat()
}

// Health kiểm tra sức khỏe của database connection.
// Trả về thông tin chi tiết về trạng thái của pool.
//
// Parameters:
//   - ctx: Context với timeout
//
// Returns:
//   - map[string]interface{}: Thông tin health check
//   - error: Lỗi nếu health check failed
func (db *PostgresDB) Health(ctx context.Context) (map[string]interface{}, error) {
	// Ping database
	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	// Lấy pool stats
	stats := db.Stats()

	// Trả về health info
	healthInfo := map[string]interface{}{
		"status":              "healthy",
		"max_connections":     db.config.MaxOpenConns,
		"open_connections":    stats.TotalConns(),
		"in_use":              stats.AcquiredConns(),
		"idle":                stats.IdleConns(),
		"wait_count":          stats.EmptyAcquireCount(),
		"max_idle_closed":     stats.MaxIdleDestroyCount(),
		"max_lifetime_closed": stats.MaxLifetimeDestroyCount(),
	}

	return healthInfo, nil
}

// GetSchemaName helper để lấy schema name theo domain.
// Useful khi cần switch giữa các schemas trong queries.
//
// Parameters:
//   - domain: Tên domain ("identify", "catalog", "community", "payment")
//
// Returns:
//   - string: Schema name tương ứng
func (db *PostgresDB) GetSchemaName(domain string) string {
	switch domain {
	case "identify":
		return db.IdentifySchema
	case "catalog":
		return db.CatalogSchema
	case "community":
		return db.CommunitySchema
	case "payment":
		return db.PaymentSchema
	default:
		return "public"
	}
}

// WithTransaction executes a function within a database transaction.
// Tự động commit nếu function thành công, rollback nếu có error.
//
// Parameters:
//   - ctx: Context cho transaction
//   - fn: Function cần execute trong transaction, nhận pgx.Tx làm parameter
//
// Returns:
//   - error: nil nếu transaction thành công, error nếu failed hoặc rollback
//
// Example:
//
//	err := db.WithTransaction(ctx, func(tx pgx.Tx) error {
//	    _, err := tx.Exec(ctx, "INSERT INTO users ...")
//	    return err
//	})
func (db *PostgresDB) WithTransaction(ctx context.Context, fn func(pgx.Tx) error) error {
	// Begin transaction
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Defer rollback - sẽ được gọi nếu commit chưa được thực hiện
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			db.logger.Error("Failed to rollback transaction", zap.Error(err))
		}
	}()

	// Execute function
	if err := fn(tx); err != nil {
		return err // Rollback sẽ được gọi bởi defer
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// pgxZapLogger adapter để bridge giữa pgx tracer và zap logger.
type pgxZapLogger struct {
	logger *zap.Logger
}

// newPgxZapLogger tạo instance mới của pgxZapLogger.
func newPgxZapLogger(logger *zap.Logger) *pgxZapLogger {
	return &pgxZapLogger{logger: logger}
}

// Log implements tracelog.Logger interface.
// Được gọi bởi pgx tracer để log queries.
func (l *pgxZapLogger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]interface{}) {
	// Convert pgx log level to zap log level
	var zapLevel zapcore.Level
	switch level {
	case tracelog.LogLevelTrace, tracelog.LogLevelDebug:
		zapLevel = zap.DebugLevel
	case tracelog.LogLevelInfo:
		zapLevel = zap.InfoLevel
	case tracelog.LogLevelWarn:
		zapLevel = zap.WarnLevel
	case tracelog.LogLevelError:
		zapLevel = zap.ErrorLevel
	default:
		zapLevel = zap.InfoLevel
	}

	// Build zap fields from data map
	fields := make([]zap.Field, 0, len(data))
	for k, v := range data {
		fields = append(fields, zap.Any(k, v))
	}

	// Log with appropriate level
	if ce := l.logger.Check(zapLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}
