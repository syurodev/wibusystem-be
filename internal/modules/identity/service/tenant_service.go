// Package service contains business logic for the Identity module.
package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"wibusystem/internal/modules/identity/domain"
	"wibusystem/internal/modules/identity/repository"
)

// TenantService defines the interface for tenant management operations.
type TenantService interface {
	// CreateTenant creates a new tenant owned by the specified user.
	CreateTenant(ctx context.Context, ownerID uuid.UUID, name, slug string, description *string) (*domain.Tenant, error)

	// GetTenant retrieves a tenant by ID.
	GetTenant(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error)

	// GetTenantBySlug retrieves a tenant by its slug.
	GetTenantBySlug(ctx context.Context, slug string) (*domain.Tenant, error)

	// UpdateTenant updates tenant information.
	UpdateTenant(ctx context.Context, tenantID uuid.UUID, name *string, description *string) (*domain.Tenant, error)

	// DeleteTenant soft-deletes a tenant.
	DeleteTenant(ctx context.Context, tenantID uuid.UUID) error

	// ListTenants lists all tenants with filtering.
	ListTenants(ctx context.Context, filter repository.TenantListFilter) ([]*domain.Tenant, int, error)

	// GetUserTenants retrieves all tenants a user is a member of.
	GetUserTenants(ctx context.Context, userID uuid.UUID) ([]*domain.Tenant, error)

	// GetUserOwnedTenants retrieves all tenants owned by a user.
	GetUserOwnedTenants(ctx context.Context, userID uuid.UUID) ([]*domain.Tenant, error)

	// AddMember adds a user to a tenant with a specific role.
	AddMember(ctx context.Context, tenantID, userID uuid.UUID, role domain.MemberRole, invitedBy uuid.UUID) (*domain.TenantMember, error)

	// RemoveMember removes a user from a tenant.
	RemoveMember(ctx context.Context, tenantID, userID uuid.UUID) error

	// UpdateMemberRole updates a member's role in a tenant.
	UpdateMemberRole(ctx context.Context, tenantID, userID uuid.UUID, newRole domain.MemberRole) error

	// UpdateMemberPermissions updates a member's custom permissions.
	UpdateMemberPermissions(ctx context.Context, tenantID, userID uuid.UUID, permissions []string) error

	// GetMember retrieves a specific tenant member.
	GetMember(ctx context.Context, tenantID, userID uuid.UUID) (*domain.TenantMember, error)

	// ListMembers lists all members of a tenant.
	ListMembers(ctx context.Context, tenantID uuid.UUID) ([]*domain.TenantMember, error)

	// ListMembersPaginated lists tenant members with pagination.
	ListMembersPaginated(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.TenantMember, int, error)

	// TransferOwnership transfers tenant ownership to another member.
	TransferOwnership(ctx context.Context, tenantID, newOwnerID uuid.UUID) error

	// CheckMemberPermission checks if a user has a specific permission in a tenant.
	CheckMemberPermission(ctx context.Context, tenantID, userID uuid.UUID, permission string) (bool, error)

	// IsMember checks if a user is a member of a tenant.
	IsMember(ctx context.Context, tenantID, userID uuid.UUID) (bool, error)

	// IsOwner checks if a user is the owner of a tenant.
	IsOwner(ctx context.Context, tenantID, userID uuid.UUID) (bool, error)

	// CanManageMembers checks if a user can manage members in a tenant.
	CanManageMembers(ctx context.Context, tenantID, userID uuid.UUID) (bool, error)

	// CanManageContent checks if a user can manage content in a tenant.
	CanManageContent(ctx context.Context, tenantID, userID uuid.UUID) (bool, error)

	// GetTenantStats retrieves statistics about a tenant.
	GetTenantStats(ctx context.Context, tenantID uuid.UUID) (*TenantStats, error)
}

// TenantStats contains statistics about a tenant.
type TenantStats struct {
	MemberCount int
	OwnerCount  int
	AdminCount  int
	EditorCount int
	ViewerCount int
}

// Tenant service errors
var (
	ErrTenantNotFound            = fmt.Errorf("tenant not found")
	ErrTenantAlreadyExists       = fmt.Errorf("tenant with this slug already exists")
	ErrTenantMemberNotFound      = fmt.Errorf("tenant member not found")
	ErrTenantMemberAlreadyExists = fmt.Errorf("user is already a member")
	ErrCannotRemoveOwner         = fmt.Errorf("cannot remove tenant owner")
	ErrCannotDemoteOwner         = fmt.Errorf("cannot demote tenant owner")
	ErrInvalidSlug               = fmt.Errorf("invalid tenant slug")
	ErrNotTenantMember           = fmt.Errorf("user is not a member of this tenant")
	ErrInsufficientPermissions   = fmt.Errorf("insufficient permissions")
)

