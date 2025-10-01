package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GlobalRoleRepository defines accessors for global roles and assignments.
type GlobalRoleRepository interface {
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*GlobalRolePermissions, error)
	ListAll(ctx context.Context) ([]*GlobalRolePermissions, error)
	AssignRole(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error
	RemoveRole(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error
	RoleExists(ctx context.Context, roleID uuid.UUID) (bool, error)
}

// GlobalRolePermissions aggregates role metadata with its permissions.
type GlobalRolePermissions struct {
	RoleID      uuid.UUID
	RoleName    string
	Description *string
	Permissions []string
}

type globalRoleRepository struct {
	pool *pgxpool.Pool
}

// NewGlobalRoleRepository creates a Postgres-backed GlobalRoleRepository.
func NewGlobalRoleRepository(pool *pgxpool.Pool) GlobalRoleRepository {
	return &globalRoleRepository{pool: pool}
}

// ListByUserID returns all global roles (and their permissions) assigned to a user.
func (r *globalRoleRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*GlobalRolePermissions, error) {
	query := `
        SELECT gr.id,
               gr.name,
               gr.description,
               COALESCE(array_remove(array_agg(DISTINCT gp.key), NULL), '{}') AS permissions
        FROM user_global_roles ugr
        JOIN global_roles gr ON gr.id = ugr.role_id
        LEFT JOIN global_role_permissions grp ON grp.role_id = gr.id
        LEFT JOIN global_permissions gp ON gp.id = grp.permission_id
        WHERE ugr.user_id = $1
        GROUP BY gr.id, gr.name, gr.description
        ORDER BY gr.name
    `

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query global roles for user: %w", err)
	}
	defer rows.Close()

	var out []*GlobalRolePermissions
	for rows.Next() {
		var (
			roleID      uuid.UUID
			roleName    string
			description *string
			perms       []string
		)
		if err := rows.Scan(&roleID, &roleName, &description, &perms); err != nil {
			return nil, fmt.Errorf("failed to scan global role row: %w", err)
		}

		out = append(out, &GlobalRolePermissions{
			RoleID:      roleID,
			RoleName:    roleName,
			Description: description,
			Permissions: perms,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate global roles: %w", err)
	}

	return out, nil
}

// ListAll returns all global roles and their permissions.
func (r *globalRoleRepository) ListAll(ctx context.Context) ([]*GlobalRolePermissions, error) {
	query := `
		SELECT gr.id,
		       gr.name,
		       gr.description,
		       COALESCE(array_remove(array_agg(DISTINCT gp.key), NULL), '{}') AS permissions
		FROM global_roles gr
		LEFT JOIN global_role_permissions grp ON grp.role_id = gr.id
		LEFT JOIN global_permissions gp ON gp.id = grp.permission_id
		GROUP BY gr.id, gr.name, gr.description
		ORDER BY gr.name
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list global roles: %w", err)
	}
	defer rows.Close()

	var out []*GlobalRolePermissions
	for rows.Next() {
		var (
			roleID      uuid.UUID
			roleName    string
			description *string
			permissions []string
		)
		if err := rows.Scan(&roleID, &roleName, &description, &permissions); err != nil {
			return nil, fmt.Errorf("failed to scan global role row: %w", err)
		}

		out = append(out, &GlobalRolePermissions{
			RoleID:      roleID,
			RoleName:    roleName,
			Description: description,
			Permissions: permissions,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate global roles: %w", err)
	}

	return out, nil
}

// AssignRole assigns a global role to the specified user (idempotent).
func (r *globalRoleRepository) AssignRole(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error {
	query := `
		INSERT INTO user_global_roles (user_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, role_id) DO NOTHING
	`

	if _, err := r.pool.Exec(ctx, query, userID, roleID); err != nil {
		return fmt.Errorf("failed to assign global role to user: %w", err)
	}

	return nil
}

// RemoveRole removes a global role assignment from a user.
func (r *globalRoleRepository) RemoveRole(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error {
	query := `DELETE FROM user_global_roles WHERE user_id = $1 AND role_id = $2`
	if _, err := r.pool.Exec(ctx, query, userID, roleID); err != nil {
		return fmt.Errorf("failed to remove global role from user: %w", err)
	}
	return nil
}

// RoleExists checks whether a global role exists.
func (r *globalRoleRepository) RoleExists(ctx context.Context, roleID uuid.UUID) (bool, error) {
	query := `SELECT 1 FROM global_roles WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, roleID)
	var dummy int
	if err := row.Scan(&dummy); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check global role existence: %w", err)
	}
	return true, nil
}
