package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	d "wibusystem/pkg/common/dto"
	"wibusystem/services/identify/repositories"
	"wibusystem/services/identify/services/interfaces"
)

// tenantRoleService provides functionality to manage tenant roles and assignments.
type tenantRoleService struct {
	rolesRepo      repositories.TenantRoleRepository
	membershipRepo repositories.MembershipRepository
}

// NewTenantRoleService creates a new tenant role service implementation.
func NewTenantRoleService(rolesRepo repositories.TenantRoleRepository, membershipRepo repositories.MembershipRepository) interfaces.TenantRoleServiceInterface {
	return &tenantRoleService{rolesRepo: rolesRepo, membershipRepo: membershipRepo}
}

func (s *tenantRoleService) ListPermissions(ctx context.Context) ([]*d.Permission, error) {
	perms, err := s.rolesRepo.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]*d.Permission, 0, len(perms))
	for _, perm := range perms {
		out = append(out, &d.Permission{
			ID:          perm.ID,
			Key:         perm.Key,
			Description: perm.Description,
		})
	}
	return out, nil
}

func (s *tenantRoleService) ListRoles(ctx context.Context, tenantID uuid.UUID) ([]*d.TenantRole, error) {
	roles, err := s.rolesRepo.ListRoles(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]*d.TenantRole, 0, len(roles))
	for _, role := range roles {
		out = append(out, &d.TenantRole{
			ID:          role.RoleID,
			TenantID:    role.TenantID,
			Name:        role.Name,
			Description: role.Description,
			IsDefault:   role.IsDefault,
			Permissions: role.Permissions,
		})
	}
	return out, nil
}

func (s *tenantRoleService) CreateRole(ctx context.Context, tenantID uuid.UUID, payload d.TenantRolePayload) (*d.TenantRole, error) {
	role, err := s.rolesRepo.CreateRole(ctx, tenantID, payload.Name, payload.Description, payload.IsDefault, payload.PermissionKeys)
	if err != nil {
		return nil, err
	}
	roles, err := s.rolesRepo.ListRoles(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	for _, r := range roles {
		if r.RoleID == role.ID {
			return &d.TenantRole{
				ID:          r.RoleID,
				TenantID:    r.TenantID,
				Name:        r.Name,
				Description: r.Description,
				IsDefault:   r.IsDefault,
				Permissions: r.Permissions,
			}, nil
		}
	}
	return nil, fmt.Errorf("failed to load role after creation")
}

func (s *tenantRoleService) UpdateRole(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID, payload d.TenantRolePayload) (*d.TenantRole, error) {
	role, err := s.rolesRepo.UpdateRole(ctx, roleID, tenantID, payload.Name, payload.Description, payload.IsDefault, payload.PermissionKeys)
	if err != nil {
		return nil, err
	}
	roles, err := s.rolesRepo.ListRoles(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	for _, r := range roles {
		if r.RoleID == role.ID {
			return &d.TenantRole{
				ID:          r.RoleID,
				TenantID:    r.TenantID,
				Name:        r.Name,
				Description: r.Description,
				IsDefault:   r.IsDefault,
				Permissions: r.Permissions,
			}, nil
		}
	}
	return nil, fmt.Errorf("failed to load role after update")
}

func (s *tenantRoleService) DeleteRole(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID) error {
	return s.rolesRepo.DeleteRole(ctx, roleID, tenantID)
}

func (s *tenantRoleService) AssignRoleToUser(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID, userID uuid.UUID) error {
	role, err := s.rolesRepo.GetRole(ctx, roleID)
	if err != nil {
		return err
	}
	if role == nil || role.TenantID == nil || *role.TenantID != tenantID {
		return fmt.Errorf("role does not belong to tenant")
	}
	membership, err := s.membershipRepo.GetByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to lookup membership: %w", err)
	}
	return s.rolesRepo.AssignRoleToMembership(ctx, roleID, membership.ID)
}

func (s *tenantRoleService) RemoveRoleFromUser(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID, userID uuid.UUID) error {
	membership, err := s.membershipRepo.GetByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to lookup membership: %w", err)
	}
	return s.rolesRepo.RemoveRoleFromMembership(ctx, roleID, membership.ID)
}
