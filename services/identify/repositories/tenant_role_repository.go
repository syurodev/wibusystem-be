package repositories

import (
	"context"
	"errors"
	"fmt"

	m "wibusystem/pkg/common/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TenantRoleWithPermissions aggregates a tenant role with its granted permission keys.
type TenantRoleWithPermissions struct {
	RoleID      uuid.UUID
	TenantID    uuid.UUID
	Name        string
	Description *string
	IsDefault   bool
	Permissions []string
}

// TenantRoleRepository manages tenant roles and permission assignments.
type TenantRoleRepository interface {
	ListPermissions(ctx context.Context) ([]*m.Permission, error)
	ListRoles(ctx context.Context, tenantID uuid.UUID) ([]*TenantRoleWithPermissions, error)
	CreateRole(ctx context.Context, tenantID uuid.UUID, name string, description *string, isDefault bool, permissionKeys []string) (*m.Role, error)
	UpdateRole(ctx context.Context, roleID uuid.UUID, tenantID uuid.UUID, name string, description *string, isDefault bool, permissionKeys []string) (*m.Role, error)
	DeleteRole(ctx context.Context, roleID uuid.UUID, tenantID uuid.UUID) error
	GetRole(ctx context.Context, roleID uuid.UUID) (*m.Role, error)
	AssignRoleToMembership(ctx context.Context, roleID uuid.UUID, membershipID uuid.UUID) error
	RemoveRoleFromMembership(ctx context.Context, roleID uuid.UUID, membershipID uuid.UUID) error
}

type tenantRoleRepository struct {
	pool *pgxpool.Pool
}

// NewTenantRoleRepository creates a Postgres-backed TenantRoleRepository.
func NewTenantRoleRepository(pool *pgxpool.Pool) TenantRoleRepository {
	return &tenantRoleRepository{pool: pool}
}

func (r *tenantRoleRepository) ListPermissions(ctx context.Context) ([]*m.Permission, error) {
	query := `SELECT id, key, description FROM permissions ORDER BY key`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}
	defer rows.Close()

	var out []*m.Permission
	for rows.Next() {
		perm := &m.Permission{}
		if err := rows.Scan(&perm.ID, &perm.Key, &perm.Description); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		out = append(out, perm)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate permissions: %w", err)
	}

	return out, nil
}

func (r *tenantRoleRepository) ListRoles(ctx context.Context, tenantID uuid.UUID) ([]*TenantRoleWithPermissions, error) {
	query := `
        SELECT r.id,
               r.tenant_id,
               r.name,
               r.description,
               r.is_default,
               COALESCE(array_remove(array_agg(DISTINCT p.key), NULL), '{}') AS permissions
        FROM roles r
        LEFT JOIN role_permissions rp ON rp.role_id = r.id
        LEFT JOIN permissions p ON p.id = rp.permission_id
        WHERE r.tenant_id = $1
        GROUP BY r.id, r.tenant_id, r.name, r.description, r.is_default
        ORDER BY r.name
    `

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenant roles: %w", err)
	}
	defer rows.Close()

	var out []*TenantRoleWithPermissions
	for rows.Next() {
		var (
			roleID      uuid.UUID
			tenant      uuid.UUID
			name        string
			description *string
			isDefault   bool
			permissions []string
		)
		if err := rows.Scan(&roleID, &tenant, &name, &description, &isDefault, &permissions); err != nil {
			return nil, fmt.Errorf("failed to scan tenant role: %w", err)
		}
		out = append(out, &TenantRoleWithPermissions{
			RoleID:      roleID,
			TenantID:    tenant,
			Name:        name,
			Description: description,
			IsDefault:   isDefault,
			Permissions: permissions,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate tenant roles: %w", err)
	}

	return out, nil
}

func (r *tenantRoleRepository) CreateRole(ctx context.Context, tenantID uuid.UUID, name string, description *string, isDefault bool, permissionKeys []string) (*m.Role, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	role := &m.Role{TenantID: &tenantID, Name: name, Description: ""}
	query := `INSERT INTO roles (tenant_id, name, description, is_default) VALUES ($1, $2, $3, $4) RETURNING id`
	var roleID uuid.UUID
	if err := tx.QueryRow(ctx, query, tenantID, name, description, isDefault).Scan(&roleID); err != nil {
		return nil, fmt.Errorf("failed to insert role: %w", err)
	}
	role.ID = roleID
	role.Name = name
	role.Description = ""
	if description != nil {
		role.Description = *description
	}
	role.IsDefault = isDefault

	if len(permissionKeys) > 0 {
		permIDs, err := r.resolvePermissionIDs(ctx, tx, permissionKeys)
		if err != nil {
			return nil, err
		}
		if err := r.replaceRolePermissions(ctx, tx, roleID, permIDs); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit role creation: %w", err)
	}

	return role, nil
}

func (r *tenantRoleRepository) UpdateRole(ctx context.Context, roleID uuid.UUID, tenantID uuid.UUID, name string, description *string, isDefault bool, permissionKeys []string) (*m.Role, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `UPDATE roles SET name=$1, description=$2, is_default=$3 WHERE id=$4 AND tenant_id=$5`
	if cmdTag, err := tx.Exec(ctx, query, name, description, isDefault, roleID, tenantID); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	} else if cmdTag.RowsAffected() == 0 {
		return nil, fmt.Errorf("role not found for tenant")
	}

	if err := r.replaceRolePermissions(ctx, tx, roleID, nil); err != nil {
		return nil, err
	}
	if len(permissionKeys) > 0 {
		permIDs, err := r.resolvePermissionIDs(ctx, tx, permissionKeys)
		if err != nil {
			return nil, err
		}
		if err := r.replaceRolePermissions(ctx, tx, roleID, permIDs); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit role update: %w", err)
	}

	return &m.Role{ID: roleID, TenantID: &tenantID, Name: name, Description: mapString(description), IsDefault: isDefault}, nil
}

