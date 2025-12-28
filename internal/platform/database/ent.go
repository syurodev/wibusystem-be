package database

import (
	"context"
	"database/sql"
	"fmt"

	"entgo.io/ent/dialect"
	esql "entgo.io/ent/dialect/sql"
	"go.uber.org/zap"

	// Register postgres driver for Ent
	_ "github.com/jackc/pgx/v5/stdlib"

	"system/configs"
	ent "system/internal/ent/generated"
)

// EntClient wrapper cho Ent client với các tiện ích bổ sung.
type EntClient struct {
	*ent.Client
	logger *zap.Logger
}

// NewEntClient tạo Ent client từ Config.
//
// Parameters:
//   - ctx: Context cho việc khởi tạo
//   - cfg: Database configuration
//   - debug: Bật/tắt debug logging
//   - logger: Zap logger instance
//
// Returns:
//   - *EntClient: Ent client instance
//   - error: Lỗi nếu không thể khởi tạo
func NewEntClient(ctx context.Context, cfg *configs.DatabaseConfig, debug bool, logger *zap.Logger) (*EntClient, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&search_path=%s,%s,%s,%s,public",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.SSLMode,
		cfg.IdentifySchema,
		cfg.CatalogSchema,
		cfg.CommunitySchema,
		cfg.PaymentSchema,
	)

	// Tạo sql.DB connection first for Ping
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Xác minh kết nối
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection passed ping check")

	// Tạo sql.Driver cho Ent từ *sql.DB
	drv := esql.OpenDB(dialect.Postgres, db)

	// Tạo Ent client options
	var opts []ent.Option
	opts = append(opts, ent.Driver(drv))

	if debug {
		opts = append(opts, ent.Debug())
		opts = append(opts, ent.Log(func(a ...any) {
			logger.Debug("ent", zap.Any("query", a))
		}))
	}

	// Force debug logs for migration troubleshooting
	opts = append(opts, ent.Debug())
	opts = append(opts, ent.Log(func(a ...any) {
		logger.Info("ENT QUERY", zap.Any("query", a))
	}))

	// Create Ent client
	client := ent.NewClient(opts...)

	// Auto-migration is handled externally or disabled by user request.
	// Use scripts/debug_migration.go or similar tools for schema management.

	logger.Info("Ent client initialized successfully")

	return &EntClient{
		Client: client,
		logger: logger,
	}, nil
}

// Close đóng Ent client.
func (c *EntClient) Close() error {
	c.logger.Info("Closing Ent client...")
	if err := c.Client.Close(); err != nil {
		return fmt.Errorf("failed to close ent client: %w", err)
	}
	c.logger.Info("Ent client closed successfully")
	return nil
}
