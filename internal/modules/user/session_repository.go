package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"system/internal/domain"
	"system/internal/platform/database"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	sessionKeyPrefix      = "session:%s"
	userSessionsKeyPrefix = "user_sessions:%s"
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
func (r *sessionRepository) CreateSession(ctx context.Context, session *domain.UserSession, ttl time.Duration) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	key := fmt.Sprintf(sessionKeyPrefix, session.SessionID)
	pipe := r.client.Client.Pipeline()

	// Store session data
	pipe.Set(ctx, key, data, ttl)

	// Add to user's session list
	userKey := fmt.Sprintf(userSessionsKeyPrefix, session.UserID)
	pipe.SAdd(ctx, userKey, session.SessionID)
	// Optionally set expiry on the set, but it's tricky as it holds multiple sessions.
	// We'll rely on lazy cleanup or a long expiry if needed.
	// For now, let's just keep it simple.

	_, err = pipe.Exec(ctx)
	return err
}

// GetSession lấy thông tin session
func (r *sessionRepository) GetSession(ctx context.Context, sessionID string) (*domain.UserSession, error) {
	key := fmt.Sprintf(sessionKeyPrefix, sessionID)
	data, err := r.client.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("session not found or expired")
		}
		return nil, err
	}

	var session domain.UserSession
	if err := json.Unmarshal([]byte(data), &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// GetUserSessions lấy danh sách tất cả session của user
func (r *sessionRepository) GetUserSessions(ctx context.Context, userID string) ([]*domain.UserSession, error) {
	userKey := fmt.Sprintf(userSessionsKeyPrefix, userID)
	sessionIDs, err := r.client.Client.SMembers(ctx, userKey).Result()
	if err != nil {
		return nil, err
	}

	if len(sessionIDs) == 0 {
		return []*domain.UserSession{}, nil
	}

	// Fetch all sessions
	var sessions []*domain.UserSession
	// Use pipeline for efficiency
	pipe := r.client.Client.Pipeline()
	cmds := make(map[string]*redis.StringCmd)

	for _, id := range sessionIDs {
		key := fmt.Sprintf(sessionKeyPrefix, id)
		cmds[id] = pipe.Get(ctx, key)
	}

	_, err = pipe.Exec(ctx)
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, err
	}

	for id, cmd := range cmds {
		data, err := cmd.Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				// Cleanup expired session from set
				r.client.Client.SRem(ctx, userKey, id)
				continue
			}
			continue // Log error?
		}

		var session domain.UserSession
		if err := json.Unmarshal([]byte(data), &session); err == nil {
			sessions = append(sessions, &session)
		}
	}

	return sessions, nil
}

// DeleteSession xóa session khỏi Redis
func (r *sessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	// First get the session to know the UserID (to remove from set)
	session, err := r.GetSession(ctx, sessionID)
	if err != nil {
		// If session is gone, we might still want to try deleting the key just in case
		key := fmt.Sprintf(sessionKeyPrefix, sessionID)
		return r.client.Del(ctx, key)
	}

	key := fmt.Sprintf(sessionKeyPrefix, sessionID)
	userKey := fmt.Sprintf(userSessionsKeyPrefix, session.UserID)

	pipe := r.client.Client.Pipeline()
	pipe.Del(ctx, key)
	pipe.SRem(ctx, userKey, sessionID)
	_, err = pipe.Exec(ctx)
	return err
}

// DeleteUserSessions xóa tất cả session của user
func (r *sessionRepository) DeleteUserSessions(ctx context.Context, userID string) error {
	userKey := fmt.Sprintf(userSessionsKeyPrefix, userID)
	sessionIDs, err := r.client.Client.SMembers(ctx, userKey).Result()
	if err != nil {
		return err
	}

	if len(sessionIDs) == 0 {
		return nil
	}

	pipe := r.client.Client.Pipeline()
	for _, id := range sessionIDs {
		key := fmt.Sprintf(sessionKeyPrefix, id)
		pipe.Del(ctx, key)
	}
	pipe.Del(ctx, userKey)

	_, err = pipe.Exec(ctx)
	return err
}

// RefreshSession gia hạn session
func (r *sessionRepository) RefreshSession(ctx context.Context, sessionID string, ttl time.Duration) error {
	key := fmt.Sprintf(sessionKeyPrefix, sessionID)

	// Check if session exists
	val, err := r.client.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("session not found: %w", err)
	}

	// Reset TTL
	return r.client.Set(ctx, key, val, ttl)
}
