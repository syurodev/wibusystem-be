package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

// OAuth2Client là domain model cho một OAuth2 client.
type OAuth2Client struct {
	ID            uuid.UUID
	Name          string
	SecretHash    string
	RedirectURIs  []string
	GrantTypes    []string
	ResponseTypes []string
	Scopes        []string
	IsPublic      bool
	TenantID      *uuid.UUID // Dùng con trỏ cho giá trị nullable
}

// OAuth2ClientRepository định nghĩa interface cho việc truy cập dữ liệu client.
type OAuth2ClientRepository interface {
	GetClientByID(ctx context.Context, id uuid.UUID) (*OAuth2Client, error)
	ClientAssertionJWTValid(ctx context.Context, jti string) error
	SetClientAssertionJWT(ctx context.Context, jti string, exp time.Time) error
}
