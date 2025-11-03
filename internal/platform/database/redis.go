package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"system/configs"
)

// RedisClient là wrapper cho redis.Client với các thông tin bổ sung.
// Struct này cung cấp interface để làm việc với Redis.
type RedisClient struct {
	Client *redis.Client // Redis client chính

	logger *zap.Logger
	config *configs.RedisConfig
}

// NewRedisClient khởi tạo kết nối Redis với go-redis/v9.
// Hàm này sẽ:
//   - Tạo Redis client với các tham số từ config
//   - Ping để kiểm tra kết nối
//   - Trả về RedisClient instance hoặc error
//
// Parameters:
//   - ctx: Context cho việc khởi tạo (thường dùng context.Background())
//   - cfg: Redis configuration từ configs package
//   - logger: Zap logger instance để log operations và errors
//
// Returns:
//   - *RedisClient: Redis client đã kết nối thành công
//   - error: Lỗi nếu không thể kết nối
func NewRedisClient(ctx context.Context, cfg *configs.RedisConfig, logger *zap.Logger) (*RedisClient, error) {
	// 1. Tạo Redis client options
	options := &redis.Options{
		// Network và Address
		Addr: fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),

		// Authentication
		Password: cfg.Password, // No password nếu empty string
		DB:       cfg.DB,       // Default DB là 0

		// Connection Pool Settings
		PoolSize:     cfg.PoolSize,     // Số lượng connections tối đa trong pool
		MinIdleConns: cfg.MinIdleConns, // Số lượng idle connections tối thiểu

		// Timeouts
		DialTimeout:  5 * time.Second, // Timeout khi tạo connection mới
		ReadTimeout:  3 * time.Second, // Timeout khi đọc data
		WriteTimeout: 3 * time.Second, // Timeout khi ghi data
		PoolTimeout:  4 * time.Second, // Timeout khi chờ connection từ pool

		// Connection Lifecycle
		ConnMaxIdleTime: 5 * time.Minute,  // Đóng idle connections sau 5 phút
		ConnMaxLifetime: 30 * time.Minute, // Đóng connections sau 30 phút (refresh)

		// Retry Strategy
		MaxRetries:      cfg.MaxRetries,         // Số lần retry tối đa khi command fail
		MinRetryBackoff: 8 * time.Millisecond,   // Backoff tối thiểu giữa các retry
		MaxRetryBackoff: 512 * time.Millisecond, // Backoff tối đa

		// Health Check
		// go-redis tự động health check connections trong pool
	}

	// 2. Tạo Redis client
	client := redis.NewClient(options)

	// 3. Ping Redis để verify kết nối
	// Timeout 5 giây cho ping operation
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		client.Close() // Đóng client nếu ping failed
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	logger.Info("Redis client connected successfully",
		zap.String("host", cfg.Host),
		zap.String("port", cfg.Port),
		zap.Int("db", cfg.DB),
		zap.Int("pool_size", cfg.PoolSize),
		zap.Int("min_idle_conns", cfg.MinIdleConns),
	)

	// 4. Khởi tạo RedisClient instance
	return &RedisClient{
		Client: client,
		logger: logger,
		config: cfg,
	}, nil
}

// Close đóng Redis client một cách graceful.
// Nên được gọi khi application shutdown để cleanup resources.
func (r *RedisClient) Close() error {
	r.logger.Info("Closing Redis client...")
	if err := r.Client.Close(); err != nil {
		r.logger.Error("Failed to close Redis client", zap.Error(err))
		return err
	}
	r.logger.Info("Redis client closed successfully")
	return nil
}

// Ping kiểm tra Redis connection có còn sống không.
// Useful cho health checks.
//
// Parameters:
//   - ctx: Context với timeout để tránh hang
//
// Returns:
//   - error: nil nếu ping thành công, error nếu thất bại
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}

