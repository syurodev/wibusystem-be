package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"system/internal/domain"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/ory/fosite"
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
		return nil, fosite.ErrNotFound.WithWrap(err).WithDebug("Invalid UUID format for client ID")
	}

	domainClient, err := s.clientRepo.GetClientByID(ctx, clientID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fosite.ErrNotFound.WithWrap(err)
		}
		return nil, fosite.ErrServerError.WithWrap(err)
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
		return fosite.ErrServerError.WithWrap(err)
	}

	return s.sessionRepo.CreateSession(
		ctx,
		signature,
		requester.GetID(),
		"refresh_token",
		data,
		requester.GetSession().GetExpiresAt(fosite.RefreshToken),
		requester.GetClient().GetID(),
		requester.GetSession().GetSubject(),
	)
}

func (s *SQLStore) GetRefreshTokenSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	data, err := s.sessionRepo.GetSessionBySignature(ctx, signature)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fosite.ErrNotFound.WithWrap(err)
		}
		return nil, fosite.ErrServerError.WithWrap(err)
	}

	requester := &fosite.Request{Session: session}
	if err := json.Unmarshal(data, requester); err != nil {
		return nil, fosite.ErrServerError.WithWrap(err)
	}
	return requester, nil
}

func (s *SQLStore) DeleteRefreshTokenSession(ctx context.Context, signature string) error {
	return s.sessionRepo.DeleteSessionBySignature(ctx, signature)
}

func (s *SQLStore) RotateRefreshToken(ctx context.Context, requestID string, refreshTokenSignature string) error {
	// For now, we can just delete the old one. A more sophisticated implementation could be added later.
	return s.sessionRepo.DeleteSessionBySignature(ctx, refreshTokenSignature)
}
