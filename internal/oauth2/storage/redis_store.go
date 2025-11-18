package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"system/internal/platform/database"
	"time"

	"github.com/ory/fosite"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	authCodeKeyPrefix     = "fosite:auth_code:%s"
	pkceKeyPrefix         = "fosite:pkce:%s"
	accessTokenKeyPrefix  = "fosite:access_token:%s"
	refreshTokenKeyPrefix = "fosite:refresh_token:%s"
	oidcSessionKeyPrefix  = "fosite:oidc:%s"
	revokedTokenKeyPrefix = "fosite:revoked:%s"
)

// RedisStore triển khai các interface lưu trữ tạm thời của Fosite bằng Redis.
type RedisStore struct {
	client        *database.RedisClient
	logger        *zap.Logger
	clientManager fosite.ClientManager // For loading clients when deserializing
}

// NewRedisStore tạo một instance mới của RedisStore.
func NewRedisStore(client *database.RedisClient, logger *zap.Logger) *RedisStore {
	return &RedisStore{
		client: client,
		logger: logger,
	}
}

// SetClientManager sets the client manager for loading clients during deserialization
func (s *RedisStore) SetClientManager(cm fosite.ClientManager) {
	s.clientManager = cm
}

// serializableRequest is a wrapper for fosite.Requester that can be JSON serialized
// It stores only the client_id instead of the full client object (which is an interface)
type serializableRequest struct {
	ID                string              `json:"id"`
	RequestedAt       time.Time           `json:"requestedAt"`
	ClientID          string              `json:"client_id"`
	RequestedScopes   []string            `json:"requestedScopes"`
	GrantedScopes     []string            `json:"grantedScopes"`
	Form              map[string][]string `json:"form"`
	Session           json.RawMessage     `json:"session"`
	RequestedAudience []string            `json:"requestedAudience"`
	GrantedAudience   []string            `json:"grantedAudience"`
}

// --- oauth2.AuthorizeCodeStorage --- //

func (s *RedisStore) CreateAuthorizeCodeSession(ctx context.Context, signature string, requester fosite.Requester) error {
	key := fmt.Sprintf(authCodeKeyPrefix, signature)
	lifespan := requester.GetSession().GetExpiresAt(fosite.AuthorizeCode).Sub(time.Now().UTC())

	// Serialize session separately
	sessionData, err := json.Marshal(requester.GetSession())
	if err != nil {
		s.logger.Error("Failed to marshal session",
			zap.String("error", err.Error()),
		)
		return fosite.ErrServerError.WithWrap(err)
	}

	// Create serializable wrapper with only client_id (not full client object)
	serializable := &serializableRequest{
		ID:                requester.GetID(),
		RequestedAt:       requester.GetRequestedAt(),
		ClientID:          requester.GetClient().GetID(), // Only store ID
		RequestedScopes:   requester.GetRequestedScopes(),
		GrantedScopes:     requester.GetGrantedScopes(),
		Form:              requester.GetRequestForm(),
		Session:           sessionData,
		RequestedAudience: requester.GetRequestedAudience(),
		GrantedAudience:   requester.GetGrantedAudience(),
	}

	data, err := json.Marshal(serializable)
	if err != nil {
		s.logger.Error("Failed to marshal authorization code session",
			zap.String("error", err.Error()),
		)
		return fosite.ErrServerError.WithWrap(err)
	}

	s.logger.Debug("Saving authorization code to Redis",
		zap.String("key", key),
		zap.String("client_id", serializable.ClientID),
		zap.Duration("lifespan", lifespan),
	)

	return s.client.Set(ctx, key, data, lifespan)
}

