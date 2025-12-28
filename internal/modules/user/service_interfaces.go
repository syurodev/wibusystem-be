package user

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
	userdto "system/internal/dto/user"
)

// UserService interface định nghĩa business logic cho User module
type UserService interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
	CreateUser(ctx context.Context, user *domain.User) error
	UpdateUser(ctx context.Context, user *domain.User) error

	GetSettings(ctx context.Context, userID uuid.UUID) (map[string]any, error)
	UpdateSettings(ctx context.Context, userID uuid.UUID, newSettings map[string]any) (map[string]any, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req userdto.UpdateProfileRequest) (*domain.User, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error
	GetSessions(ctx context.Context, userIDStr string) ([]*domain.UserSession, error)
	DeleteSession(ctx context.Context, userIDStr, sessionID string) error

	// Methods for OAuth2Service integration
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
	CreateSession(ctx context.Context, session *domain.UserSession, ttl time.Duration) error
	GetSession(ctx context.Context, sessionID string) (*domain.UserSession, error)

	GetGlobalPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
	GetOrganizationPermissions(ctx context.Context, userID, organizationID uuid.UUID) ([]string, error)
	GetGlobalRoles(ctx context.Context, userID uuid.UUID) ([]string, error)
	GetOrganizationRoles(ctx context.Context, userID, organizationID uuid.UUID) ([]string, error)
}