// tenantServiceImpl is the implementation of TenantService.
type tenantServiceImpl struct {
	tenantRepo       repository.TenantRepository
	tenantMemberRepo repository.TenantMemberRepository
	userRepo         repository.UserRepository
}

// NewTenantService creates a new TenantService instance.
func NewTenantService(
	tenantRepo repository.TenantRepository,
	tenantMemberRepo repository.TenantMemberRepository,
	userRepo repository.UserRepository,
) TenantService {
	return &tenantServiceImpl{
		tenantRepo:       tenantRepo,
		tenantMemberRepo: tenantMemberRepo,
		userRepo:         userRepo,
	}
}

// CreateTenant creates a new tenant.
func (s *tenantServiceImpl) CreateTenant(ctx context.Context, ownerID uuid.UUID, name, slug string, description *string) (*domain.Tenant, error) {
	// Validate owner exists
	exists, err := s.userRepo.ExistsByID(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	// Check if slug already exists
	slugExists, err := s.tenantRepo.ExistsBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if slugExists {
		return nil, ErrTenantAlreadyExists
	}

	// Create tenant
	tenant, err := domain.NewTenant(name, slug, ownerID)
	if err != nil {
		return nil, err
	}

	if description != nil {
		tenant.UpdateDescription(*description)
	}

	// Save tenant
	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		if err == repository.ErrTenantAlreadyExists {
			return nil, ErrTenantAlreadyExists
		}
		return nil, err
	}

	// Add owner as a member
	ownerMember, err := domain.NewTenantMember(tenant.ID, ownerID, domain.MemberRoleOwner, nil)
	if err != nil {
		return nil, err
	}

	if err := s.tenantMemberRepo.Create(ctx, ownerMember); err != nil {
		// If member creation fails, we should ideally rollback tenant creation
		// For now, log the error
		return tenant, fmt.Errorf("tenant created but failed to add owner as member: %w", err)
	}

	return tenant, nil
}

// GetTenant retrieves a tenant by ID.
func (s *tenantServiceImpl) GetTenant(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		if err == repository.ErrTenantNotFound {
			return nil, ErrTenantNotFound
		}
		return nil, err
	}
	return tenant, nil
}

// GetTenantBySlug retrieves a tenant by slug.
func (s *tenantServiceImpl) GetTenantBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	tenant, err := s.tenantRepo.GetBySlug(ctx, slug)
	if err != nil {
		if err == repository.ErrTenantNotFound {
			return nil, ErrTenantNotFound
		}
		return nil, err
	}
	return tenant, nil
}

// UpdateTenant updates tenant information.
func (s *tenantServiceImpl) UpdateTenant(ctx context.Context, tenantID uuid.UUID, name *string, description *string) (*domain.Tenant, error) {
	// Get existing tenant
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		if err == repository.ErrTenantNotFound {
			return nil, ErrTenantNotFound
		}
		return nil, err
	}

	// Update fields if provided
	if name != nil {
		if err := tenant.UpdateName(*name); err != nil {
			return nil, err
		}
	}

	if description != nil {
		tenant.UpdateDescription(*description)
	}

	// Save changes
	if err := s.tenantRepo.Update(ctx, tenant); err != nil {
		return nil, err
	}

	return tenant, nil
}

// DeleteTenant soft-deletes a tenant.
func (s *tenantServiceImpl) DeleteTenant(ctx context.Context, tenantID uuid.UUID) error {
	// Check if tenant exists
	exists, err := s.tenantRepo.ExistsByID(ctx, tenantID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrTenantNotFound
	}

	// Delete tenant (soft delete)
	if err := s.tenantRepo.Delete(ctx, tenantID); err != nil {
		return err
	}

	// TODO: Consider cleaning up tenant members
	// _ = s.tenantMemberRepo.DeleteAllByTenant(ctx, tenantID)

	return nil
}

// ListTenants lists tenants with filtering.
func (s *tenantServiceImpl) ListTenants(ctx context.Context, filter repository.TenantListFilter) ([]*domain.Tenant, int, error) {
	tenants, total, err := s.tenantRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	return tenants, total, nil
}

