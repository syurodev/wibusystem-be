package oauth2

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
)

// OAuth2Service interface định nghĩa business logic cho OAuth2
type OAuth2Service interface {
	AuthenticateUser(ctx context.Context, identifier, password string) (*domain.User, error)
	CreateUserSession(ctx context.Context, userID uuid.UUID, ttl time.Duration, userAgent, ip string) (string, error)
	GetUserSession(ctx context.Context, sessionID string) (string, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	CheckUserConsent(ctx context.Context, userID, clientID uuid.UUID) (bool, error)
	CreateUserConsent(ctx context.Context, userID, clientID uuid.UUID, scopes []string) error
	LogoutUser(ctx context.Context, sessionID string, revokeTokens bool) error
	RevokeUserTokens(ctx context.Context, userID uuid.UUID) error
	GetClientInfo(ctx context.Context, clientID uuid.UUID) (*domain.OAuth2Client, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByIdentifier(ctx context.Context, identifier string) (*domain.User, error)
	GetGlobalPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
	GetOrganizationPermissions(ctx context.Context, userID, organizationID uuid.UUID) ([]string, error)
	GetGlobalRoles(ctx context.Context, userID uuid.UUID) ([]string, error)
	GetOrganizationRoles(ctx context.Context, userID, organizationID uuid.UUID) ([]string, error)
}

// WebAuthnService interface cho passkey operations
type WebAuthnService interface {
	HasPasskeys(ctx context.Context, userID uuid.UUID) (bool, error)
}

// OAuth2AdminService interface cho OAuth2 client management
type OAuth2AdminService interface {
	ValidateGrantTypes(grantTypes []string) (string, error)
	ValidateResponseTypes(responseTypes []string) (string, error)
	GenerateClientSecret() (plainSecret string, hashedSecret string, err error)
	CreateClient(ctx context.Context, req AdminCreateClientRequest) (*domain.OAuth2Client, *string, error)
	GetClientByID(ctx context.Context, clientID uuid.UUID) (*domain.OAuth2Client, error)
	ListClients(ctx context.Context, organizationID *uuid.UUID, active *bool, limit, offset int) ([]*domain.OAuth2Client, int, error)
	UpdateClient(ctx context.Context, clientID uuid.UUID, req AdminUpdateClientRequest) (*domain.OAuth2Client, error)
	DeleteClient(ctx context.Context, clientID uuid.UUID) error
	RegenerateClientSecret(ctx context.Context, clientID uuid.UUID) (string, error)
}
