package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// RedisTransaction implements the Transaction interface for Redis
type RedisTransaction struct {
	pipeline redis.Pipeliner
}

// Commit executes the Redis transaction (EXEC)
func (t *RedisTransaction) Commit(ctx context.Context) error {
	if t.pipeline == nil {
		return fmt.Errorf("redis pipeline is nil")
	}

	_, err := t.pipeline.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute redis transaction: %w", err)
	}

	return nil
}

// Rollback discards the Redis transaction (DISCARD)
func (t *RedisTransaction) Rollback(ctx context.Context) error {
	if t.pipeline == nil {
		return fmt.Errorf("redis pipeline is nil")
	}

	// Redis doesn't have explicit rollback, just discard the pipeline
	t.pipeline.Discard()
	return nil
}

// GetPipeline returns the underlying Redis pipeline for transaction operations
func (t *RedisTransaction) GetPipeline() redis.Pipeliner {
	return t.pipeline
}
