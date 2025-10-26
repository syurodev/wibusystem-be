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
	client *database.RedisClient
}

// NewRedisStore tạo một instance mới của RedisStore.
func NewRedisStore(client *database.RedisClient) *RedisStore {
	return &RedisStore{client: client}
}

// --- oauth2.AuthorizeCodeStorage --- //

func (s *RedisStore) CreateAuthorizeCodeSession(ctx context.Context, signature string, requester fosite.Requester) error {
	key := fmt.Sprintf(authCodeKeyPrefix, signature)
	lifespan := requester.GetSession().GetExpiresAt(fosite.AuthorizeCode).Sub(time.Now().UTC())
	data, err := json.Marshal(requester)
	if err != nil {
		return fosite.ErrServerError.WithWrap(err)
	}
	return s.client.Set(ctx, key, data, lifespan)
}

func (s *RedisStore) GetAuthorizeCodeSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	key := fmt.Sprintf(authCodeKeyPrefix, signature)
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
	data, err := json.Marshal(requester)
	if err != nil {
		return fosite.ErrServerError.WithWrap(err)
	}
	return s.client.Set(ctx, key, data, lifespan)
}

func (s *RedisStore) GetPKCERequestSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	key := fmt.Sprintf(pkceKeyPrefix, signature)
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