// GetUserTenants retrieves all tenants a user is a member of.
func (s *tenantServiceImpl) GetUserTenants(ctx context.Context, userID uuid.UUID) ([]*domain.Tenant, error) {
	tenants, err := s.tenantRepo.GetUserTenants(ctx, userID)
	if err != nil {
		return nil, err
	}
	return tenants, nil
}

// GetUserOwnedTenants retrieves all tenants owned by a user.
func (s *tenantServiceImpl) GetUserOwnedTenants(ctx context.Context, userID uuid.UUID) ([]*domain.Tenant, error) {
	tenants, err := s.tenantRepo.GetByOwnerID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return tenants, nil
}

// AddMember adds a user to a tenant.
func (s *tenantServiceImpl) AddMember(ctx context.Context, tenantID, userID uuid.UUID, role domain.MemberRole, invitedBy uuid.UUID) (*domain.TenantMember, error) {
	// Verify tenant exists
	exists, err := s.tenantRepo.ExistsByID(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrTenantNotFound
	}

	// Verify user exists
	userExists, err := s.userRepo.ExistsByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !userExists {
		return nil, ErrUserNotFound
	}

	// Check if already a member
	alreadyMember, err := s.tenantMemberRepo.ExistsByTenantAndUser(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}
	if alreadyMember {
		return nil, ErrTenantMemberAlreadyExists
	}

	// Create member
	member, err := domain.NewTenantMember(tenantID, userID, role, &invitedBy)
	if err != nil {
		return nil, err
	}

	// Save member
	if err := s.tenantMemberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	return member, nil
}

// RemoveMember removes a user from a tenant.
func (s *tenantServiceImpl) RemoveMember(ctx context.Context, tenantID, userID uuid.UUID) error {
	// Check if member is owner
	isOwner, err := s.tenantMemberRepo.IsOwner(ctx, tenantID, userID)
	if err != nil {
		return err
	}
	if isOwner {
		return ErrCannotRemoveOwner
	}

	// Remove member
	if err := s.tenantMemberRepo.DeleteByTenantAndUser(ctx, tenantID, userID); err != nil {
		if err == repository.ErrTenantMemberNotFound {
			return ErrTenantMemberNotFound
		}
		if err == repository.ErrCannotDeleteOwner {
			return ErrCannotRemoveOwner
		}
		return err
	}

	return nil
}

// UpdateMemberRole updates a member's role.
func (s *tenantServiceImpl) UpdateMemberRole(ctx context.Context, tenantID, userID uuid.UUID, newRole domain.MemberRole) error {
	// Get member
	member, err := s.tenantMemberRepo.GetByTenantAndUser(ctx, tenantID, userID)
	if err != nil {
		if err == repository.ErrTenantMemberNotFound {
			return ErrTenantMemberNotFound
		}
		return err
	}

	// Cannot demote owner
	if member.Role == domain.MemberRoleOwner {
		return ErrCannotDemoteOwner
	}

	// Update role
	if err := member.UpdateRole(newRole); err != nil {
		return err
	}

	// Save changes
	if err := s.tenantMemberRepo.Update(ctx, member); err != nil {
		return err
	}

	return nil
}

// UpdateMemberPermissions updates member permissions.
func (s *tenantServiceImpl) UpdateMemberPermissions(ctx context.Context, tenantID, userID uuid.UUID, permissions []string) error {
	// Get member
	member, err := s.tenantMemberRepo.GetByTenantAndUser(ctx, tenantID, userID)
	if err != nil {
		if err == repository.ErrTenantMemberNotFound {
			return ErrTenantMemberNotFound
		}
		return err
	}

	// Update permissions
	member.Permissions = permissions

	// Save changes
	if err := s.tenantMemberRepo.Update(ctx, member); err != nil {
		return err
	}

	return nil
}

// GetMember retrieves a specific member.
func (s *tenantServiceImpl) GetMember(ctx context.Context, tenantID, userID uuid.UUID) (*domain.TenantMember, error) {
	member, err := s.tenantMemberRepo.GetByTenantAndUser(ctx, tenantID, userID)
	if err != nil {
		if err == repository.ErrTenantMemberNotFound {
			return nil, ErrTenantMemberNotFound
		}
		return nil, err
	}
	return member, nil
}

// ListMembers lists all tenant members.
func (s *tenantServiceImpl) ListMembers(ctx context.Context, tenantID uuid.UUID) ([]*domain.TenantMember, error) {
	members, err := s.tenantMemberRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	return members, nil
}

