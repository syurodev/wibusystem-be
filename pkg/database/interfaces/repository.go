// Package interfaces defines repository abstractions for the wibusystem monorepo.
// Repositories handle business logic operations using database providers.
package interfaces

import (
	"context"

	"wibusystem/pkg/common"

	"github.com/google/uuid"
)

// UserRepository defines user-related database operations
type UserRepository interface {
	GetUser(ctx context.Context, id uuid.UUID) (*common.User, error)
	GetUserByEmail(ctx context.Context, email string) (*common.User, error)
	CreateUser(ctx context.Context, user *common.User) error
	UpdateUser(ctx context.Context, user *common.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
	ListUsers(ctx context.Context, limit, offset int) ([]*common.User, error)
}

// TenantRepository defines tenant-related database operations
type TenantRepository interface {
	GetTenant(ctx context.Context, id uuid.UUID) (*common.Tenant, error)
	CreateTenant(ctx context.Context, tenant *common.Tenant) error
	UpdateTenant(ctx context.Context, tenant *common.Tenant) error
	DeleteTenant(ctx context.Context, id uuid.UUID) error
	ListTenants(ctx context.Context, limit, offset int) ([]*common.Tenant, error)
}
