// Package repository defines interfaces for data persistence in the Identity module.
package repository

import (
	"context"

	"github.com/google/uuid"

	"wibusystem/internal/modules/identity/domain"
)

// TenantRepository defines the interface for tenant data persistence.
type TenantRepository interface {
	// Create creates a new tenant in the database.
	Create(ctx context.Context, tenant *domain.Tenant) error

	// GetByID retrieves a tenant by its ID.
	// Returns ErrTenantNotFound if the tenant doesn't exist.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error)

	// GetBySlug retrieves a tenant by its slug.
	// Returns ErrTenantNotFound if the tenant doesn't exist.
	GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error)

	// Update updates an existing tenant's information.
	// Returns ErrTenantNotFound if the tenant doesn't exist.
	Update(ctx context.Context, tenant *domain.Tenant) error

	// Delete soft-deletes a tenant by its ID.
	// Returns ErrTenantNotFound if the tenant doesn't exist.
	Delete(ctx context.Context, id uuid.UUID) error

	// HardDelete permanently deletes a tenant from the database.
	// Use with caution - this operation is irreversible.
	HardDelete(ctx context.Context, id uuid.UUID) error

	// List retrieves a paginated list of tenants based on the provided filter.
	// Returns the tenants, total count, and any error.
	List(ctx context.Context, filter TenantListFilter) ([]*domain.Tenant, int, error)

	// GetByOwnerID retrieves all tenants owned by a specific user.
	GetByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*domain.Tenant, error)

	// GetUserTenants retrieves all tenants a user is a member of.
	GetUserTenants(ctx context.Context, userID uuid.UUID) ([]*domain.Tenant, error)

	// ExistsBySlug checks if a tenant with the given slug exists.
	ExistsBySlug(ctx context.Context, slug string) (bool, error)

	// ExistsByID checks if a tenant with the given ID exists.
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)

	// CountAll returns the total number of tenants (including soft-deleted if specified).
	CountAll(ctx context.Context, includeDeleted bool) (int, error)

	// CountByStatus returns the number of tenants with a specific status.
	CountByStatus(ctx context.Context, status domain.TenantStatus) (int, error)

	// UpdateStatus updates a tenant's status.
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.TenantStatus) error

	// UpdateOwner transfers ownership of a tenant to another user.
	UpdateOwner(ctx context.Context, tenantID, newOwnerID uuid.UUID) error

	// Search searches for tenants by name or slug.
	Search(ctx context.Context, query string, limit int) ([]*domain.Tenant, error)

	// GetByIDs retrieves multiple tenants by their IDs.
	// Returns tenants that exist, skips non-existent ones.
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*domain.Tenant, error)
}

