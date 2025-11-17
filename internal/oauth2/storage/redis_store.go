package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"system/internal/domain"
	"system/internal/platform/database"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
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
	client     *database.RedisClient
	clientRepo domain.OAuth2ClientRepository
}

// NewRedisStore tạo một instance mới của RedisStore.
func NewRedisStore(client *database.RedisClient, clientRepo domain.OAuth2ClientRepository) *RedisStore {
	return &RedisStore{client: client, clientRepo: clientRepo}
}

// storedRequest là dữ liệu đã được strip client để (de)serialize an toàn.
type storedRequest struct {
	Request  *fosite.Request        `json:"request"`
	Session  *openid.DefaultSession `json:"session"`
	ClientID string                 `json:"client_id"`
}

// --- oauth2.AuthorizeCodeStorage --- //

func (s *RedisStore) CreateAuthorizeCodeSession(ctx context.Context, signature string, requester fosite.Requester) error {
	req, ok := requester.(*fosite.Request)
	if !ok {
		return fosite.ErrServerError.WithDebug("unexpected requester type for auth code storage")
	}

	session, ok := req.Session.(*openid.DefaultSession)
	if !ok || session == nil {
		return fosite.ErrServerError.WithDebug("unexpected session type for auth code storage")
	}

	payload := storedRequest{
		Request:  stripClient(req),
		Session:  session,
		ClientID: req.GetClient().GetID(),
	}

	key := fmt.Sprintf(authCodeKeyPrefix, signature)
	lifespan := requester.GetSession().GetExpiresAt(fosite.AuthorizeCode).Sub(time.Now().UTC())
	data, err := json.Marshal(payload)
	if err != nil {
		zap.L().Error("redis store: marshal auth code failed", zap.String("signature", signature), zap.Error(err))
		return fosite.ErrServerError.WithWrap(err)
	}
	if err := s.client.Set(ctx, key, data, lifespan); err != nil {
		zap.L().Error("redis store: set auth code failed", zap.String("signature", signature), zap.Error(err))
		return fosite.ErrServerError.WithWrap(err)
	}
	return nil
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
	var stored storedRequest
	if err := json.Unmarshal([]byte(data), &stored); err != nil {
		return nil, fosite.ErrServerError.WithWrap(err)
	}

	req, err := s.hydrateRequestWithClient(ctx, stored)
	if err != nil {
		zap.L().Error("redis store: hydrate auth code failed", zap.String("signature", signature), zap.Error(err))
		return nil, err
	}
	return req, nil
}

func (s *RedisStore) InvalidateAuthorizeCodeSession(ctx context.Context, signature string) error {
	if err := s.client.Del(ctx, fmt.Sprintf(authCodeKeyPrefix, signature)); err != nil {
		zap.L().Error("redis store: invalidate auth code failed", zap.String("signature", signature), zap.Error(err))
		return fosite.ErrServerError.WithWrap(err)
	}
	return nil
}

// --- openid.OpenIDConnectRequestStorage --- //

func (s *RedisStore) CreateOpenIDConnectSession(ctx context.Context, authorizeCode string, requester fosite.Requester) error {
	req, ok := requester.(*fosite.Request)
	if !ok {
		return fosite.ErrServerError.WithDebug("unexpected requester type for oidc session")
	}

	session, ok := req.Session.(*openid.DefaultSession)
	if !ok || session == nil {
		return fosite.ErrServerError.WithDebug("unexpected session type for oidc session")
	}

	payload := storedRequest{
		Request:  stripClient(req),
		Session:  session,
		ClientID: req.GetClient().GetID(),
	}

	key := fmt.Sprintf(oidcSessionKeyPrefix, authorizeCode)
	lifespan := requester.GetSession().GetExpiresAt(fosite.AuthorizeCode).Sub(time.Now().UTC())
	data, err := json.Marshal(payload)
	if err != nil {
		zap.L().Error("redis store: marshal oidc session failed", zap.String("code", authorizeCode), zap.Error(err))
		return fosite.ErrServerError.WithWrap(err)
	}
	if err := s.client.Set(ctx, key, data, lifespan); err != nil {
		zap.L().Error("redis store: set oidc session failed", zap.String("code", authorizeCode), zap.Error(err))
		return fosite.ErrServerError.WithWrap(err)
	}
	return nil
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

	var stored storedRequest
	if err := json.Unmarshal([]byte(data), &stored); err != nil {
		return nil, fosite.ErrServerError.WithWrap(err)
	}

	req, err := s.hydrateRequestWithClient(ctx, stored)
	if err != nil {
		zap.L().Error("redis store: hydrate oidc session failed", zap.String("code", authorizeCode), zap.Error(err))
		return nil, err
	}
	return req, nil
}