// Stats trả về thống kê của connection pool.
// Hữu ích cho monitoring và debugging.
//
// Returns:
//   - *redis.PoolStats: Thống kê gồm số connections, hits, misses, timeouts, etc.
func (r *RedisClient) Stats() *redis.PoolStats {
	return r.Client.PoolStats()
}

// Health kiểm tra sức khỏe của Redis connection.
// Trả về thông tin chi tiết về trạng thái của pool và server.
//
// Parameters:
//   - ctx: Context với timeout
//
// Returns:
//   - map[string]any: Thông tin health check
//   - error: Lỗi nếu health check failed
func (r *RedisClient) Health(ctx context.Context) (map[string]any, error) {
	// 1. Ping Redis
	start := time.Now()
	if err := r.Ping(ctx); err != nil {
		return nil, fmt.Errorf("Redis ping failed: %w", err)
	}
	latency := time.Since(start)

	// 2. Lấy pool stats
	stats := r.Stats()

	// 3. Lấy Redis server info (optional - có thể bỏ qua nếu không cần)
	info, err := r.Client.Info(ctx, "server").Result()
	if err != nil {
		r.logger.Warn("Failed to get Redis server info", zap.Error(err))
	}

	// 4. Trả về health info
	healthInfo := map[string]any{
		"status":     "healthy",
		"latency":    latency.String(),
		"latency_ms": latency.Milliseconds(),

		// Pool Stats
		"pool": map[string]any{
			"total_conns": stats.TotalConns,
			"idle_conns":  stats.IdleConns,
			"stale_conns": stats.StaleConns,
			"hits":        stats.Hits,
			"misses":      stats.Misses,
			"timeouts":    stats.Timeouts,
		},

		// Config
		"config": map[string]any{
			"db":             r.config.DB,
			"pool_size":      r.config.PoolSize,
			"min_idle_conns": r.config.MinIdleConns,
		},
	}

	// Thêm server info nếu có
	if info != "" {
		healthInfo["server_info_available"] = true
	}

	return healthInfo, nil
}

// Set lưu key-value vào Redis với expiration time.
// Helper function để đơn giản hóa việc set data.
//
// Parameters:
//   - ctx: Context
//   - key: Redis key
//   - value: Giá trị cần lưu (sẽ được serialize)
//   - expiration: Thời gian expire (0 = không expire)
//
// Returns:
//   - error: nil nếu thành công
//
// Example:
//
//	err := redis.Set(ctx, "user:123", userData, 1*time.Hour)
func (r *RedisClient) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return r.Client.Set(ctx, key, value, expiration).Err()
}

// Get lấy giá trị từ Redis theo key.
// Helper function để đơn giản hóa việc get data.
//
// Parameters:
//   - ctx: Context
//   - key: Redis key
//
// Returns:
//   - string: Giá trị của key
//   - error: redis.Nil nếu key không tồn tại, error khác nếu có lỗi
//
// Example:
//
//	value, err := redis.Get(ctx, "user:123")
//	if err == redis.Nil {
//	    // Key không tồn tại
//	} else if err != nil {
//	    // Error khác
//	}
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

// Del xóa một hoặc nhiều keys từ Redis.
// Helper function để đơn giản hóa việc delete data.
//
// Parameters:
//   - ctx: Context
//   - keys: Danh sách keys cần xóa
//
// Returns:
//   - error: nil nếu thành công
//
// Example:
//
//	err := redis.Del(ctx, "user:123", "session:abc")
func (r *RedisClient) Del(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}

// Exists kiểm tra xem key có tồn tại không.
// Helper function để check key existence.
//
// Parameters:
//   - ctx: Context
//   - key: Redis key
//
// Returns:
//   - bool: true nếu key tồn tại, false nếu không
//   - error: nil nếu thành công
//
// Example:
//
//	exists, err := redis.Exists(ctx, "user:123")
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.Client.Exists(ctx, key).Result()
	return count > 0, err
}