func (s *RedisStore) GetAuthorizeCodeSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	key := fmt.Sprintf(authCodeKeyPrefix, signature)

	s.logger.Debug("Fetching authorization code from Redis",
		zap.String("key", key),
		zap.String("signature", signature),
	)

	data, err := s.client.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			s.logger.Warn("Authorization code not found in Redis",
				zap.String("key", key),
				zap.String("signature", signature),
			)
			return nil, fosite.ErrNotFound.WithWrap(err)
		}
		s.logger.Error("Redis error fetching authorization code",
			zap.String("error", err.Error()),
			zap.String("key", key),
		)
		return nil, fosite.ErrServerError.WithWrap(err)
	}

	s.logger.Debug("Authorization code found in Redis",
		zap.String("key", key),
		zap.Int("data_length", len(data)),
	)

	// Unmarshal into serializable wrapper
	var serializable serializableRequest
	if err := json.Unmarshal([]byte(data), &serializable); err != nil {
		s.logger.Error("Failed to unmarshal authorization code session",
			zap.String("error", err.Error()),
			zap.String("key", key),
			zap.String("data_preview", data[:min(len(data), 200)]),
		)
		return nil, fosite.ErrServerError.WithWrap(err).WithDebug("JSON unmarshal failed: " + err.Error())
	}

	s.logger.Debug("Deserialized authorization code data",
		zap.String("client_id", serializable.ClientID),
		zap.String("request_id", serializable.ID),
	)

	// Load client from database using ClientManager
	if s.clientManager == nil {
		s.logger.Error("ClientManager not set in RedisStore")
		return nil, fosite.ErrServerError.WithDebug("ClientManager not configured")
	}

	client, err := s.clientManager.GetClient(ctx, serializable.ClientID)
	if err != nil {
		s.logger.Error("Failed to load client for authorization code",
			zap.String("error", err.Error()),
			zap.String("client_id", serializable.ClientID),
		)
		return nil, fosite.ErrServerError.WithWrap(err).WithDebug("Failed to load client: " + err.Error())
	}

	// Unmarshal session
	if err := json.Unmarshal(serializable.Session, session); err != nil {
		s.logger.Error("Failed to unmarshal session",
			zap.String("error", err.Error()),
		)
		return nil, fosite.ErrServerError.WithWrap(err).WithDebug("Session unmarshal failed: " + err.Error())
	}

	// Reconstruct fosite.Request
	requester := &fosite.Request{
		ID:                serializable.ID,
		RequestedAt:       serializable.RequestedAt,
		Client:            client, // Client loaded from database
		RequestedScope:    serializable.RequestedScopes,
		GrantedScope:      serializable.GrantedScopes,
		Form:              serializable.Form,
		Session:           session,
		RequestedAudience: serializable.RequestedAudience,
		GrantedAudience:   serializable.GrantedAudience,
	}

	s.logger.Debug("Authorization code session loaded successfully",
		zap.String("key", key),
		zap.String("client_id", client.GetID()),
	)

	return requester, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (s *RedisStore) InvalidateAuthorizeCodeSession(ctx context.Context, signature string) error {
	return s.client.Del(ctx, fmt.Sprintf(authCodeKeyPrefix, signature))
}

// --- openid.OpenIDConnectRequestStorage --- //

func (s *RedisStore) CreateOpenIDConnectSession(ctx context.Context, authorizeCode string, requester fosite.Requester) error {
	key := fmt.Sprintf(oidcSessionKeyPrefix, authorizeCode)
	lifespan := requester.GetSession().GetExpiresAt(fosite.AuthorizeCode).Sub(time.Now().UTC())
	data, err := json.Marshal(requester)
	if err != nil {
		return fosite.ErrServerError.WithWrap(err)
	}
	return s.client.Set(ctx, key, data, lifespan)
}

func (s *RedisStore) GetOpenIDConnectSession(ctx context.Context, authorizeCode string, requester fosite.Requester) (fosite.Requester, error) {
	key := fmt.Sprintf(oidcSessionKeyPrefix, authorizeCode)
	data, err := s.client.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fosite.ErrNotFound.WithWrap(err)
		}
		return nil, fosite.ErrServerError.WithWrap(err)
	}
	// Unmarshal vào chính requester được truyền vào, vì nó chứa session
	if err := json.Unmarshal([]byte(data), requester); err != nil {
		return nil, fosite.ErrServerError.WithWrap(err)
	}
	return requester, nil
}

