// ============================================================================
// User Role Repository (Ent Implementation)
// ============================================================================

package auth

import (
	"context"
	"system/internal/platform/database"
	"time"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
	ent "system/internal/ent/generated"
	"system/internal/ent/generated/userrole"
)

// userRoleEntRepository triển khai UserRoleRepository sử dụng Ent
type userRoleEntRepository struct {
	client *ent.Client
}

// NewUserRoleEntRepository tạo instance mới
func NewUserRoleEntRepository(client *ent.Client) domain.UserRoleRepository {
	return &userRoleEntRepository{client: client}
}

// =============================================================================
// Global Roles
// =============================================================================

// GetGlobalRoles lấy danh sách global roles của user
func (r *userRoleEntRepository) GetGlobalRoles(ctx context.Context, userID uuid.UUID) ([]*domain.UserGlobalRole, error) {
	roles, err := database.GetClientFromContext(ctx, r.client).UserRole.Query().
		Where(
			userrole.UserIDEQ(userID),
			userrole.OrganizationIDIsNil(),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]*domain.UserGlobalRole, len(roles))
	for i, r := range roles {
		results[i] = &domain.UserGlobalRole{
			UserID:    r.UserID,
			RoleID:    r.RoleID,
			CreatedAt: r.CreatedAt,
		}
	}
	return results, nil
}

// AddGlobalRole thêm global role cho user
func (r *userRoleEntRepository) AddGlobalRole(ctx context.Context, userID, roleID uuid.UUID) error {
	_, err := database.GetClientFromContext(ctx, r.client).UserRole.Create().
		SetUserID(userID).
		SetRoleID(roleID).
		SetAssignedAt(time.Now()).
		Save(ctx)
	return err
}

// RemoveGlobalRole xóa global role của user
func (r *userRoleEntRepository) RemoveGlobalRole(ctx context.Context, userID, roleID uuid.UUID) error {
	_, err := database.GetClientFromContext(ctx, r.client).UserRole.Delete().
		Where(
			userrole.UserIDEQ(userID),
			userrole.RoleIDEQ(roleID),
			userrole.OrganizationIDIsNil(),
		).
		Exec(ctx)
	return err
}

// HasGlobalRole kiểm tra user có global role không
func (r *userRoleEntRepository) HasGlobalRole(ctx context.Context, userID, roleID uuid.UUID) (bool, error) {
	return database.GetClientFromContext(ctx, r.client).UserRole.Query().
		Where(
			userrole.UserIDEQ(userID),
			userrole.RoleIDEQ(roleID),
			userrole.OrganizationIDIsNil(),
		).
		Exist(ctx)
}

// =============================================================================
// Organization Roles
// =============================================================================

// GetOrganizationRoles lấy danh sách organization roles của user
func (r *userRoleEntRepository) GetOrganizationRoles(ctx context.Context, userID, organizationID uuid.UUID) ([]*domain.UserOrganizationRole, error) {
	roles, err := database.GetClientFromContext(ctx, r.client).UserRole.Query().
		Where(
			userrole.UserIDEQ(userID),
			userrole.OrganizationIDEQ(organizationID),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]*domain.UserOrganizationRole, len(roles))
	for i, role := range roles {
		results[i] = &domain.UserOrganizationRole{
			UserID:         role.UserID,
			OrganizationID: *role.OrganizationID,
			RoleID:         role.RoleID,
			CreatedAt:      role.CreatedAt,
		}
	}
	return results, nil
}

// AddOrganizationRole thêm organization role cho user
func (r *userRoleEntRepository) AddOrganizationRole(ctx context.Context, userID, organizationID, roleID uuid.UUID) error {
	_, err := database.GetClientFromContext(ctx, r.client).UserRole.Create().
		SetUserID(userID).
		SetRoleID(roleID).
		SetOrganizationID(organizationID).
		SetAssignedAt(time.Now()).
		Save(ctx)
	return err
}

// RemoveOrganizationRole xóa organization role của user
func (r *userRoleEntRepository) RemoveOrganizationRole(ctx context.Context, userID, organizationID, roleID uuid.UUID) error {
	_, err := database.GetClientFromContext(ctx, r.client).UserRole.Delete().
		Where(
			userrole.UserIDEQ(userID),
			userrole.RoleIDEQ(roleID),
			userrole.OrganizationIDEQ(organizationID),
		).
		Exec(ctx)
	return err
}

// HasOrganizationRole kiểm tra user có organization role không
func (r *userRoleEntRepository) HasOrganizationRole(ctx context.Context, userID, organizationID, roleID uuid.UUID) (bool, error) {
	return database.GetClientFromContext(ctx, r.client).UserRole.Query().
		Where(
			userrole.UserIDEQ(userID),
			userrole.RoleIDEQ(roleID),
			userrole.OrganizationIDEQ(organizationID),
		).
		Exist(ctx)
}

// =============================================================================
// Get Users with Role
// =============================================================================

// GetUsersWithGlobalRole lấy danh sách users có global role
func (r *userRoleEntRepository) GetUsersWithGlobalRole(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error) {
	roles, err := database.GetClientFromContext(ctx, r.client).UserRole.Query().
		Where(
			userrole.RoleIDEQ(roleID),
			userrole.OrganizationIDIsNil(),
		).
		Select(userrole.FieldUserID).
		All(ctx)
	if err != nil {
		return nil, err
	}

	userIDs := make([]uuid.UUID, len(roles))
	for i, role := range roles {
		userIDs[i] = role.UserID
	}
	return userIDs, nil
}

// GetUsersWithOrganizationRole lấy danh sách users có organization role
func (r *userRoleEntRepository) GetUsersWithOrganizationRole(ctx context.Context, organizationID, roleID uuid.UUID) ([]uuid.UUID, error) {
	roles, err := database.GetClientFromContext(ctx, r.client).UserRole.Query().
		Where(
			userrole.RoleIDEQ(roleID),
			userrole.OrganizationIDEQ(organizationID),
		).
		Select(userrole.FieldUserID).
		All(ctx)
	if err != nil {
		return nil, err
	}

	userIDs := make([]uuid.UUID, len(roles))
	for i, role := range roles {
		userIDs[i] = role.UserID
	}
	return userIDs, nil
}
