// ============================================================================
// Role Repository (Ent Implementation)
// ============================================================================

package user

import (
	"context"
	"system/internal/platform/database"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
	ent "system/internal/ent/generated"
	"system/internal/ent/generated/role"
	"system/internal/ent/generated/userglobalrole"
	"system/internal/ent/generated/userorganizationrole"
)

// roleEntRepository triển khai RoleRepository sử dụng Ent
type roleEntRepository struct {
	client *ent.Client
}

// NewRoleEntRepository tạo instance mới
func NewRoleEntRepository(client *ent.Client) domain.RoleRepository {
	return &roleEntRepository{client: client}
}

// GetRoleIDByName gets the role ID by name
func (r *roleEntRepository) GetRoleIDByName(ctx context.Context, name domain.Role) (uuid.UUID, error) {
	roleEntity, err := database.GetClientFromContext(ctx, r.client).Role.Query().
		Where(role.NameEQ(name.String())).
		Only(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return roleEntity.ID, nil
}

// AssignGlobalRole assigns a global role to a user
func (r *roleEntRepository) AssignGlobalRole(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error {
	// Check if exists
	exists, err := database.GetClientFromContext(ctx, r.client).UserGlobalRole.Query().
		Where(
			userglobalrole.UserIDEQ(userID),
			userglobalrole.RoleIDEQ(roleID),
		).
		Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil // Already assigned
	}

	_, err = database.GetClientFromContext(ctx, r.client).UserGlobalRole.Create().
		SetUserID(userID).
		SetRoleID(roleID).
		Save(ctx)
	return err
}

// RemoveGlobalRole removes a global role from a user
func (r *roleEntRepository) RemoveGlobalRole(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error {
	_, err := database.GetClientFromContext(ctx, r.client).UserGlobalRole.Delete().
		Where(
			userglobalrole.UserIDEQ(userID),
			userglobalrole.RoleIDEQ(roleID),
		).
		Exec(ctx)
	return err
}

// AssignOrganizationRole assigns a role to a user within an organization
func (r *roleEntRepository) AssignOrganizationRole(ctx context.Context, userID, organizationID, roleID uuid.UUID) error {
	// Check if exists
	exists, err := database.GetClientFromContext(ctx, r.client).UserOrganizationRole.Query().
		Where(
			userorganizationrole.UserIDEQ(userID),
			userorganizationrole.OrganizationIDEQ(organizationID),
			userorganizationrole.RoleIDEQ(roleID),
		).
		Exist(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil // Already assigned
	}

	_, err = database.GetClientFromContext(ctx, r.client).UserOrganizationRole.Create().
		SetUserID(userID).
		SetOrganizationID(organizationID).
		SetRoleID(roleID).
		Save(ctx)
	return err
}

// RemoveOrganizationRole removes a role from a user within an organization
func (r *roleEntRepository) RemoveOrganizationRole(ctx context.Context, userID, organizationID, roleID uuid.UUID) error {
	_, err := database.GetClientFromContext(ctx, r.client).UserOrganizationRole.Delete().
		Where(
			userorganizationrole.UserIDEQ(userID),
			userorganizationrole.OrganizationIDEQ(organizationID),
			userorganizationrole.RoleIDEQ(roleID),
		).
		Exec(ctx)
	return err
}

// HasGlobalRole checks if a user has a specific global role
func (r *roleEntRepository) HasGlobalRole(ctx context.Context, userID uuid.UUID, roleName domain.Role) (bool, error) {
	// Get role ID first
	roleEntity, err := database.GetClientFromContext(ctx, r.client).Role.Query().
		Where(role.NameEQ(roleName.String())).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return database.GetClientFromContext(ctx, r.client).UserGlobalRole.Query().
		Where(
			userglobalrole.UserIDEQ(userID),
			userglobalrole.RoleIDEQ(roleEntity.ID),
		).
		Exist(ctx)
}

// HasOrganizationRole checks if a user has a specific role in an organization
func (r *roleEntRepository) HasOrganizationRole(ctx context.Context, userID, organizationID uuid.UUID, roleName domain.Role) (bool, error) {
	// Get role ID first
	roleEntity, err := database.GetClientFromContext(ctx, r.client).Role.Query().
		Where(role.NameEQ(roleName.String())).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return database.GetClientFromContext(ctx, r.client).UserOrganizationRole.Query().
		Where(
			userorganizationrole.UserIDEQ(userID),
			userorganizationrole.OrganizationIDEQ(organizationID),
			userorganizationrole.RoleIDEQ(roleEntity.ID),
		).
		Exist(ctx)
}