func (s *RedisStore) DeleteOpenIDConnectSession(ctx context.Context, authorizeCode string) error {
	return s.client.Del(ctx, fmt.Sprintf(oidcSessionKeyPrefix, authorizeCode))
}

// --- oauth2.PKCERequestStorage --- //

func (s *RedisStore) CreatePKCERequestSession(ctx context.Context, signature string, requester fosite.Requester) error {
	key := fmt.Sprintf(pkceKeyPrefix, signature)
	lifespan := requester.GetSession().GetExpiresAt(fosite.AuthorizeCode).Sub(time.Now().UTC())

	// Serialize session separately
	sessionData, err := json.Marshal(requester.GetSession())
	if err != nil {
		s.logger.Error("Failed to marshal PKCE session",
			zap.String("error", err.Error()),
		)
		return fosite.ErrServerError.WithWrap(err)
	}

	// Create serializable wrapper with only client_id
	serializable := &serializableRequest{
		ID:                requester.GetID(),
		RequestedAt:       requester.GetRequestedAt(),
		ClientID:          requester.GetClient().GetID(),
		RequestedScopes:   requester.GetRequestedScopes(),
		GrantedScopes:     requester.GetGrantedScopes(),
		Form:              requester.GetRequestForm(),
		Session:           sessionData,
		RequestedAudience: requester.GetRequestedAudience(),
		GrantedAudience:   requester.GetGrantedAudience(),
	}

	data, err := json.Marshal(serializable)
	if err != nil {
		s.logger.Error("Failed to marshal PKCE request session",
			zap.String("error", err.Error()),
		)
		return fosite.ErrServerError.WithWrap(err)
	}

	s.logger.Debug("Saving PKCE request to Redis",
		zap.String("key", key),
		zap.String("client_id", serializable.ClientID),
		zap.Duration("lifespan", lifespan),
	)

	return s.client.Set(ctx, key, data, lifespan)
}

func (s *RedisStore) GetPKCERequestSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	key := fmt.Sprintf(pkceKeyPrefix, signature)

	s.logger.Debug("Fetching PKCE request from Redis",
		zap.String("key", key),
		zap.String("signature", signature),
	)

	data, err := s.client.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			s.logger.Warn("PKCE request not found in Redis",
				zap.String("key", key),
			)
			return nil, fosite.ErrNotFound.WithWrap(err)
		}
		s.logger.Error("Redis error fetching PKCE request",
			zap.String("error", err.Error()),
			zap.String("key", key),
		)
		return nil, fosite.ErrServerError.WithWrap(err)
	}

	s.logger.Debug("PKCE request found in Redis",
		zap.String("key", key),
		zap.Int("data_length", len(data)),
	)

	// Unmarshal into serializable wrapper
	var serializable serializableRequest
	if err := json.Unmarshal([]byte(data), &serializable); err != nil {
		s.logger.Error("Failed to unmarshal PKCE request session",
			zap.String("error", err.Error()),
			zap.String("key", key),
		)
		return nil, fosite.ErrServerError.WithWrap(err).WithDebug("JSON unmarshal failed: " + err.Error())
	}

	s.logger.Debug("Deserialized PKCE request data",
		zap.String("client_id", serializable.ClientID),
		zap.String("request_id", serializable.ID),
	)

	// Load client from database
	if s.clientManager == nil {
		s.logger.Error("ClientManager not set in RedisStore for PKCE")
		return nil, fosite.ErrServerError.WithDebug("ClientManager not configured")
	}

	client, err := s.clientManager.GetClient(ctx, serializable.ClientID)
	if err != nil {
		s.logger.Error("Failed to load client for PKCE request",
			zap.String("error", err.Error()),
			zap.String("client_id", serializable.ClientID),
		)
		return nil, fosite.ErrServerError.WithWrap(err).WithDebug("Failed to load client: " + err.Error())
	}

	// Unmarshal session
	if err := json.Unmarshal(serializable.Session, session); err != nil {
		s.logger.Error("Failed to unmarshal PKCE session",
			zap.String("error", err.Error()),
		)
		return nil, fosite.ErrServerError.WithWrap(err).WithDebug("Session unmarshal failed: " + err.Error())
	}

	// Reconstruct fosite.Request
	requester := &fosite.Request{
		ID:                serializable.ID,
		RequestedAt:       serializable.RequestedAt,
		Client:            client,
		RequestedScope:    serializable.RequestedScopes,
		GrantedScope:      serializable.GrantedScopes,
		Form:              serializable.Form,
		Session:           session,
		RequestedAudience: serializable.RequestedAudience,
		GrantedAudience:   serializable.GrantedAudience,
	}

	s.logger.Debug("PKCE request session loaded successfully",
		zap.String("key", key),
		zap.String("client_id", client.GetID()),
	)

	return requester, nil
}

