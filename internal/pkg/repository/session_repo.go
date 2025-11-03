package repository

import (
	"context"
	"errors"
	"fmt"
	"system/internal/domain"
	"system/internal/platform/database"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	sessionKeyPrefix = "session:%s"
)

// sessionRepository triển khai SessionRepository sử dụng Redis
type sessionRepository struct {
	client *database.RedisClient
}

// NewSessionRepository tạo một instance mới của sessionRepository
func NewSessionRepository(client *database.RedisClient) domain.SessionRepository {
	return &sessionRepository{client: client}
}

// CreateSession tạo một session mới trong Redis
func (r *sessionRepository) CreateSession(ctx context.Context, sessionID string, userID string, ttl time.Duration) error {
	key := fmt.Sprintf(sessionKeyPrefix, sessionID)
	return r.client.Set(ctx, key, userID, ttl)
}

// GetSession lấy userID từ session
func (r *sessionRepository) GetSession(ctx context.Context, sessionID string) (string, error) {
	key := fmt.Sprintf(sessionKeyPrefix, sessionID)
	userID, err := r.client.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", fmt.Errorf("session not found or expired")
		}
		return "", err
	}
	return userID, nil
}

// DeleteSession xóa session khỏi Redis
func (r *sessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf(sessionKeyPrefix, sessionID)
	return r.client.Del(ctx, key)
}

// RefreshSession gia hạn session
func (r *sessionRepository) RefreshSession(ctx context.Context, sessionID string, ttl time.Duration) error {
	key := fmt.Sprintf(sessionKeyPrefix, sessionID)

	// Check if session exists
	userID, err := r.client.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Reset TTL by setting again with new expiration
	return r.client.Set(ctx, key, userID, ttl)
}