// TenantMemberRepository defines the interface for tenant member data persistence.
type TenantMemberRepository interface {
	// Create creates a new tenant member in the database.
	Create(ctx context.Context, member *domain.TenantMember) error

	// GetByID retrieves a tenant member by its ID.
	// Returns ErrTenantMemberNotFound if the member doesn't exist.
	GetByID(ctx context.Context, id uuid.UUID) (*domain.TenantMember, error)

	// GetByTenantAndUser retrieves a tenant member by tenant ID and user ID.
	// Returns ErrTenantMemberNotFound if the member doesn't exist.
	GetByTenantAndUser(ctx context.Context, tenantID, userID uuid.UUID) (*domain.TenantMember, error)

	// Update updates an existing tenant member's information.
	// Returns ErrTenantMemberNotFound if the member doesn't exist.
	Update(ctx context.Context, member *domain.TenantMember) error

	// Delete removes a tenant member from the database.
	// Returns ErrTenantMemberNotFound if the member doesn't exist.
	Delete(ctx context.Context, id uuid.UUID) error

	// DeleteByTenantAndUser removes a tenant member by tenant ID and user ID.
	// Returns ErrTenantMemberNotFound if the member doesn't exist.
	DeleteByTenantAndUser(ctx context.Context, tenantID, userID uuid.UUID) error

	// ListByTenant retrieves all members of a specific tenant.
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.TenantMember, error)

	// ListByTenantPaginated retrieves members of a tenant with pagination.
	ListByTenantPaginated(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.TenantMember, int, error)

	// ListByUser retrieves all tenant memberships for a specific user.
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.TenantMember, error)

	// ListByTenantAndRole retrieves all members of a tenant with a specific role.
	ListByTenantAndRole(ctx context.Context, tenantID uuid.UUID, role domain.MemberRole) ([]*domain.TenantMember, error)

	// CountByTenant returns the number of members in a specific tenant.
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error)

	// CountByUser returns the number of tenants a user is a member of.
	CountByUser(ctx context.Context, userID uuid.UUID) (int, error)

	// ExistsByTenantAndUser checks if a user is a member of a tenant.
	ExistsByTenantAndUser(ctx context.Context, tenantID, userID uuid.UUID) (bool, error)

	// UpdateRole updates a member's role in a tenant.
	UpdateRole(ctx context.Context, id uuid.UUID, role domain.MemberRole) error

	// UpdatePermissions updates a member's permissions in a tenant.
	UpdatePermissions(ctx context.Context, id uuid.UUID, permissions []string) error

	// GetOwner retrieves the owner of a tenant.
	GetOwner(ctx context.Context, tenantID uuid.UUID) (*domain.TenantMember, error)

	// IsOwner checks if a user is the owner of a tenant.
	IsOwner(ctx context.Context, tenantID, userID uuid.UUID) (bool, error)

	// HasRole checks if a user has a specific role in a tenant.
	HasRole(ctx context.Context, tenantID, userID uuid.UUID, role domain.MemberRole) (bool, error)

	// GetUserRole retrieves a user's role in a tenant.
	GetUserRole(ctx context.Context, tenantID, userID uuid.UUID) (domain.MemberRole, error)

	// DeleteAllByTenant removes all members from a tenant (when deleting tenant).
	DeleteAllByTenant(ctx context.Context, tenantID uuid.UUID) error

	// DeleteAllByUser removes all tenant memberships for a user (when deleting user).
	DeleteAllByUser(ctx context.Context, userID uuid.UUID) error
}

// TenantListFilter contains filtering and pagination options for listing tenants.
type TenantListFilter struct {
	// Pagination
	Limit  int // Maximum number of results to return
	Offset int // Number of results to skip

	// Filters
	Status         *domain.TenantStatus // Filter by tenant status
	OwnerID        *uuid.UUID           // Filter by owner ID
	IncludeDeleted bool                 // Include soft-deleted tenants

	// Search
	NameContains string // Filter by name containing string
	SlugContains string // Filter by slug containing string

	// Sorting
	SortBy    string // Field to sort by (created_at, name, slug)
	SortOrder string // Sort order (asc, desc)
}

// Repository errors for Tenant
var (
	// ErrTenantNotFound is returned when a tenant cannot be found.
	ErrTenantNotFound = NewRepositoryError("tenant not found", "TENANT_NOT_FOUND")

	// ErrTenantAlreadyExists is returned when attempting to create a tenant with a slug that already exists.
	ErrTenantAlreadyExists = NewRepositoryError("tenant with this slug already exists", "TENANT_ALREADY_EXISTS")

	// ErrInvalidTenantData is returned when tenant data is invalid.
	ErrInvalidTenantData = NewRepositoryError("invalid tenant data", "INVALID_TENANT_DATA")

	// ErrTenantMemberNotFound is returned when a tenant member cannot be found.
	ErrTenantMemberNotFound = NewRepositoryError("tenant member not found", "TENANT_MEMBER_NOT_FOUND")

	// ErrTenantMemberAlreadyExists is returned when a user is already a member of a tenant.
	ErrTenantMemberAlreadyExists = NewRepositoryError("user is already a member of this tenant", "TENANT_MEMBER_ALREADY_EXISTS")

	// ErrInvalidTenantMemberData is returned when tenant member data is invalid.
	ErrInvalidTenantMemberData = NewRepositoryError("invalid tenant member data", "INVALID_TENANT_MEMBER_DATA")

	// ErrCannotDeleteOwner is returned when attempting to delete a tenant owner.
	ErrCannotDeleteOwner = NewRepositoryError("cannot delete tenant owner", "CANNOT_DELETE_OWNER")
)
