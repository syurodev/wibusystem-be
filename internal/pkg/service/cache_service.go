package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"system/internal/platform/database"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// CacheService provides generic caching operations with Redis.
type CacheService struct {
	redis  *database.RedisClient
	logger *zap.Logger
}

// NewCacheService creates a new CacheService instance.
func NewCacheService(redis *database.RedisClient, logger *zap.Logger) *CacheService {
	return &CacheService{
		redis:  redis,
		logger: logger,
	}
}

// Get retrieves a value from cache and unmarshals it into the provided type.
// Returns (value, true) if found, (zero, false) if not found or error.
func Get[T any](ctx context.Context, cs *CacheService, key string) (T, bool) {
	var zero T

	data, err := cs.redis.Get(ctx, key)
	if err != nil {
		if err != redis.Nil {
			cs.logger.Warn("cache get error", zap.String("key", key), zap.Error(err))
		}
		return zero, false
	}

	var result T
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		cs.logger.Warn("cache unmarshal error", zap.String("key", key), zap.Error(err))
		return zero, false
	}

	return result, true
}

// Set stores a value in cache with the given TTL.
func Set[T any](ctx context.Context, cs *CacheService, key string, value T, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	if err := cs.redis.Set(ctx, key, data, ttl); err != nil {
		cs.logger.Warn("cache set error", zap.String("key", key), zap.Error(err))
		return err
	}

	cs.logger.Debug("cache set", zap.String("key", key), zap.Duration("ttl", ttl))
	return nil
}

// Delete removes a key from cache.
func (cs *CacheService) Delete(ctx context.Context, keys ...string) error {
	return cs.redis.Del(ctx, keys...)
}

// GetOrSet retrieves from cache, or calls fetchFunc to get value and caches it.
// This is the most commonly used pattern for caching.
func GetOrSet[T any](ctx context.Context, cs *CacheService, key string, ttl time.Duration, fetchFunc func() (T, error)) (T, error) {
	// Try to get from cache first
	if cached, found := Get[T](ctx, cs, key); found {
		cs.logger.Debug("cache hit", zap.String("key", key))
		return cached, nil
	}

	cs.logger.Debug("cache miss", zap.String("key", key))

	// Fetch fresh data
	value, err := fetchFunc()
	if err != nil {
		var zero T
		return zero, err
	}

	// Store in cache (non-blocking, don't fail if cache fails)
	if err := Set(ctx, cs, key, value, ttl); err != nil {
		cs.logger.Warn("failed to cache value", zap.String("key", key), zap.Error(err))
	}

	return value, nil
}