func (r *tenantRoleRepository) DeleteRole(ctx context.Context, roleID uuid.UUID, tenantID uuid.UUID) error {
	query := `DELETE FROM roles WHERE id = $1 AND tenant_id = $2`
	if cmdTag, err := r.pool.Exec(ctx, query, roleID, tenantID); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	} else if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("role not found for tenant")
	}
	return nil
}

func (r *tenantRoleRepository) GetRole(ctx context.Context, roleID uuid.UUID) (*m.Role, error) {
	query := `SELECT id, tenant_id, name, description, is_default FROM roles WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, roleID)
	role := &m.Role{}
	if err := row.Scan(&role.ID, &role.TenantID, &role.Name, &role.Description, &role.IsDefault); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	return role, nil
}

func (r *tenantRoleRepository) AssignRoleToMembership(ctx context.Context, roleID uuid.UUID, membershipID uuid.UUID) error {
	query := `INSERT INTO role_assignments (membership_id, role_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	if _, err := r.pool.Exec(ctx, query, membershipID, roleID); err != nil {
		return fmt.Errorf("failed to assign role to membership: %w", err)
	}
	return nil
}

func (r *tenantRoleRepository) RemoveRoleFromMembership(ctx context.Context, roleID uuid.UUID, membershipID uuid.UUID) error {
	query := `DELETE FROM role_assignments WHERE membership_id = $1 AND role_id = $2`
	if _, err := r.pool.Exec(ctx, query, membershipID, roleID); err != nil {
		return fmt.Errorf("failed to remove role from membership: %w", err)
	}
	return nil
}

func (r *tenantRoleRepository) resolvePermissionIDs(ctx context.Context, tx pgx.Tx, keys []string) ([]uuid.UUID, error) {
	query := `SELECT id, key FROM permissions WHERE key = ANY($1)`
	rows, err := tx.Query(ctx, query, keys)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve permission keys: %w", err)
	}
	defer rows.Close()

	found := make(map[string]uuid.UUID)
	for rows.Next() {
		var id uuid.UUID
		var key string
		if err := rows.Scan(&id, &key); err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		found[key] = id
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate permissions: %w", err)
	}

	if len(found) != len(keys) {
		missing := make([]string, 0)
		for _, key := range keys {
			if _, ok := found[key]; !ok {
				missing = append(missing, key)
			}
		}
		return nil, fmt.Errorf("unknown permission keys: %v", missing)
	}

	ids := make([]uuid.UUID, 0, len(keys))
	for _, key := range keys {
		ids = append(ids, found[key])
	}
	return ids, nil
}

func (r *tenantRoleRepository) replaceRolePermissions(ctx context.Context, tx pgx.Tx, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	if _, err := tx.Exec(ctx, `DELETE FROM role_permissions WHERE role_id = $1`, roleID); err != nil {
		return fmt.Errorf("failed to clear role permissions: %w", err)
	}
	if len(permissionIDs) == 0 {
		return nil
	}
	for _, permID := range permissionIDs {
		if _, err := tx.Exec(ctx, `INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2)`, roleID, permID); err != nil {
			return fmt.Errorf("failed to insert role permission: %w", err)
		}
	}
	return nil
}

func mapString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
