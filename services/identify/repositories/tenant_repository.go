package repositories

import (
	"context"
	"fmt"

	m "wibusystem/pkg/common/model"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TenantRepository defines operations to manage tenants and their members.
type TenantRepository interface {
	Create(ctx context.Context, tenant *m.Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*m.Tenant, error)
	GetByName(ctx context.Context, name string) (*m.Tenant, error)
	List(ctx context.Context, limit, offset int) ([]*m.Tenant, int64, error)
	Update(ctx context.Context, tenant *m.Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*m.Tenant, error)
}

type tenantRepository struct {
	pool *pgxpool.Pool
}

// NewTenantRepository creates a Postgres-backed TenantRepository.
func NewTenantRepository(pool *pgxpool.Pool) TenantRepository {
	return &tenantRepository{pool: pool}
}

// Create inserts a new tenant.
func (r *tenantRepository) Create(ctx context.Context, tenant *m.Tenant) error {
	query := `
		INSERT INTO tenants (name)
		VALUES ($1)
		RETURNING id, created_at
	`

	err := r.pool.QueryRow(ctx, query, tenant.Name).Scan(
		&tenant.ID,
		&tenant.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	return nil
}

// GetByID returns a tenant by ID.
func (r *tenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*m.Tenant, error) {
	query := `
		SELECT id, name, created_at
		FROM tenants
		WHERE id = $1
	`

	tenant := &m.Tenant{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get tenant by ID: %w", err)
	}

	return tenant, nil
}

// GetByName returns a tenant by its unique name.
func (r *tenantRepository) GetByName(ctx context.Context, name string) (*m.Tenant, error) {
	query := `
		SELECT id, name, created_at
		FROM tenants
		WHERE name = $1
	`

	tenant := &m.Tenant{}
	err := r.pool.QueryRow(ctx, query, name).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get tenant by name: %w", err)
	}

	return tenant, nil
}

// List returns paginated tenants and total count.
func (r *tenantRepository) List(ctx context.Context, limit, offset int) ([]*m.Tenant, int64, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM tenants`
	var total int64
	err := r.pool.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get tenant count: %w", err)
	}

	// Get tenants with pagination
	query := `
		SELECT id, name, created_at
		FROM tenants
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*m.Tenant
	for rows.Next() {
		tenant := &m.Tenant{}
		err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan tenant: %w", err)
		}

		tenants = append(tenants, tenant)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("failed to iterate tenants: %w", err)
	}

	return tenants, total, nil
}

// Update modifies tenant fields (e.g., name).
func (r *tenantRepository) Update(ctx context.Context, tenant *m.Tenant) error {
	query := `
		UPDATE tenants
		SET name = $2
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, tenant.ID, tenant.Name)
	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("tenant with ID %d not found", tenant.ID)
	}

	return nil
}

// Delete removes a tenant by ID.
func (r *tenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tenants WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("tenant with ID %d not found", id)
	}

	return nil
}

// GetByUserID returns active tenants for a user via membership join.
func (r *tenantRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*m.Tenant, error) {
	query := `
		SELECT t.id, t.name, t.created_at
		FROM tenants t
		JOIN memberships m ON t.id = m.tenant_id
		WHERE m.user_id = $1 AND m.status = 'active'
		ORDER BY t.name
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenants by user ID: %w", err)
	}
	defer rows.Close()

	var tenants []*m.Tenant
	for rows.Next() {
		tenant := &m.Tenant{}
		err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tenant: %w", err)
		}

		tenants = append(tenants, tenant)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate tenants: %w", err)
	}

	return tenants, nil
}
