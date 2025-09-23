package interfaces

import (
	"context"

	"github.com/google/uuid"
	d "wibusystem/pkg/common/dto"
	m "wibusystem/pkg/common/model"
)

// TenantServiceInterface defines the contract for tenant operations
type TenantServiceInterface interface {
	// CreateTenant creates a new tenant
	CreateTenant(ctx context.Context, req d.CreateTenantRequest) (*m.Tenant, error)

	// GetTenantByID retrieves a tenant by ID
	GetTenantByID(ctx context.Context, tenantID uuid.UUID) (*m.Tenant, error)

	// GetTenantBySlug retrieves a tenant by slug
	GetTenantBySlug(ctx context.Context, slug string) (*m.Tenant, error)

	// ListTenants retrieves paginated list of all tenants (admin only)
	ListTenants(ctx context.Context, page, pageSize int) ([]*m.Tenant, int64, error)

	// GetUserTenants retrieves tenants for a specific user
	GetUserTenants(ctx context.Context, userID uuid.UUID) ([]*m.Tenant, error)

	// UpdateTenant updates tenant information
	UpdateTenant(ctx context.Context, tenantID uuid.UUID, req d.UpdateTenantRequest) (*m.Tenant, error)

	// DeleteTenant soft deletes a tenant
	DeleteTenant(ctx context.Context, tenantID uuid.UUID) error

	// AddUserToTenant adds a user to a tenant with specific role
	AddUserToTenant(ctx context.Context, tenantID, userID uuid.UUID, role string) error

	// RemoveUserFromTenant removes a user from a tenant
	RemoveUserFromTenant(ctx context.Context, tenantID, userID uuid.UUID) error

	// GetTenantMembers retrieves all members of a tenant with their roles
	GetTenantMembers(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]*m.Membership, int64, error)

	// UpdateUserRole updates a user's role in a tenant
	UpdateUserRole(ctx context.Context, tenantID, userID uuid.UUID, newRole string) error

	// ValidateTenantData validates tenant data before creation/update
	ValidateTenantData(req d.CreateTenantRequest) error

	// CheckSlugExists checks if tenant slug is already taken
	CheckSlugExists(ctx context.Context, slug string) (bool, error)

	// CheckUserTenantAccess checks if user has access to a tenant
	CheckUserTenantAccess(ctx context.Context, userID, tenantID uuid.UUID) (bool, error)
}