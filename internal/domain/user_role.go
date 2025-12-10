package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

// UserGlobalRole là domain model cho việc gán global roles cho user
type UserGlobalRole struct {
	UserID    uuid.UUID
	RoleID    uuid.UUID
	Role      *Role // Optional: loaded by JOIN
	CreatedAt time.Time
}

// UserOrganizationRole là domain model cho việc gán roles trong context organization
type UserOrganizationRole struct {
	UserID         uuid.UUID
	OrganizationID uuid.UUID
	RoleID         uuid.UUID
	Role           *Role // Optional: loaded by JOIN
	CreatedAt      time.Time
}

// UserRoleRepository định nghĩa interface cho việc truy cập dữ liệu user roles
type UserRoleRepository interface {
	// Global roles
	GetGlobalRoles(ctx context.Context, userID uuid.UUID) ([]*UserGlobalRole, error)
	AddGlobalRole(ctx context.Context, userID, roleID uuid.UUID) error
	RemoveGlobalRole(ctx context.Context, userID, roleID uuid.UUID) error
	HasGlobalRole(ctx context.Context, userID, roleID uuid.UUID) (bool, error)

	// Organization roles
	GetOrganizationRoles(ctx context.Context, userID, organizationID uuid.UUID) ([]*UserOrganizationRole, error)
	AddOrganizationRole(ctx context.Context, userID, organizationID, roleID uuid.UUID) error
	RemoveOrganizationRole(ctx context.Context, userID, organizationID, roleID uuid.UUID) error
	HasOrganizationRole(ctx context.Context, userID, organizationID, roleID uuid.UUID) (bool, error)

	// Get all users with specific role
	GetUsersWithGlobalRole(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error)
	GetUsersWithOrganizationRole(ctx context.Context, organizationID, roleID uuid.UUID) ([]uuid.UUID, error)
}