// ListMembersPaginated lists members with pagination.
func (s *tenantServiceImpl) ListMembersPaginated(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.TenantMember, int, error) {
	members, total, err := s.tenantMemberRepo.ListByTenantPaginated(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return members, total, nil
}

// TransferOwnership transfers ownership to another member.
func (s *tenantServiceImpl) TransferOwnership(ctx context.Context, tenantID, newOwnerID uuid.UUID) error {
	// Check if new owner is a member
	isMember, err := s.tenantMemberRepo.ExistsByTenantAndUser(ctx, tenantID, newOwnerID)
	if err != nil {
		return err
	}
	if !isMember {
		return ErrNotTenantMember
	}

	// Get current owner
	currentOwner, err := s.tenantMemberRepo.GetOwner(ctx, tenantID)
	if err != nil {
		return err
	}

	// Get new owner member
	newOwnerMember, err := s.tenantMemberRepo.GetByTenantAndUser(ctx, tenantID, newOwnerID)
	if err != nil {
		return err
	}

	// Update tenant owner
	if err := s.tenantRepo.UpdateOwner(ctx, tenantID, newOwnerID); err != nil {
		return err
	}

	// Update member roles
	// Demote current owner to admin
	currentOwner.Role = domain.MemberRoleAdmin
	if err := s.tenantMemberRepo.Update(ctx, currentOwner); err != nil {
		return err
	}

	// Promote new owner
	newOwnerMember.Role = domain.MemberRoleOwner
	if err := s.tenantMemberRepo.Update(ctx, newOwnerMember); err != nil {
		return err
	}

	return nil
}

// CheckMemberPermission checks if a user has a specific permission.
func (s *tenantServiceImpl) CheckMemberPermission(ctx context.Context, tenantID, userID uuid.UUID, permission string) (bool, error) {
	member, err := s.tenantMemberRepo.GetByTenantAndUser(ctx, tenantID, userID)
	if err != nil {
		if err == repository.ErrTenantMemberNotFound {
			return false, nil
		}
		return false, err
	}

	return member.HasPermission(permission), nil
}

// IsMember checks if a user is a member.
func (s *tenantServiceImpl) IsMember(ctx context.Context, tenantID, userID uuid.UUID) (bool, error) {
	exists, err := s.tenantMemberRepo.ExistsByTenantAndUser(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// IsOwner checks if a user is the owner.
func (s *tenantServiceImpl) IsOwner(ctx context.Context, tenantID, userID uuid.UUID) (bool, error) {
	isOwner, err := s.tenantMemberRepo.IsOwner(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}
	return isOwner, nil
}

// CanManageMembers checks if user can manage members.
func (s *tenantServiceImpl) CanManageMembers(ctx context.Context, tenantID, userID uuid.UUID) (bool, error) {
	member, err := s.tenantMemberRepo.GetByTenantAndUser(ctx, tenantID, userID)
	if err != nil {
		if err == repository.ErrTenantMemberNotFound {
			return false, nil
		}
		return false, err
	}

	return member.CanManageMembers(), nil
}

// CanManageContent checks if user can manage content.
func (s *tenantServiceImpl) CanManageContent(ctx context.Context, tenantID, userID uuid.UUID) (bool, error) {
	member, err := s.tenantMemberRepo.GetByTenantAndUser(ctx, tenantID, userID)
	if err != nil {
		if err == repository.ErrTenantMemberNotFound {
			return false, nil
		}
		return false, err
	}

	return member.CanManageContent(), nil
}

// GetTenantStats retrieves tenant statistics.
func (s *tenantServiceImpl) GetTenantStats(ctx context.Context, tenantID uuid.UUID) (*TenantStats, error) {
	stats := &TenantStats{}

	// Count total members
	total, err := s.tenantMemberRepo.CountByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	stats.MemberCount = total

	// Count by role
	owners, err := s.tenantMemberRepo.ListByTenantAndRole(ctx, tenantID, domain.MemberRoleOwner)
	if err != nil {
		return nil, err
	}
	stats.OwnerCount = len(owners)

	admins, err := s.tenantMemberRepo.ListByTenantAndRole(ctx, tenantID, domain.MemberRoleAdmin)
	if err != nil {
		return nil, err
	}
	stats.AdminCount = len(admins)

	editors, err := s.tenantMemberRepo.ListByTenantAndRole(ctx, tenantID, domain.MemberRoleEditor)
	if err != nil {
		return nil, err
	}
	stats.EditorCount = len(editors)

	viewers, err := s.tenantMemberRepo.ListByTenantAndRole(ctx, tenantID, domain.MemberRoleViewer)
	if err != nil {
		return nil, err
	}
	stats.ViewerCount = len(viewers)

	return stats, nil
}