func (s *RedisStore) DeleteOpenIDConnectSession(ctx context.Context, authorizeCode string) error {
	return s.client.Del(ctx, fmt.Sprintf(oidcSessionKeyPrefix, authorizeCode))
}

// --- oauth2.PKCERequestStorage --- //

func (s *RedisStore) CreatePKCERequestSession(ctx context.Context, signature string, requester fosite.Requester) error {
	req, ok := requester.(*fosite.Request)
	if !ok {
		return fosite.ErrServerError.WithDebug("unexpected requester type for pkce storage")
	}

	session, ok := req.Session.(*openid.DefaultSession)
	if !ok || session == nil {
		return fosite.ErrServerError.WithDebug("unexpected session type for pkce storage")
	}

	payload := storedRequest{
		Request:  stripClient(req),
		Session:  session,
		ClientID: req.GetClient().GetID(),
	}

	key := fmt.Sprintf(pkceKeyPrefix, signature)
	lifespan := requester.GetSession().GetExpiresAt(fosite.AuthorizeCode).Sub(time.Now().UTC())
	data, err := json.Marshal(payload)
	if err != nil {
		zap.L().Error("redis store: marshal pkce session failed", zap.String("signature", signature), zap.Error(err))
		return fosite.ErrServerError.WithWrap(err)
	}
	if err := s.client.Set(ctx, key, data, lifespan); err != nil {
		zap.L().Error("redis store: set pkce session failed", zap.String("signature", signature), zap.Error(err))
		return fosite.ErrServerError.WithWrap(err)
	}
	return nil
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
	var stored storedRequest
	if err := json.Unmarshal([]byte(data), &stored); err != nil {
		return nil, fosite.ErrServerError.WithWrap(err)
	}

	req, err := s.hydrateRequestWithClient(ctx, stored)
	if err != nil {
		zap.L().Error("redis store: hydrate pkce session failed", zap.String("signature", signature), zap.Error(err))
		return nil, err
	}
	return req, nil
}

func (s *RedisStore) DeletePKCERequestSession(ctx context.Context, signature string) error {
	return s.client.Del(ctx, fmt.Sprintf(pkceKeyPrefix, signature))
}

// --- oauth2.AccessTokenStorage --- //

func (s *RedisStore) CreateAccessTokenSession(ctx context.Context, signature string, requester fosite.Requester) error {
	req, ok := requester.(*fosite.Request)
	if !ok {
		return fosite.ErrServerError.WithDebug("unexpected requester type for access token storage")
	}

	session, ok := req.Session.(*openid.DefaultSession)
	if !ok || session == nil {
		return fosite.ErrServerError.WithDebug("unexpected session type for access token storage")
	}

	payload := storedRequest{
		Request:  stripClient(req),
		Session:  session,
		ClientID: req.GetClient().GetID(),
	}

	key := fmt.Sprintf(accessTokenKeyPrefix, signature)
	lifespan := requester.GetSession().GetExpiresAt(fosite.AccessToken).Sub(time.Now().UTC())
	data, err := json.Marshal(payload)
	if err != nil {
		zap.L().Error("redis store: marshal access token failed", zap.String("signature", signature), zap.Error(err))
		return fosite.ErrServerError.WithWrap(err)
	}
	if err := s.client.Set(ctx, key, data, lifespan); err != nil {
		zap.L().Error("redis store: set access token failed", zap.String("signature", signature), zap.Error(err))
		return fosite.ErrServerError.WithWrap(err)
	}
	return nil
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
	var stored storedRequest
	if err := json.Unmarshal([]byte(data), &stored); err != nil {
		return nil, fosite.ErrServerError.WithWrap(err)
	}

	req, err := s.hydrateRequestWithClient(ctx, stored)
	if err != nil {
		zap.L().Error("redis store: hydrate access token failed", zap.String("signature", signature), zap.Error(err))
		return nil, err
	}
	return req, nil
}

