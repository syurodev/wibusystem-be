package user

import (
	"context"
	"system/internal/domain"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// roleRepository triển khai RoleRepository sử dụng pgx
type roleRepository struct {
	pool *pgxpool.Pool
}

// NewRoleRepository tạo một instance mới của roleRepository
func NewRoleRepository(pool *pgxpool.Pool) domain.RoleRepository {
	return &roleRepository{pool: pool}
}

// GetRoleIDByName gets the role ID by name
func (r *roleRepository) GetRoleIDByName(ctx context.Context, name domain.Role) (uuid.UUID, error) {
	query := `SELECT id FROM identify.roles WHERE name = $1`

	var id uuid.UUID
	err := r.pool.QueryRow(ctx, query, name.String()).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}

// AssignGlobalRole assigns a global role to a user
func (r *roleRepository) AssignGlobalRole(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error {
	query := `
		INSERT INTO identify.user_global_roles (user_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query, userID, roleID)
	return err
}

// RemoveGlobalRole removes a global role from a user
func (r *roleRepository) RemoveGlobalRole(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error {
	query := `
		DELETE FROM identify.user_global_roles
		WHERE user_id = $1 AND role_id = $2
	`

	_, err := r.pool.Exec(ctx, query, userID, roleID)
	return err
}

// AssignOrganizationRole assigns a role to a user within an organization
func (r *roleRepository) AssignOrganizationRole(ctx context.Context, userID, organizationID, roleID uuid.UUID) error {
	query := `
		INSERT INTO identify.user_organization_roles (user_id, organization_id, role_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, organization_id, role_id) DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query, userID, organizationID, roleID)
	return err
}

// RemoveOrganizationRole removes a role from a user within an organization
func (r *roleRepository) RemoveOrganizationRole(ctx context.Context, userID, organizationID, roleID uuid.UUID) error {
	query := `
		DELETE FROM identify.user_organization_roles
		WHERE user_id = $1 AND organization_id = $2 AND role_id = $3
	`

	_, err := r.pool.Exec(ctx, query, userID, organizationID, roleID)
	return err
}

// HasGlobalRole checks if a user has a specific global role
func (r *roleRepository) HasGlobalRole(ctx context.Context, userID uuid.UUID, roleName domain.Role) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM identify.user_global_roles ugr
			JOIN identify.roles r ON ugr.role_id = r.id
			WHERE ugr.user_id = $1 AND r.name = $2
		)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, userID, roleName.String()).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// HasOrganizationRole checks if a user has a specific role in an organization
func (r *roleRepository) HasOrganizationRole(ctx context.Context, userID, organizationID uuid.UUID, roleName domain.Role) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM identify.user_organization_roles uor
			JOIN identify.roles r ON uor.role_id = r.id
			WHERE uor.user_id = $1 AND uor.organization_id = $2 AND r.name = $3
		)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, userID, organizationID, roleName.String()).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
