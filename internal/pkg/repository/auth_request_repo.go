package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"system/internal/domain"
	"system/internal/platform/database"
	"time"

	"github.com/ory/fosite"
	"github.com/redis/go-redis/v9"
)

const (
	authRequestKeyPrefix = "auth_request:%s"
	authUserIDKeyPrefix  = "auth_request_user:%s"
)

// authRequestRepository triển khai AuthRequestRepository sử dụng Redis
type authRequestRepository struct {
	client *database.RedisClient
}

// NewAuthRequestRepository tạo một instance mới của authRequestRepository
func NewAuthRequestRepository(client *database.RedisClient) domain.AuthRequestRepository {
	return &authRequestRepository{client: client}
}

// SaveAuthRequest lưu authorization request vào Redis
func (r *authRequestRepository) SaveAuthRequest(ctx context.Context, requestID string, ar fosite.AuthorizeRequester, ttl time.Duration) error {
	key := fmt.Sprintf(authRequestKeyPrefix, requestID)

	// Serialize authorization request to JSON
	data, err := json.Marshal(ar)
	if err != nil {
		return fmt.Errorf("failed to marshal auth request: %w", err)
	}

	return r.client.Set(ctx, key, data, ttl)
}

// GetAuthRequest lấy authorization request từ Redis
func (r *authRequestRepository) GetAuthRequest(ctx context.Context, requestID string) (fosite.AuthorizeRequester, error) {
	key := fmt.Sprintf(authRequestKeyPrefix, requestID)

	data, err := r.client.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("auth request not found or expired")
		}
		return nil, err
	}

	// Deserialize from JSON
	var ar fosite.AuthorizeRequest
	if err := json.Unmarshal([]byte(data), &ar); err != nil {
		return nil, fmt.Errorf("failed to unmarshal auth request: %w", err)
	}

	return &ar, nil
}

// DeleteAuthRequest xóa authorization request khỏi Redis
func (r *authRequestRepository) DeleteAuthRequest(ctx context.Context, requestID string) error {
	key := fmt.Sprintf(authRequestKeyPrefix, requestID)
	userKey := fmt.Sprintf(authUserIDKeyPrefix, requestID)

	// Delete both keys
	if err := r.client.Del(ctx, key); err != nil {
		return err
	}
	return r.client.Del(ctx, userKey)
}

// SaveAuthRequestWithUserID lưu authorization request cùng với userID
func (r *authRequestRepository) SaveAuthRequestWithUserID(ctx context.Context, requestID string, ar fosite.AuthorizeRequester, userID string, ttl time.Duration) error {
	// Save authorization request
	if err := r.SaveAuthRequest(ctx, requestID, ar, ttl); err != nil {
		return err
	}

	// Save userID separately for quick lookup
	userKey := fmt.Sprintf(authUserIDKeyPrefix, requestID)
	return r.client.Set(ctx, userKey, userID, ttl)
}

// GetUserIDFromAuthRequest lấy userID từ stored authorization request
func (r *authRequestRepository) GetUserIDFromAuthRequest(ctx context.Context, requestID string) (string, error) {
	userKey := fmt.Sprintf(authUserIDKeyPrefix, requestID)

	userID, err := r.client.Get(ctx, userKey)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", fmt.Errorf("user id not found for auth request")
		}
		return "", err
	}

	return userID, nil
}
