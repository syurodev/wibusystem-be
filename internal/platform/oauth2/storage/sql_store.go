package storage

import (
	"context"
	"encoding/json"
	"errors"
	"system/internal/domain"
	pkgerrors "system/pkg/errors"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/ory/fosite"
	"go.uber.org/zap"
)

// SQLStore triển khai các interface lưu trữ bền vững của Fosite bằng cách sử dụng repository layer.
type SQLStore struct {
	clientRepo  domain.OAuth2ClientRepository
	sessionRepo domain.OAuth2SessionRepository
}

// NewSQLStore tạo một instance mới của SQLStore.
func NewSQLStore(clientRepo domain.OAuth2ClientRepository, sessionRepo domain.OAuth2SessionRepository) *SQLStore {
	return &SQLStore{
		clientRepo:  clientRepo,
		sessionRepo: sessionRepo,
	}
}

// --- fosite.ClientManager --- //

func (s *SQLStore) GetClient(ctx context.Context, id string) (fosite.Client, error) {
	clientID, err := uuid.FromString(id)
	if err != nil {
		zap.L().Warn("oauth2 sql store: invalid client id format", zap.String("client_id", id), zap.Error(err))
		return nil, fosite.ErrNotFound.WithWrap(err).WithDebug("Invalid UUID format for client ID")
	}

	domainClient, err := s.clientRepo.GetClientByID(ctx, clientID)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrResourceNotFound) {
			zap.L().Warn("oauth2 sql store: client not found", zap.String("client_id", id))
			return nil, fosite.ErrNotFound.WithWrap(err).WithDebug("Client not found in database")
		}
		zap.L().Error("oauth2 sql store: database error fetching client",
			zap.String("client_id", id),
			zap.Error(err),
		)
		return nil, fosite.ErrServerError.WithWrap(err).WithDebug("Database error: " + err.Error())
	}

	client := &fosite.DefaultClient{
		ID:            domainClient.ID.String(),
		Secret:        []byte(domainClient.SecretHash),
		RedirectURIs:  domainClient.RedirectURIs,
		GrantTypes:    domainClient.GrantTypes,
		ResponseTypes: domainClient.ResponseTypes,
		Scopes:        domainClient.Scopes,
		Public:        domainClient.IsPublic,
	}

	return client, nil
}

func (s *SQLStore) ClientAssertionJWTValid(ctx context.Context, jti string) error {
	return s.clientRepo.ClientAssertionJWTValid(ctx, jti)
}

func (s *SQLStore) SetClientAssertionJWT(ctx context.Context, jti string, exp time.Time) error {
	return s.clientRepo.SetClientAssertionJWT(ctx, jti, exp)
}

// --- oauth2.RefreshTokenStorage --- //

func (s *SQLStore) CreateRefreshTokenSession(ctx context.Context, signature string, accessSignature string, requester fosite.Requester) error {
	data, err := json.Marshal(requester)
	if err != nil {
		zap.L().Error("sql store: marshal refresh token session failed", zap.String("signature", signature), zap.Error(err))
		return fosite.ErrServerError.WithWrap(err)
	}

	// Debug: log the JSON data to check if it's valid
	zap.L().Debug("sql store: marshaled session data", zap.String("signature", signature), zap.String("data", string(data)))

	if err := s.sessionRepo.CreateSession(
		ctx,
		signature,
		requester.GetID(),
		"refresh_token",
		data,
		requester.GetSession().GetExpiresAt(fosite.RefreshToken),
		requester.GetClient().GetID(),
		requester.GetSession().GetSubject(),
	); err != nil {
		zap.L().Error("sql store: create refresh token session failed", zap.String("signature", signature), zap.Error(err))
		return fosite.ErrServerError.WithWrap(err)
	}
	return nil
}

func (s *SQLStore) GetRefreshTokenSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	// Lấy cả session_data và client_id từ DB
	data, clientID, err := s.sessionRepo.GetSessionWithClientBySignature(ctx, signature)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrResourceNotFound) {
			return nil, fosite.ErrNotFound.WithWrap(err)
		}
		return nil, fosite.ErrServerError.WithWrap(err)
	}

	requester := &fosite.Request{Session: session}
	if err := json.Unmarshal(data, requester); err != nil {
		return nil, fosite.ErrServerError.WithWrap(err)
	}

	// Hydrate lại client từ database
	clientUUID, err := uuid.FromString(clientID)
	if err != nil {
		zap.L().Error("sql store: invalid client id format in session",
			zap.String("signature", signature),
			zap.String("client_id", clientID),
			zap.Error(err),
		)
		return nil, fosite.ErrInvalidRequest.WithDebug("invalid client id in stored session")
	}

	domainClient, err := s.clientRepo.GetClientByID(ctx, clientUUID)
	if err != nil {
		if errors.Is(err, pkgerrors.ErrResourceNotFound) {
			zap.L().Warn("sql store: client not found for refresh token session",
				zap.String("signature", signature),
				zap.String("client_id", clientID),
			)
			return nil, fosite.ErrNotFound.WithWrap(err).WithDebug("Client not found")
		}
		zap.L().Error("sql store: failed to load client for refresh token session",
			zap.String("signature", signature),
			zap.String("client_id", clientID),
			zap.Error(err),
		)
		return nil, fosite.ErrServerError.WithWrap(err)
	}

	// Tạo fosite client từ domain client
	fositeClient := &fosite.DefaultClient{
		ID:            domainClient.ID.String(),
		Secret:        []byte(domainClient.SecretHash),
		RedirectURIs:  domainClient.RedirectURIs,
		GrantTypes:    domainClient.GrantTypes,
		ResponseTypes: domainClient.ResponseTypes,
		Scopes:        domainClient.Scopes,
		Public:        domainClient.IsPublic,
	}

	requester.Client = fositeClient
	return requester, nil
}

func (s *SQLStore) DeleteRefreshTokenSession(ctx context.Context, signature string) error {
	return s.sessionRepo.DeleteSessionBySignature(ctx, signature)
}

func (s *SQLStore) RotateRefreshToken(ctx context.Context, requestID string, refreshTokenSignature string) error {
	// For now, we can just delete the old one. A more sophisticated implementation could be added later.
	return s.sessionRepo.DeleteSessionBySignature(ctx, refreshTokenSignature)
}