// Expire set expiration time cho một key.
// Helper function để set TTL cho existing key.
//
// Parameters:
//   - ctx: Context
//   - key: Redis key
//   - expiration: Thời gian expire
//
// Returns:
//   - error: nil nếu thành công
//
// Example:
//
//	err := redis.Expire(ctx, "session:abc", 30*time.Minute)
func (r *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.Client.Expire(ctx, key, expiration).Err()
}

// TTL lấy thời gian còn lại trước khi key expire.
// Helper function để check TTL.
//
// Parameters:
//   - ctx: Context
//   - key: Redis key
//
// Returns:
//   - time.Duration: Thời gian còn lại (-2 nếu key không tồn tại, -1 nếu không có expiration)
//   - error: nil nếu thành công
//
// Example:
//
//	ttl, err := redis.TTL(ctx, "session:abc")
func (r *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.Client.TTL(ctx, key).Result()
}

// Incr tăng giá trị của key lên 1.
// Nếu key không tồn tại, sẽ được set thành 0 trước khi increment.
//
// Parameters:
//   - ctx: Context
//   - key: Redis key
//
// Returns:
//   - int64: Giá trị sau khi increment
//   - error: nil nếu thành công
//
// Example:
//
//	count, err := redis.Incr(ctx, "page_views")
func (r *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	return r.Client.Incr(ctx, key).Result()
}

// Decr giảm giá trị của key xuống 1.
// Nếu key không tồn tại, sẽ được set thành 0 trước khi decrement.
//
// Parameters:
//   - ctx: Context
//   - key: Redis key
//
// Returns:
//   - int64: Giá trị sau khi decrement
//   - error: nil nếu thành công
//
// Example:
//
//	count, err := redis.Decr(ctx, "remaining_stock")
func (r *RedisClient) Decr(ctx context.Context, key string) (int64, error) {
	return r.Client.Decr(ctx, key).Result()
}

// HSet set giá trị cho field trong hash.
// Helper function cho hash operations.
//
// Parameters:
//   - ctx: Context
//   - key: Redis hash key
//   - values: Field-value pairs (field1, value1, field2, value2, ...)
//
// Returns:
//   - error: nil nếu thành công
//
// Example:
//
//	err := redis.HSet(ctx, "user:123", "name", "John", "email", "john@example.com")
func (r *RedisClient) HSet(ctx context.Context, key string, values ...any) error {
	return r.Client.HSet(ctx, key, values...).Err()
}

// HGet lấy giá trị của field trong hash.
// Helper function cho hash operations.
//
// Parameters:
//   - ctx: Context
//   - key: Redis hash key
//   - field: Field name
//
// Returns:
//   - string: Giá trị của field
//   - error: redis.Nil nếu field không tồn tại
//
// Example:
//
//	name, err := redis.HGet(ctx, "user:123", "name")
func (r *RedisClient) HGet(ctx context.Context, key, field string) (string, error) {
	return r.Client.HGet(ctx, key, field).Result()
}

// HGetAll lấy tất cả fields và values trong hash.
// Helper function cho hash operations.
//
// Parameters:
//   - ctx: Context
//   - key: Redis hash key
//
// Returns:
//   - map[string]string: Map chứa tất cả field-value pairs
//   - error: nil nếu thành công
//
// Example:
//
//	userData, err := redis.HGetAll(ctx, "user:123")
func (r *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return r.Client.HGetAll(ctx, key).Result()
}

// FlushDB xóa tất cả keys trong database hiện tại.
// ⚠️ CẢNH BÁO: Chỉ sử dụng cho testing hoặc development!
//
// Parameters:
//   - ctx: Context
//
// Returns:
//   - error: nil nếu thành công
func (r *RedisClient) FlushDB(ctx context.Context) error {
	r.logger.Warn("FlushDB called - all keys in current DB will be deleted!")
	return r.Client.FlushDB(ctx).Err()
}
