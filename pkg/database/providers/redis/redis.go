package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"wibusystem/pkg/database/config"
	"wibusystem/pkg/database/interfaces"
)

// RedisProvider implements the CacheDatabase interface
type RedisProvider struct {
	client *redis.Client
	config *config.CacheConfig
}

// NewRedisProvider creates a new Redis cache provider
func NewRedisProvider(config *config.CacheConfig) (*RedisProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("redis config cannot be nil")
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid redis config: %w", err)
	}

	return &RedisProvider{
		config: config,
	}, nil
}

// Connect establishes connection to Redis
func (r *RedisProvider) Connect(ctx context.Context) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%d", r.config.Host, r.config.Port),
		Password:        r.config.Password,
		DB:              r.config.Database,
		MaxRetries:      r.config.MaxRetries,
		MinRetryBackoff: r.config.MinRetryBackoff,
		MaxRetryBackoff: r.config.MaxRetryBackoff,
		DialTimeout:     r.config.DialTimeout,
		ReadTimeout:     r.config.ReadTimeout,
		WriteTimeout:    r.config.WriteTimeout,
		PoolSize:        r.config.PoolSize,
		MinIdleConns:    r.config.MinIdleConns,
		ConnMaxLifetime: r.config.MaxConnAge,
		PoolTimeout:     r.config.PoolTimeout,
		ConnMaxIdleTime: r.config.IdleTimeout,
	})

	// Test the connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to ping Redis: %w", err)
	}

	r.client = rdb
	log.Printf("Successfully connected to Redis: %s:%d DB:%d", r.config.Host, r.config.Port, r.config.Database)
	return nil
}

// Close closes the Redis connection
func (r *RedisProvider) Close() error {
	if r.client != nil {
		if err := r.client.Close(); err != nil {
			return fmt.Errorf("failed to close Redis connection: %w", err)
		}
		log.Printf("Redis connection closed: %s:%d", r.config.Host, r.config.Port)
	}
	return nil
}

// Health checks if the Redis connection is healthy
func (r *RedisProvider) Health(ctx context.Context) error {
	if r.client == nil {
		return fmt.Errorf("redis provider not connected")
	}
	return r.client.Ping(ctx).Err()
}

// GetType returns the database type
func (r *RedisProvider) GetType() interfaces.DatabaseType {
	return interfaces.Redis
}

// BeginTx starts a new transaction (Redis doesn't support traditional transactions)
func (r *RedisProvider) BeginTx(ctx context.Context) (interfaces.Transaction, error) {
	if r.client == nil {
		return nil, fmt.Errorf("redis provider not connected")
	}

	// Redis uses MULTI/EXEC for transactions
	pipe := r.client.TxPipeline()
	return &RedisTransaction{pipeline: pipe}, nil
}

// Key-value operations

// Get retrieves a value by key
func (r *RedisProvider) Get(ctx context.Context, key string) (string, error) {
	if r.client == nil {
		return "", fmt.Errorf("redis provider not connected")
	}

	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key not found: %s", key)
		}
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}

	return result, nil
}

// Set sets a key-value pair with optional expiration
func (r *RedisProvider) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if r.client == nil {
		return fmt.Errorf("redis provider not connected")
	}

	err := r.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	return nil
}

// Delete removes one or more keys
func (r *RedisProvider) Delete(ctx context.Context, keys ...string) error {
	if r.client == nil {
		return fmt.Errorf("redis provider not connected")
	}

	err := r.client.Del(ctx, keys...).Err()
	if err != nil {
		return fmt.Errorf("failed to delete keys: %w", err)
	}

	return nil
}

// Exists checks if keys exist
func (r *RedisProvider) Exists(ctx context.Context, keys ...string) (int64, error) {
	if r.client == nil {
		return 0, fmt.Errorf("redis provider not connected")
	}

	result, err := r.client.Exists(ctx, keys...).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to check key existence: %w", err)
	}

	return result, nil
}

// Hash operations

// HGet gets a field value from a hash
func (r *RedisProvider) HGet(ctx context.Context, key, field string) (string, error) {
	if r.client == nil {
		return "", fmt.Errorf("redis provider not connected")
	}

	result, err := r.client.HGet(ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("hash field not found: %s.%s", key, field)
		}
		return "", fmt.Errorf("failed to get hash field %s.%s: %w", key, field, err)
	}

	return result, nil
}

// HSet sets field values in a hash
func (r *RedisProvider) HSet(ctx context.Context, key string, values ...interface{}) error {
	if r.client == nil {
		return fmt.Errorf("redis provider not connected")
	}

	err := r.client.HSet(ctx, key, values...).Err()
	if err != nil {
		return fmt.Errorf("failed to set hash fields for key %s: %w", key, err)
	}

	return nil
}

// HDelete deletes fields from a hash
func (r *RedisProvider) HDelete(ctx context.Context, key string, fields ...string) error {
	if r.client == nil {
		return fmt.Errorf("redis provider not connected")
	}

	err := r.client.HDel(ctx, key, fields...).Err()
	if err != nil {
		return fmt.Errorf("failed to delete hash fields for key %s: %w", key, err)
	}

	return nil
}

// List operations

// LPush pushes values to the left (head) of a list
func (r *RedisProvider) LPush(ctx context.Context, key string, values ...interface{}) error {
	if r.client == nil {
		return fmt.Errorf("redis provider not connected")
	}

	err := r.client.LPush(ctx, key, values...).Err()
	if err != nil {
		return fmt.Errorf("failed to push to list %s: %w", key, err)
	}

	return nil
}

// RPop pops a value from the right (tail) of a list
func (r *RedisProvider) RPop(ctx context.Context, key string) (string, error) {
	if r.client == nil {
		return "", fmt.Errorf("redis provider not connected")
	}

	result, err := r.client.RPop(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("list is empty: %s", key)
		}
		return "", fmt.Errorf("failed to pop from list %s: %w", key, err)
	}

	return result, nil
}

// Set operations

// SAdd adds members to a set
func (r *RedisProvider) SAdd(ctx context.Context, key string, members ...interface{}) error {
	if r.client == nil {
		return fmt.Errorf("redis provider not connected")
	}

	err := r.client.SAdd(ctx, key, members...).Err()
	if err != nil {
		return fmt.Errorf("failed to add to set %s: %w", key, err)
	}

	return nil
}

// SMembers returns all members of a set
func (r *RedisProvider) SMembers(ctx context.Context, key string) ([]string, error) {
	if r.client == nil {
		return nil, fmt.Errorf("redis provider not connected")
	}

	result, err := r.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get set members for %s: %w", key, err)
	}

	return result, nil
}

// GetClient returns the Redis client for advanced operations
func (r *RedisProvider) GetClient() *redis.Client {
	return r.client
}
