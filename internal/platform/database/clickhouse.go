package database

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"go.uber.org/zap"

	"system/configs"
)

// ClickHouseClient là wrapper cho ClickHouse driver.Conn với các thông tin bổ sung.
// Struct này cung cấp interface để làm việc với ClickHouse.
type ClickHouseClient struct {
	Conn   driver.Conn // ClickHouse connection chính
	logger *zap.Logger
	config *configs.ClickHouseConfig
}

// NewClickHouseClient khởi tạo kết nối ClickHouse với clickhouse-go/v2.
// Hàm này sẽ:
//   - Tạo ClickHouse connection với các tham số từ config
//   - Ping để kiểm tra kết nối
//   - Trả về ClickHouseClient instance hoặc error
//
// Parameters:
//   - ctx: Context cho việc khởi tạo (thường dùng context.Background())
//   - cfg: ClickHouse configuration từ configs package
//   - logger: Zap logger instance để log operations và errors
//
// Returns:
//   - *ClickHouseClient: ClickHouse client đã kết nối thành công
//   - error: Lỗi nếu không thể kết nối
func NewClickHouseClient(ctx context.Context, cfg *configs.ClickHouseConfig, logger *zap.Logger) (*ClickHouseClient, error) {
	// 1. Tạo ClickHouse connection options
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.User,
			Password: cfg.Password,
		},

		// Connection Pool Settings
		MaxOpenConns:    cfg.MaxOpenConns,
		MaxIdleConns:    cfg.MaxIdleConns,
		ConnMaxLifetime: cfg.ConnMaxLifetime,

		// Timeouts
		DialTimeout: 5 * time.Second,

		// Compression
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},

		// Settings
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},

		// Debug (set to false in production)
		Debug: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	// 2. Ping ClickHouse để verify kết nối
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := conn.Ping(pingCtx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	logger.Info("ClickHouse client connected successfully",
		zap.String("host", cfg.Host),
		zap.String("port", cfg.Port),
		zap.String("database", cfg.Database),
		zap.Int("max_open_conns", cfg.MaxOpenConns),
		zap.Int("max_idle_conns", cfg.MaxIdleConns),
	)

	// 3. Khởi tạo ClickHouseClient instance
	return &ClickHouseClient{
		Conn:   conn,
		logger: logger,
		config: cfg,
	}, nil
}

// Close đóng ClickHouse connection một cách graceful.
// Nên được gọi khi application shutdown để cleanup resources.
func (c *ClickHouseClient) Close() error {
	c.logger.Info("Closing ClickHouse client...")
	if err := c.Conn.Close(); err != nil {
		c.logger.Error("Failed to close ClickHouse client", zap.Error(err))
		return err
	}
	c.logger.Info("ClickHouse client closed successfully")
	return nil
}

// Ping kiểm tra ClickHouse connection có còn sống không.
// Useful cho health checks.
//
// Parameters:
//   - ctx: Context với timeout để tránh hang
//
// Returns:
//   - error: nil nếu ping thành công, error nếu thất bại
func (c *ClickHouseClient) Ping(ctx context.Context) error {
	return c.Conn.Ping(ctx)
}

// Health kiểm tra sức khỏe của ClickHouse connection.
// Trả về thông tin chi tiết về trạng thái của connection và server.
//
// Parameters:
//   - ctx: Context với timeout
//
// Returns:
//   - map[string]any: Thông tin health check
//   - error: Lỗi nếu health check failed
func (c *ClickHouseClient) Health(ctx context.Context) (map[string]any, error) {
	// 1. Ping ClickHouse
	start := time.Now()
	if err := c.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ClickHouse ping failed: %w", err)
	}
	latency := time.Since(start)

	// 2. Lấy server version (optional)
	var version string
	row := c.Conn.QueryRow(ctx, "SELECT version()")
	if err := row.Scan(&version); err != nil {
		c.logger.Warn("Failed to get ClickHouse version", zap.Error(err))
	}

	// 3. Trả về health info
	healthInfo := map[string]any{
		"status":     "healthy",
		"latency":    latency.String(),
		"latency_ms": latency.Milliseconds(),

		// Config
		"config": map[string]any{
			"database":       c.config.Database,
			"max_open_conns": c.config.MaxOpenConns,
			"max_idle_conns": c.config.MaxIdleConns,
		},
	}

	// Thêm version nếu có
	if version != "" {
		healthInfo["version"] = version
	}

	return healthInfo, nil
}
