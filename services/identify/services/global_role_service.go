package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	d "wibusystem/pkg/common/dto"
	"wibusystem/services/identify/repositories"
	"wibusystem/services/identify/services/interfaces"
)

// globalRoleService implements global role management logic.
type globalRoleService struct {
	repo repositories.GlobalRoleRepository
}

// NewGlobalRoleService creates a new service for managing global roles.
func NewGlobalRoleService(repo repositories.GlobalRoleRepository) interfaces.GlobalRoleServiceInterface {
	return &globalRoleService{repo: repo}
}

func (s *globalRoleService) ListRoles(ctx context.Context) ([]*d.GlobalRole, error) {
	roles, err := s.repo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	return convertGlobalRoles(roles), nil
}

func (s *globalRoleService) ListUserRoles(ctx context.Context, userID uuid.UUID) ([]*d.GlobalRole, error) {
	roles, err := s.repo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return convertGlobalRoles(roles), nil
}

func (s *globalRoleService) AssignRole(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error {
	exists, err := s.repo.RoleExists(ctx, roleID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("global role %s not found", roleID)
	}
	return s.repo.AssignRole(ctx, userID, roleID)
}

func (s *globalRoleService) RemoveRole(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error {
	exists, err := s.repo.RoleExists(ctx, roleID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("global role %s not found", roleID)
	}
	return s.repo.RemoveRole(ctx, userID, roleID)
}

func convertGlobalRoles(items []*repositories.GlobalRolePermissions) []*d.GlobalRole {
	out := make([]*d.GlobalRole, 0, len(items))
	for _, item := range items {
		out = append(out, &d.GlobalRole{
			ID:          item.RoleID,
			Name:        item.RoleName,
			Description: item.Description,
			Permissions: item.Permissions,
		})
	}
	return out
}