func (s *RedisStore) DeleteAccessTokenSession(ctx context.Context, signature string) error {
	return s.client.Del(ctx, fmt.Sprintf(accessTokenKeyPrefix, signature))
}

// --- oauth2.RefreshTokenStorage --- //

func (s *RedisStore) CreateRefreshTokenSession(ctx context.Context, signature string, accessSignature string, requester fosite.Requester) error {
	req, ok := requester.(*fosite.Request)
	if !ok {
		return fosite.ErrServerError.WithDebug("unexpected requester type for refresh token storage")
	}

	session, ok := req.Session.(*openid.DefaultSession)
	if !ok || session == nil {
		return fosite.ErrServerError.WithDebug("unexpected session type for refresh token storage")
	}

	payload := storedRequest{
		Request:  stripClient(req),
		Session:  session,
		ClientID: req.GetClient().GetID(),
	}

	key := fmt.Sprintf(refreshTokenKeyPrefix, signature)
	lifespan := requester.GetSession().GetExpiresAt(fosite.RefreshToken).Sub(time.Now().UTC())
	data, err := json.Marshal(payload)
	if err != nil {
		zap.L().Error("redis store: marshal refresh token failed", zap.String("signature", signature), zap.Error(err))
		return fosite.ErrServerError.WithWrap(err)
	}
	if err := s.client.Set(ctx, key, data, lifespan); err != nil {
		zap.L().Error("redis store: set refresh token failed", zap.String("signature", signature), zap.Error(err))
		return fosite.ErrServerError.WithWrap(err)
	}
	return nil
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
	var stored storedRequest
	if err := json.Unmarshal([]byte(data), &stored); err != nil {
		return nil, fosite.ErrServerError.WithWrap(err)
	}

	req, err := s.hydrateRequestWithClient(ctx, stored)
	if err != nil {
		zap.L().Error("redis store: hydrate refresh token failed", zap.String("signature", signature), zap.Error(err))
		return nil, err
	}
	return req, nil
}

func (s *RedisStore) DeleteRefreshTokenSession(ctx context.Context, signature string) error {
	return s.client.Del(ctx, fmt.Sprintf(refreshTokenKeyPrefix, signature))
}

func (s *RedisStore) RotateRefreshToken(ctx context.Context, requestID string, refreshTokenSignature string) error {
	return nil
}

// stripClient sao chép request và bỏ client để tránh lỗi serialize interface.
func stripClient(req *fosite.Request) *fosite.Request {
	sanitized := *req
	sanitized.Client = nil
	sanitized.Session = nil
	return &sanitized
}

// hydrateRequestWithClient nạp lại client từ DB và gán vào request đã deserialize.
func (s *RedisStore) hydrateRequestWithClient(ctx context.Context, stored storedRequest) (*fosite.Request, error) {
	clientUUID, err := uuid.FromString(stored.ClientID)
	if err != nil {
		return nil, fosite.ErrInvalidRequest.WithDebug("invalid client id in stored request")
	}

	domainClient, err := s.clientRepo.GetClientByID(ctx, clientUUID)
	if err != nil {
		if errors.Is(err, fosite.ErrNotFound) {
			return nil, err
		}
		return nil, fosite.ErrServerError.WithWrap(err)
	}

	fositeClient := &fosite.DefaultClient{
		ID:            domainClient.ID.String(),
		Secret:        []byte(domainClient.SecretHash),
		RedirectURIs:  domainClient.RedirectURIs,
		GrantTypes:    domainClient.GrantTypes,
		ResponseTypes: domainClient.ResponseTypes,
		Scopes:        domainClient.Scopes,
		Public:        domainClient.IsPublic,
	}

	stored.Request.Client = fositeClient
	// Gắn lại session đã lưu
	stored.Request.Session = stored.Session
	return stored.Request, nil
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
