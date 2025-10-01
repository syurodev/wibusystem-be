package repositories

import (
	"context"
	"fmt"

	m "wibusystem/pkg/common/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MembershipRepository defines operations to manage tenant memberships.
type MembershipRepository interface {
	Create(ctx context.Context, membership *m.Membership) error
	GetByID(ctx context.Context, id uuid.UUID) (*m.Membership, error)
	GetByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) (*m.Membership, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*m.Membership, error)
	ListByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*m.Membership, error)
	Update(ctx context.Context, membership *m.Membership) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	ListRolesWithPermissionsByUserID(ctx context.Context, userID uuid.UUID) ([]*MembershipRolePermissions, error)
}

// MembershipRolePermissions aggregates role metadata and permissions per tenant for a membership.
type MembershipRolePermissions struct {
	TenantID    uuid.UUID
	RoleID      *uuid.UUID
	RoleName    *string
	Permissions []string
}

type membershipRepository struct {
	pool *pgxpool.Pool
}

// NewMembershipRepository creates a Postgres-backed MembershipRepository.
func NewMembershipRepository(pool *pgxpool.Pool) MembershipRepository {
	return &membershipRepository{pool: pool}
}

// Create inserts a new membership (user <-> tenant association).
func (r *membershipRepository) Create(ctx context.Context, membership *m.Membership) error {
	query := `
		INSERT INTO memberships (user_id, tenant_id, role_id, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, joined_at
	`

	err := r.pool.QueryRow(
		ctx, query,
		membership.UserID,
		membership.TenantID,
		membership.RoleID,
		membership.Status,
	).Scan(
		&membership.ID,
		&membership.JoinedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create membership: %w", err)
	}

	return nil
}

// GetByID returns a membership by ID.
func (r *membershipRepository) GetByID(ctx context.Context, id uuid.UUID) (*m.Membership, error) {
	query := `
		SELECT id, user_id, tenant_id, role_id, status, joined_at
		FROM memberships
		WHERE id = $1
	`

	membership := &m.Membership{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&membership.ID,
		&membership.UserID,
		&membership.TenantID,
		&membership.RoleID,
		&membership.Status,
		&membership.JoinedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get membership by ID: %w", err)
	}

	return membership, nil
}

// GetByUserAndTenant returns the membership for a given user and tenant.
func (r *membershipRepository) GetByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) (*m.Membership, error) {
	query := `
		SELECT id, user_id, tenant_id, role_id, status, joined_at
		FROM memberships
		WHERE user_id = $1 AND tenant_id = $2
	`

	membership := &m.Membership{}
	err := r.pool.QueryRow(ctx, query, userID, tenantID).Scan(
		&membership.ID,
		&membership.UserID,
		&membership.TenantID,
		&membership.RoleID,
		&membership.Status,
		&membership.JoinedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get membership by user and tenant: %w", err)
	}

	return membership, nil
}

// ListByUserID returns memberships for the given user.
func (r *membershipRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*m.Membership, error) {
	query := `
		SELECT id, user_id, tenant_id, role_id, status, joined_at
		FROM memberships
		WHERE user_id = $1
		ORDER BY joined_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list memberships by user ID: %w", err)
	}
	defer rows.Close()

	var memberships []*m.Membership
	for rows.Next() {
		membership := &m.Membership{}
		err := rows.Scan(
			&membership.ID,
			&membership.UserID,
			&membership.TenantID,
			&membership.RoleID,
			&membership.Status,
			&membership.JoinedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan membership: %w", err)
		}

		memberships = append(memberships, membership)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate memberships: %w", err)
	}

	return memberships, nil
}

// ListByTenantID returns memberships for the given tenant.
func (r *membershipRepository) ListByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*m.Membership, error) {
	query := `
		SELECT m.id, m.user_id, m.tenant_id, m.role_id, m.status, m.joined_at
		FROM memberships m
		WHERE m.tenant_id = $1
		ORDER BY m.joined_at DESC
	`

	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list memberships by tenant ID: %w", err)
	}
	defer rows.Close()

	var memberships []*m.Membership
	for rows.Next() {
		membership := &m.Membership{}
		err := rows.Scan(
			&membership.ID,
			&membership.UserID,
			&membership.TenantID,
			&membership.RoleID,
			&membership.Status,
			&membership.JoinedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan membership: %w", err)
		}

		memberships = append(memberships, membership)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate memberships: %w", err)
	}

	return memberships, nil
}

// Update modifies role/status for a membership.
func (r *membershipRepository) Update(ctx context.Context, membership *m.Membership) error {
	query := `
		UPDATE memberships
		SET role_id = $2, status = $3
		WHERE id = $1
	`

	result, err := r.pool.Exec(
		ctx, query,
		membership.ID,
		membership.RoleID,
		membership.Status,
	)
	if err != nil {
		return fmt.Errorf("failed to update membership: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("membership with ID %d not found", membership.ID)
	}

	return nil
}

// Delete removes a membership by ID.
func (r *membershipRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM memberships WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete membership: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("membership with ID %d not found", id)
	}

	return nil
}

// UpdateStatus changes the membership status (e.g., active/inactive).
func (r *membershipRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE memberships SET status = $2 WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update membership status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("membership with ID %d not found", id)
	}

	return nil
}

// ListRolesWithPermissionsByUserID returns each tenant-role assignment with the associated permissions for the user.
func (r *membershipRepository) ListRolesWithPermissionsByUserID(ctx context.Context, userID uuid.UUID) ([]*MembershipRolePermissions, error) {
	query := `
	WITH membership_roles AS (
	    SELECT m.id AS membership_id,
	           m.tenant_id,
	           COALESCE(ra.role_id, m.role_id) AS role_id
	    FROM memberships m
	    LEFT JOIN role_assignments ra ON ra.membership_id = m.id
	    WHERE m.user_id = $1 AND m.status = 'active'
	)
	SELECT mr.tenant_id,
	       r.id,
	       r.name,
	       COALESCE(array_remove(array_agg(DISTINCT p.key), NULL), '{}') AS permissions
	FROM membership_roles mr
	LEFT JOIN roles r ON r.id = mr.role_id
	LEFT JOIN role_permissions rp ON rp.role_id = r.id
	LEFT JOIN permissions p ON p.id = rp.permission_id
	GROUP BY mr.tenant_id, r.id, r.name
	ORDER BY mr.tenant_id, r.name NULLS LAST
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles for user: %w", err)
	}
	defer rows.Close()

	var results []*MembershipRolePermissions
	for rows.Next() {
		var (
			tenantID    uuid.UUID
			roleID      pgtype.UUID
			roleName    *string
			permissions []string
		)

		if err := rows.Scan(&tenantID, &roleID, &roleName, &permissions); err != nil {
			return nil, fmt.Errorf("failed to scan membership role row: %w", err)
		}

		var roleIDPtr *uuid.UUID
		if roleID.Valid {
			id := uuid.UUID(roleID.Bytes)
			roleIDPtr = &id
		}

		results = append(results, &MembershipRolePermissions{
			TenantID:    tenantID,
			RoleID:      roleIDPtr,
			RoleName:    roleName,
			Permissions: permissions,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate membership roles: %w", err)
	}

	return results, nil
}