func (s *RedisStore) DeletePKCERequestSession(ctx context.Context, signature string) error {
	return s.client.Del(ctx, fmt.Sprintf(pkceKeyPrefix, signature))
}

// --- oauth2.AccessTokenStorage --- //

func (s *RedisStore) CreateAccessTokenSession(ctx context.Context, signature string, requester fosite.Requester) error {
	key := fmt.Sprintf(accessTokenKeyPrefix, signature)
	lifespan := requester.GetSession().GetExpiresAt(fosite.AccessToken).Sub(time.Now().UTC())
	data, err := json.Marshal(requester)
	if err != nil {
		return fosite.ErrServerError.WithWrap(err)
	}
	return s.client.Set(ctx, key, data, lifespan)
}

func (s *RedisStore) GetAccessTokenSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	key := fmt.Sprintf(accessTokenKeyPrefix, signature)
	data, err := s.client.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fosite.ErrNotFound.WithWrap(err)
		}
		return nil, fosite.ErrServerError.WithWrap(err)
	}
	requester := &fosite.Request{Session: session}
	if err := json.Unmarshal([]byte(data), requester); err != nil {
		return nil, fosite.ErrServerError.WithWrap(err)
	}
	return requester, nil
}

func (s *RedisStore) DeleteAccessTokenSession(ctx context.Context, signature string) error {
	return s.client.Del(ctx, fmt.Sprintf(accessTokenKeyPrefix, signature))
}

// --- oauth2.RefreshTokenStorage --- //

func (s *RedisStore) CreateRefreshTokenSession(ctx context.Context, signature string, accessSignature string, requester fosite.Requester) error {
	key := fmt.Sprintf(refreshTokenKeyPrefix, signature)
	lifespan := requester.GetSession().GetExpiresAt(fosite.RefreshToken).Sub(time.Now().UTC())
	data, err := json.Marshal(requester)
	if err != nil {
		return fosite.ErrServerError.WithWrap(err)
	}
	return s.client.Set(ctx, key, data, lifespan)
}

func (s *RedisStore) GetRefreshTokenSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	key := fmt.Sprintf(refreshTokenKeyPrefix, signature)
	data, err := s.client.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fosite.ErrNotFound.WithWrap(err)
		}
		return nil, fosite.ErrServerError.WithWrap(err)
	}
	requester := &fosite.Request{Session: session}
	if err := json.Unmarshal([]byte(data), requester); err != nil {
		return nil, fosite.ErrServerError.WithWrap(err)
	}
	return requester, nil
}

func (s *RedisStore) DeleteRefreshTokenSession(ctx context.Context, signature string) error {
	return s.client.Del(ctx, fmt.Sprintf(refreshTokenKeyPrefix, signature))
}

func (s *RedisStore) RotateRefreshToken(ctx context.Context, requestID string, refreshTokenSignature string) error {
	return nil
}

// --- oauth2.TokenRevocationStorage --- //

func (s *RedisStore) RevokeAccessToken(ctx context.Context, requestID string) error {
	key := fmt.Sprintf(revokedTokenKeyPrefix, requestID)
	return s.client.Set(ctx, key, "revoked", time.Hour*24)
}

func (s *RedisStore) RevokeRefreshToken(ctx context.Context, requestID string) error {
	key := fmt.Sprintf(revokedTokenKeyPrefix, requestID)
	return s.client.Set(ctx, key, "revoked", time.Hour*24*35)
}
