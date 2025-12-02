package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

// OAuth2Client là domain model cho một OAuth2 client.
type OAuth2Client struct {
	ID                uuid.UUID  `db:"id"`
	ClientName        string     `db:"client_name"`
	SecretHash        string     `db:"secret_hash"`
	RedirectURIs      []string   `db:"redirect_uris"`
	GrantTypes        []string   `db:"grant_types"`
	ResponseTypes     []string   `db:"response_types"`
	Scopes            []string   `db:"scopes"`
	IsPublic          bool       `db:"is_public"`
	IsInternal        bool       `db:"is_internal"`
	TokenEndpointAuth string     `db:"token_endpoint_auth_method"`
	OrganizationID    *uuid.UUID `db:"organization_id"`
	ClientURI         *string    `db:"client_uri"`
	LogoURL           *string    `db:"logo_url"`
	Active            bool       `db:"active"`
	CreatedAt         time.Time  `db:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at"`
}

// OAuth2ClientRepository định nghĩa interface cho việc truy cập dữ liệu client.
type OAuth2ClientRepository interface {
	// Legacy methods
	GetClientByID(ctx context.Context, id uuid.UUID) (*OAuth2Client, error)
	ClientAssertionJWTValid(ctx context.Context, jti string) error
	SetClientAssertionJWT(ctx context.Context, jti string, exp time.Time) error

	// CRUD methods for admin API
	GetByID(ctx context.Context, id uuid.UUID) (*OAuth2Client, error)
	Create(ctx context.Context, client *OAuth2Client) error
	Update(ctx context.Context, client *OAuth2Client) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, organizationID *uuid.UUID, active *bool, limit, offset int) ([]*OAuth2Client, int, error)
}
