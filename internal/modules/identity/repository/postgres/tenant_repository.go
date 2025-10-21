// Package postgres provides PostgreSQL implementations of repository interfaces.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"wibusystem/internal/modules/identity/domain"
	"wibusystem/internal/modules/identity/repository"
)

// TenantRepository is the PostgreSQL implementation of repository.TenantRepository.
type TenantRepository struct {
	db     *pgxpool.Pool
	schema string
}

// NewTenantRepository creates a new PostgreSQL tenant repository.
func NewTenantRepository(db *pgxpool.Pool, schema string) *TenantRepository {
	if schema == "" {
		schema = "identity"
	}
	return &TenantRepository{
		db:     db,
		schema: schema,
	}
}

// Create creates a new tenant in the database.
func (r *TenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	if err := tenant.Validate(); err != nil {
		return repository.WrapRepositoryError(err, "invalid tenant data", "INVALID_TENANT_DATA")
	}

	query := fmt.Sprintf(`
		INSERT INTO %s.tenants (
			id, name, slug, description, logo_url, status, owner_id,
			settings, metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		)
	`, r.schema)

	_, err := r.db.Exec(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Slug,
		tenant.Description,
		tenant.LogoURL,
		tenant.Status,
		tenant.OwnerID,
		tenant.Settings,
		tenant.Metadata,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return repository.ErrTenantAlreadyExists
		}
		return repository.WrapRepositoryError(err, "failed to create tenant", "CREATE_FAILED")
	}

	return nil
}

// GetByID retrieves a tenant by its ID.
func (r *TenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	query := fmt.Sprintf(`
		SELECT id, name, slug, description, logo_url, status, owner_id,
			   settings, metadata, created_at, updated_at, deleted_at
		FROM %s.tenants
		WHERE id = $1 AND deleted_at IS NULL
	`, r.schema)

	tenant := &domain.Tenant{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Slug,
		&tenant.Description,
		&tenant.LogoURL,
		&tenant.Status,
		&tenant.OwnerID,
		&tenant.Settings,
		&tenant.Metadata,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
		&tenant.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrTenantNotFound
		}
		return nil, repository.WrapRepositoryError(err, "failed to get tenant by ID", "GET_FAILED")
	}

	return tenant, nil
}

// GetBySlug retrieves a tenant by its slug.
func (r *TenantRepository) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))

	query := fmt.Sprintf(`
		SELECT id, name, slug, description, logo_url, status, owner_id,
			   settings, metadata, created_at, updated_at, deleted_at
		FROM %s.tenants
		WHERE slug = $1 AND deleted_at IS NULL
	`, r.schema)

	tenant := &domain.Tenant{}
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Slug,
		&tenant.Description,
		&tenant.LogoURL,
		&tenant.Status,
		&tenant.OwnerID,
		&tenant.Settings,
		&tenant.Metadata,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
		&tenant.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrTenantNotFound
		}
		return nil, repository.WrapRepositoryError(err, "failed to get tenant by slug", "GET_FAILED")
	}

	return tenant, nil
}

// Update updates an existing tenant's information.
func (r *TenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	if err := tenant.Validate(); err != nil {
		return repository.WrapRepositoryError(err, "invalid tenant data", "INVALID_TENANT_DATA")
	}

	tenant.UpdatedAt = time.Now().UTC()

	query := fmt.Sprintf(`
		UPDATE %s.tenants
		SET name = $2,
			slug = $3,
			description = $4,
			logo_url = $5,
			status = $6,
			owner_id = $7,
			settings = $8,
			metadata = $9,
			updated_at = $10
		WHERE id = $1 AND deleted_at IS NULL
	`, r.schema)

	result, err := r.db.Exec(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Slug,
		tenant.Description,
		tenant.LogoURL,
		tenant.Status,
		tenant.OwnerID,
		tenant.Settings,
		tenant.Metadata,
		tenant.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return repository.ErrTenantAlreadyExists
		}
		return repository.WrapRepositoryError(err, "failed to update tenant", "UPDATE_FAILED")
	}

	if result.RowsAffected() == 0 {
		return repository.ErrTenantNotFound
	}

	return nil
}

// Delete soft-deletes a tenant by its ID.
func (r *TenantRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()

	query := fmt.Sprintf(`
		UPDATE %s.tenants
		SET deleted_at = $2,
			status = $3,
			updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL
	`, r.schema)

	result, err := r.db.Exec(ctx, query, id, now, domain.TenantStatusInactive, now)
	if err != nil {
		return repository.WrapRepositoryError(err, "failed to delete tenant", "DELETE_FAILED")
	}

	if result.RowsAffected() == 0 {
		return repository.ErrTenantNotFound
	}

	return nil
}

// HardDelete permanently deletes a tenant from the database.
func (r *TenantRepository) HardDelete(ctx context.Context, id uuid.UUID) error {
	query := fmt.Sprintf(`DELETE FROM %s.tenants WHERE id = $1`, r.schema)

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return repository.WrapRepositoryError(err, "failed to hard delete tenant", "HARD_DELETE_FAILED")
	}

	if result.RowsAffected() == 0 {
		return repository.ErrTenantNotFound
	}

	return nil
}

// List retrieves a paginated list of tenants based on the provided filter.
func (r *TenantRepository) List(ctx context.Context, filter repository.TenantListFilter) ([]*domain.Tenant, int, error) {
	// Build WHERE clause
	whereClauses := []string{}
	args := []any{}
	argCount := 1

	if !filter.IncludeDeleted {
		whereClauses = append(whereClauses, "deleted_at IS NULL")
	}

	if filter.Status != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argCount))
		args = append(args, *filter.Status)
		argCount++
	}

	if filter.OwnerID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("owner_id = $%d", argCount))
		args = append(args, *filter.OwnerID)
		argCount++
	}

	if filter.NameContains != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE $%d", argCount))
		args = append(args, "%"+filter.NameContains+"%")
		argCount++
	}

	if filter.SlugContains != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("slug ILIKE $%d", argCount))
		args = append(args, "%"+filter.SlugContains+"%")
		argCount++
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s.tenants %s`, r.schema, whereClause)
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, repository.WrapRepositoryError(err, "failed to count tenants", "COUNT_FAILED")
	}

	// Build ORDER BY clause
	orderBy := "created_at DESC"
	if filter.SortBy != "" {
		order := "DESC"
		if filter.SortOrder == "asc" {
			order = "ASC"
		}
		switch filter.SortBy {
		case "created_at", "name", "slug", "status":
			orderBy = fmt.Sprintf("%s %s", filter.SortBy, order)
		}
	}

	// Build LIMIT and OFFSET
	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	// Query tenants
	query := fmt.Sprintf(`
		SELECT id, name, slug, description, logo_url, status, owner_id,
			   settings, metadata, created_at, updated_at, deleted_at
		FROM %s.tenants
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, r.schema, whereClause, orderBy, argCount, argCount+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, repository.WrapRepositoryError(err, "failed to list tenants", "LIST_FAILED")
	}
	defer rows.Close()

	tenants := []*domain.Tenant{}
	for rows.Next() {
		tenant := &domain.Tenant{}
		err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.Slug,
			&tenant.Description,
			&tenant.LogoURL,
			&tenant.Status,
			&tenant.OwnerID,
			&tenant.Settings,
			&tenant.Metadata,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
			&tenant.DeletedAt,
		)
		if err != nil {
			return nil, 0, repository.WrapRepositoryError(err, "failed to scan tenant", "SCAN_FAILED")
		}
		tenants = append(tenants, tenant)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, repository.WrapRepositoryError(err, "error iterating tenants", "ITERATION_FAILED")
	}

	return tenants, total, nil
}

// GetByOwnerID retrieves all tenants owned by a specific user.
func (r *TenantRepository) GetByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*domain.Tenant, error) {
	query := fmt.Sprintf(`
		SELECT id, name, slug, description, logo_url, status, owner_id,
			   settings, metadata, created_at, updated_at, deleted_at
		FROM %s.tenants
		WHERE owner_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, r.schema)

	rows, err := r.db.Query(ctx, query, ownerID)
	if err != nil {
		return nil, repository.WrapRepositoryError(err, "failed to get tenants by owner", "GET_FAILED")
	}
	defer rows.Close()

	tenants := []*domain.Tenant{}
	for rows.Next() {
		tenant := &domain.Tenant{}
		err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.Slug,
			&tenant.Description,
			&tenant.LogoURL,
			&tenant.Status,
			&tenant.OwnerID,
			&tenant.Settings,
			&tenant.Metadata,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
			&tenant.DeletedAt,
		)
		if err != nil {
			return nil, repository.WrapRepositoryError(err, "failed to scan tenant", "SCAN_FAILED")
		}
		tenants = append(tenants, tenant)
	}

	if err := rows.Err(); err != nil {
		return nil, repository.WrapRepositoryError(err, "error iterating tenants", "ITERATION_FAILED")
	}

	return tenants, nil
}

// GetUserTenants retrieves all tenants a user is a member of.
func (r *TenantRepository) GetUserTenants(ctx context.Context, userID uuid.UUID) ([]*domain.Tenant, error) {
	query := fmt.Sprintf(`
		SELECT DISTINCT t.id, t.name, t.slug, t.description, t.logo_url, t.status, t.owner_id,
			   t.settings, t.metadata, t.created_at, t.updated_at, t.deleted_at
		FROM %s.tenants t
		INNER JOIN %s.tenant_members tm ON t.id = tm.tenant_id
		WHERE tm.user_id = $1 AND t.deleted_at IS NULL
		ORDER BY t.created_at DESC
	`, r.schema, r.schema)

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, repository.WrapRepositoryError(err, "failed to get user tenants", "GET_FAILED")
	}
	defer rows.Close()

	tenants := []*domain.Tenant{}
	for rows.Next() {
		tenant := &domain.Tenant{}
		err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.Slug,
			&tenant.Description,
			&tenant.LogoURL,
			&tenant.Status,
			&tenant.OwnerID,
			&tenant.Settings,
			&tenant.Metadata,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
			&tenant.DeletedAt,
		)
		if err != nil {
			return nil, repository.WrapRepositoryError(err, "failed to scan tenant", "SCAN_FAILED")
		}
		tenants = append(tenants, tenant)
	}

	if err := rows.Err(); err != nil {
		return nil, repository.WrapRepositoryError(err, "error iterating tenants", "ITERATION_FAILED")
	}

	return tenants, nil
}

// ExistsBySlug checks if a tenant with the given slug exists.
func (r *TenantRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	slug = strings.ToLower(strings.TrimSpace(slug))

	query := fmt.Sprintf(`
		SELECT EXISTS(SELECT 1 FROM %s.tenants WHERE slug = $1 AND deleted_at IS NULL)
	`, r.schema)

	var exists bool
	err := r.db.QueryRow(ctx, query, slug).Scan(&exists)
	if err != nil {
		return false, repository.WrapRepositoryError(err, "failed to check slug existence", "EXISTS_CHECK_FAILED")
	}

	return exists, nil
}

// ExistsByID checks if a tenant with the given ID exists.
func (r *TenantRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	query := fmt.Sprintf(`
		SELECT EXISTS(SELECT 1 FROM %s.tenants WHERE id = $1 AND deleted_at IS NULL)
	`, r.schema)

	var exists bool
	err := r.db.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, repository.WrapRepositoryError(err, "failed to check ID existence", "EXISTS_CHECK_FAILED")
	}

	return exists, nil
}

// CountAll returns the total number of tenants.
func (r *TenantRepository) CountAll(ctx context.Context, includeDeleted bool) (int, error) {
	whereClause := ""
	if !includeDeleted {
		whereClause = "WHERE deleted_at IS NULL"
	}

	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s.tenants %s`, r.schema, whereClause)

	var count int
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, repository.WrapRepositoryError(err, "failed to count tenants", "COUNT_FAILED")
	}

	return count, nil
}

// CountByStatus returns the number of tenants with a specific status.
func (r *TenantRepository) CountByStatus(ctx context.Context, status domain.TenantStatus) (int, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) FROM %s.tenants WHERE status = $1 AND deleted_at IS NULL
	`, r.schema)

	var count int
	err := r.db.QueryRow(ctx, query, status).Scan(&count)
	if err != nil {
		return 0, repository.WrapRepositoryError(err, "failed to count tenants by status", "COUNT_FAILED")
	}

	return count, nil
}

// UpdateStatus updates a tenant's status.
func (r *TenantRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.TenantStatus) error {
	now := time.Now().UTC()

	query := fmt.Sprintf(`
		UPDATE %s.tenants
		SET status = $2, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`, r.schema)

	result, err := r.db.Exec(ctx, query, id, status, now)
	if err != nil {
		return repository.WrapRepositoryError(err, "failed to update status", "UPDATE_FAILED")
	}

	if result.RowsAffected() == 0 {
		return repository.ErrTenantNotFound
	}

	return nil
}

// UpdateOwner transfers ownership of a tenant to another user.
func (r *TenantRepository) UpdateOwner(ctx context.Context, tenantID, newOwnerID uuid.UUID) error {
	now := time.Now().UTC()

	query := fmt.Sprintf(`
		UPDATE %s.tenants
		SET owner_id = $2, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`, r.schema)

	result, err := r.db.Exec(ctx, query, tenantID, newOwnerID, now)
	if err != nil {
		return repository.WrapRepositoryError(err, "failed to update owner", "UPDATE_FAILED")
	}

	if result.RowsAffected() == 0 {
		return repository.ErrTenantNotFound
	}

	return nil
}

// Search searches for tenants by name or slug.
func (r *TenantRepository) Search(ctx context.Context, query string, limit int) ([]*domain.Tenant, error) {
	if limit <= 0 {
		limit = 10
	}

	searchQuery := fmt.Sprintf(`
		SELECT id, name, slug, description, logo_url, status, owner_id,
			   settings, metadata, created_at, updated_at, deleted_at
		FROM %s.tenants
		WHERE deleted_at IS NULL
		  AND (name ILIKE $1 OR slug ILIKE $1)
		ORDER BY
			CASE WHEN slug = $2 THEN 0 ELSE 1 END,
			created_at DESC
		LIMIT $3
	`, r.schema)

	pattern := "%" + query + "%"
	rows, err := r.db.Query(ctx, searchQuery, pattern, query, limit)
	if err != nil {
		return nil, repository.WrapRepositoryError(err, "failed to search tenants", "SEARCH_FAILED")
	}
	defer rows.Close()

	tenants := []*domain.Tenant{}
	for rows.Next() {
		tenant := &domain.Tenant{}
		err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.Slug,
			&tenant.Description,
			&tenant.LogoURL,
			&tenant.Status,
			&tenant.OwnerID,
			&tenant.Settings,
			&tenant.Metadata,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
			&tenant.DeletedAt,
		)
		if err != nil {
			return nil, repository.WrapRepositoryError(err, "failed to scan tenant", "SCAN_FAILED")
		}
		tenants = append(tenants, tenant)
	}

	if err := rows.Err(); err != nil {
		return nil, repository.WrapRepositoryError(err, "error iterating tenants", "ITERATION_FAILED")
	}

	return tenants, nil
}

// GetByIDs retrieves multiple tenants by their IDs.
func (r *TenantRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*domain.Tenant, error) {
	if len(ids) == 0 {
		return []*domain.Tenant{}, nil
	}

	query := fmt.Sprintf(`
		SELECT id, name, slug, description, logo_url, status, owner_id,
			   settings, metadata, created_at, updated_at, deleted_at
		FROM %s.tenants
		WHERE id = ANY($1) AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, r.schema)

	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, repository.WrapRepositoryError(err, "failed to get tenants by IDs", "GET_FAILED")
	}
	defer rows.Close()

	tenants := []*domain.Tenant{}
	for rows.Next() {
		tenant := &domain.Tenant{}
		err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.Slug,
			&tenant.Description,
			&tenant.LogoURL,
			&tenant.Status,
			&tenant.OwnerID,
			&tenant.Settings,
			&tenant.Metadata,
			&tenant.CreatedAt,
			&tenant.UpdatedAt,
			&tenant.DeletedAt,
		)
		if err != nil {
			return nil, repository.WrapRepositoryError(err, "failed to scan tenant", "SCAN_FAILED")
		}
		tenants = append(tenants, tenant)
	}

	if err := rows.Err(); err != nil {
		return nil, repository.WrapRepositoryError(err, "error iterating tenants", "ITERATION_FAILED")
	}

	return tenants, nil
}

// TenantMemberRepository is the PostgreSQL implementation of repository.TenantMemberRepository.
type TenantMemberRepository struct {
	db     *pgxpool.Pool
	schema string
}

// NewTenantMemberRepository creates a new PostgreSQL tenant member repository.
func NewTenantMemberRepository(db *pgxpool.Pool, schema string) *TenantMemberRepository {
	if schema == "" {
		schema = "identity"
	}
	return &TenantMemberRepository{
		db:     db,
		schema: schema,
	}
}

// Create creates a new tenant member in the database.
func (r *TenantMemberRepository) Create(ctx context.Context, member *domain.TenantMember) error {
	query := fmt.Sprintf(`
		INSERT INTO %s.tenant_members (
			id, tenant_id, user_id, role, permissions, invited_by,
			joined_at, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`, r.schema)

	_, err := r.db.Exec(ctx, query,
		member.ID,
		member.TenantID,
		member.UserID,
		member.Role,
		member.Permissions,
		member.InvitedBy,
		member.JoinedAt,
		member.CreatedAt,
		member.UpdatedAt,
	)

	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			return repository.ErrTenantMemberAlreadyExists
		}
		return repository.WrapRepositoryError(err, "failed to create tenant member", "CREATE_FAILED")
	}

	return nil
}

// GetByID retrieves a tenant member by its ID.
func (r *TenantMemberRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.TenantMember, error) {
	query := fmt.Sprintf(`
		SELECT id, tenant_id, user_id, role, permissions, invited_by,
			   joined_at, created_at, updated_at
		FROM %s.tenant_members
		WHERE id = $1
	`, r.schema)

	member := &domain.TenantMember{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&member.ID,
		&member.TenantID,
		&member.UserID,
		&member.Role,
		&member.Permissions,
		&member.InvitedBy,
		&member.JoinedAt,
		&member.CreatedAt,
		&member.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrTenantMemberNotFound
		}
		return nil, repository.WrapRepositoryError(err, "failed to get tenant member", "GET_FAILED")
	}

	return member, nil
}

// GetByTenantAndUser retrieves a tenant member by tenant ID and user ID.
func (r *TenantMemberRepository) GetByTenantAndUser(ctx context.Context, tenantID, userID uuid.UUID) (*domain.TenantMember, error) {
	query := fmt.Sprintf(`
		SELECT id, tenant_id, user_id, role, permissions, invited_by,
			   joined_at, created_at, updated_at
		FROM %s.tenant_members
		WHERE tenant_id = $1 AND user_id = $2
	`, r.schema)

	member := &domain.TenantMember{}
	err := r.db.QueryRow(ctx, query, tenantID, userID).Scan(
		&member.ID,
		&member.TenantID,
		&member.UserID,
		&member.Role,
		&member.Permissions,
		&member.InvitedBy,
		&member.JoinedAt,
		&member.CreatedAt,
		&member.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrTenantMemberNotFound
		}
		return nil, repository.WrapRepositoryError(err, "failed to get tenant member", "GET_FAILED")
	}

	return member, nil
}

// Update updates an existing tenant member's information.
func (r *TenantMemberRepository) Update(ctx context.Context, member *domain.TenantMember) error {
	member.UpdatedAt = time.Now().UTC()

	query := fmt.Sprintf(`
		UPDATE %s.tenant_members
		SET role = $2,
			permissions = $3,
			updated_at = $4
		WHERE id = $1
	`, r.schema)

	result, err := r.db.Exec(ctx, query,
		member.ID,
		member.Role,
		member.Permissions,
		member.UpdatedAt,
	)

	if err != nil {
		return repository.WrapRepositoryError(err, "failed to update tenant member", "UPDATE_FAILED")
	}

	if result.RowsAffected() == 0 {
		return repository.ErrTenantMemberNotFound
	}

	return nil
}

// Delete removes a tenant member from the database.
func (r *TenantMemberRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := fmt.Sprintf(`DELETE FROM %s.tenant_members WHERE id = $1`, r.schema)

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return repository.WrapRepositoryError(err, "failed to delete tenant member", "DELETE_FAILED")
	}

	if result.RowsAffected() == 0 {
		return repository.ErrTenantMemberNotFound
	}

	return nil
}

// DeleteByTenantAndUser removes a tenant member by tenant ID and user ID.
func (r *TenantMemberRepository) DeleteByTenantAndUser(ctx context.Context, tenantID, userID uuid.UUID) error {
	// Check if user is owner
	member, err := r.GetByTenantAndUser(ctx, tenantID, userID)
	if err != nil {
		return err
	}

	if member.Role == domain.MemberRoleOwner {
		return repository.ErrCannotDeleteOwner
	}

	query := fmt.Sprintf(`
		DELETE FROM %s.tenant_members
		WHERE tenant_id = $1 AND user_id = $2
	`, r.schema)

	result, err := r.db.Exec(ctx, query, tenantID, userID)
	if err != nil {
		return repository.WrapRepositoryError(err, "failed to delete tenant member", "DELETE_FAILED")
	}

	if result.RowsAffected() == 0 {
		return repository.ErrTenantMemberNotFound
	}

	return nil
}

// ListByTenant retrieves all members of a specific tenant.
func (r *TenantMemberRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.TenantMember, error) {
	query := fmt.Sprintf(`
		SELECT id, tenant_i
d, user_id, role, permissions, invited_by,
			   joined_at, created_at, updated_at
		FROM %s.tenant_members
		WHERE tenant_id = $1
		ORDER BY joined_at DESC
	`, r.schema)

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, repository.WrapRepositoryError(err, "failed to list tenant members", "LIST_FAILED")
	}
	defer rows.Close()

	members := []*domain.TenantMember{}
	for rows.Next() {
		member := &domain.TenantMember{}
		err := rows.Scan(
			&member.ID,
			&member.TenantID,
			&member.UserID,
			&member.Role,
			&member.Permissions,
			&member.InvitedBy,
			&member.JoinedAt,
			&member.CreatedAt,
			&member.UpdatedAt,
		)
		if err != nil {
			return nil, repository.WrapRepositoryError(err, "failed to scan member", "SCAN_FAILED")
		}
		members = append(members, member)
	}

	return members, rows.Err()
}

// ListByTenantPaginated retrieves members with pagination.
func (r *TenantMemberRepository) ListByTenantPaginated(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.TenantMember, int, error) {
	// Count total
	var total int
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM %s.tenant_members WHERE tenant_id = $1`, r.schema)
	if err := r.db.QueryRow(ctx, countQuery, tenantID).Scan(&total); err != nil {
		return nil, 0, repository.WrapRepositoryError(err, "failed to count members", "COUNT_FAILED")
	}

	// Get members
	query := fmt.Sprintf(`
		SELECT id, tenant_id, user_id, role, permissions, invited_by,
			   joined_at, created_at, updated_at
		FROM %s.tenant_members
		WHERE tenant_id = $1
		ORDER BY joined_at DESC
		LIMIT $2 OFFSET $3
	`, r.schema)

	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, 0, repository.WrapRepositoryError(err, "failed to list members", "LIST_FAILED")
	}
	defer rows.Close()

	members := []*domain.TenantMember{}
	for rows.Next() {
		member := &domain.TenantMember{}
		if err := rows.Scan(&member.ID, &member.TenantID, &member.UserID, &member.Role, &member.Permissions, &member.InvitedBy, &member.JoinedAt, &member.CreatedAt, &member.UpdatedAt); err != nil {
			return nil, 0, repository.WrapRepositoryError(err, "failed to scan member", "SCAN_FAILED")
		}
		members = append(members, member)
	}

	return members, total, rows.Err()
}

// ListByUser retrieves all tenant memberships for a user.
func (r *TenantMemberRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.TenantMember, error) {
	query := fmt.Sprintf(`
		SELECT id, tenant_id, user_id, role, permissions, invited_by,
			   joined_at, created_at, updated_at
		FROM %s.tenant_members
		WHERE user_id = $1
		ORDER BY joined_at DESC
	`, r.schema)

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, repository.WrapRepositoryError(err, "failed to list user memberships", "LIST_FAILED")
	}
	defer rows.Close()

	members := []*domain.TenantMember{}
	for rows.Next() {
		member := &domain.TenantMember{}
		if err := rows.Scan(&member.ID, &member.TenantID, &member.UserID, &member.Role, &member.Permissions, &member.InvitedBy, &member.JoinedAt, &member.CreatedAt, &member.UpdatedAt); err != nil {
			return nil, repository.WrapRepositoryError(err, "failed to scan member", "SCAN_FAILED")
		}
		members = append(members, member)
	}

	return members, rows.Err()
}

// ListByTenantAndRole retrieves members with a specific role.
func (r *TenantMemberRepository) ListByTenantAndRole(ctx context.Context, tenantID uuid.UUID, role domain.MemberRole) ([]*domain.TenantMember, error) {
	query := fmt.Sprintf(`
		SELECT id, tenant_id, user_id, role, permissions, invited_by,
			   joined_at, created_at, updated_at
		FROM %s.tenant_members
		WHERE tenant_id = $1 AND role = $2
		ORDER BY joined_at DESC
	`, r.schema)

	rows, err := r.db.Query(ctx, query, tenantID, role)
	if err != nil {
		return nil, repository.WrapRepositoryError(err, "failed to list members by role", "LIST_FAILED")
	}
	defer rows.Close()

	members := []*domain.TenantMember{}
	for rows.Next() {
		member := &domain.TenantMember{}
		if err := rows.Scan(&member.ID, &member.TenantID, &member.UserID, &member.Role, &member.Permissions, &member.InvitedBy, &member.JoinedAt, &member.CreatedAt, &member.UpdatedAt); err != nil {
			return nil, repository.WrapRepositoryError(err, "failed to scan member", "SCAN_FAILED")
		}
		members = append(members, member)
	}

	return members, rows.Err()
}

// CountByTenant returns the number of members in a tenant.
func (r *TenantMemberRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int, error) {
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s.tenant_members WHERE tenant_id = $1`, r.schema)
	var count int
	if err := r.db.QueryRow(ctx, query, tenantID).Scan(&count); err != nil {
		return 0, repository.WrapRepositoryError(err, "failed to count members", "COUNT_FAILED")
	}
	return count, nil
}

// CountByUser returns the number of tenants a user is member of.
func (r *TenantMemberRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int, error) {
	query := fmt.Sprintf(`SELECT COUNT(*) FROM %s.tenant_members WHERE user_id = $1`, r.schema)
	var count int
	if err := r.db.QueryRow(ctx, query, userID).Scan(&count); err != nil {
		return 0, repository.WrapRepositoryError(err, "failed to count memberships", "COUNT_FAILED")
	}
	return count, nil
}

// ExistsByTenantAndUser checks if a user is a member of a tenant.
func (r *TenantMemberRepository) ExistsByTenantAndUser(ctx context.Context, tenantID, userID uuid.UUID) (bool, error) {
	query := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s.tenant_members WHERE tenant_id = $1 AND user_id = $2)`, r.schema)
	var exists bool
	if err := r.db.QueryRow(ctx, query, tenantID, userID).Scan(&exists); err != nil {
		return false, repository.WrapRepositoryError(err, "failed to check membership", "EXISTS_CHECK_FAILED")
	}
	return exists, nil
}

// UpdateRole updates a member's role.
func (r *TenantMemberRepository) UpdateRole(ctx context.Context, id uuid.UUID, role domain.MemberRole) error {
	query := fmt.Sprintf(`UPDATE %s.tenant_members SET role = $2, updated_at = $3 WHERE id = $1`, r.schema)
	result, err := r.db.Exec(ctx, query, id, role, time.Now().UTC())
	if err != nil {
		return repository.WrapRepositoryError(err, "failed to update role", "UPDATE_FAILED")
	}
	if result.RowsAffected() == 0 {
		return repository.ErrTenantMemberNotFound
	}
	return nil
}

// UpdatePermissions updates a member's permissions.
func (r *TenantMemberRepository) UpdatePermissions(ctx context.Context, id uuid.UUID, permissions []string) error {
	query := fmt.Sprintf(`UPDATE %s.tenant_members SET permissions = $2, updated_at = $3 WHERE id = $1`, r.schema)
	result, err := r.db.Exec(ctx, query, id, permissions, time.Now().UTC())
	if err != nil {
		return repository.WrapRepositoryError(err, "failed to update permissions", "UPDATE_FAILED")
	}
	if result.RowsAffected() == 0 {
		return repository.ErrTenantMemberNotFound
	}
	return nil
}

// GetOwner retrieves the owner of a tenant.
func (r *TenantMemberRepository) GetOwner(ctx context.Context, tenantID uuid.UUID) (*domain.TenantMember, error) {
	query := fmt.Sprintf(`
		SELECT id, tenant_id, user_id, role, permissions, invited_by,
			   joined_at, created_at, updated_at
		FROM %s.tenant_members
		WHERE tenant_id = $1 AND role = $2
	`, r.schema)

	member := &domain.TenantMember{}
	err := r.db.QueryRow(ctx, query, tenantID, domain.MemberRoleOwner).Scan(
		&member.ID, &member.TenantID, &member.UserID, &member.Role,
		&member.Permissions, &member.InvitedBy, &member.JoinedAt,
		&member.CreatedAt, &member.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrTenantMemberNotFound
		}
		return nil, repository.WrapRepositoryError(err, "failed to get owner", "GET_FAILED")
	}

	return member, nil
}

// IsOwner checks if a user is the owner.
func (r *TenantMemberRepository) IsOwner(ctx context.Context, tenantID, userID uuid.UUID) (bool, error) {
	query := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s.tenant_members WHERE tenant_id = $1 AND user_id = $2 AND role = $3)`, r.schema)
	var isOwner bool
	if err := r.db.QueryRow(ctx, query, tenantID, userID, domain.MemberRoleOwner).Scan(&isOwner); err != nil {
		return false, repository.WrapRepositoryError(err, "failed to check owner", "CHECK_FAILED")
	}
	return isOwner, nil
}

// HasRole checks if a user has a specific role.
func (r *TenantMemberRepository) HasRole(ctx context.Context, tenantID, userID uuid.UUID, role domain.MemberRole) (bool, error) {
	query := fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s.tenant_members WHERE tenant_id = $1 AND user_id = $2 AND role = $3)`, r.schema)
	var hasRole bool
	if err := r.db.QueryRow(ctx, query, tenantID, userID, role).Scan(&hasRole); err != nil {
		return false, repository.WrapRepositoryError(err, "failed to check role", "CHECK_FAILED")
	}
	return hasRole, nil
}

// GetUserRole retrieves a user's role in a tenant.
func (r *TenantMemberRepository) GetUserRole(ctx context.Context, tenantID, userID uuid.UUID) (domain.MemberRole, error) {
	query := fmt.Sprintf(`SELECT role FROM %s.tenant_members WHERE tenant_id = $1 AND user_id = $2`, r.schema)
	var role domain.MemberRole
	if err := r.db.QueryRow(ctx, query, tenantID, userID).Scan(&role); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", repository.ErrTenantMemberNotFound
		}
		return "", repository.WrapRepositoryError(err, "failed to get role", "GET_FAILED")
	}
	return role, nil
}

// DeleteAllByTenant removes all members from a tenant.
func (r *TenantMemberRepository) DeleteAllByTenant(ctx context.Context, tenantID uuid.UUID) error {
	query := fmt.Sprintf(`DELETE FROM %s.tenant_members WHERE tenant_id = $1`, r.schema)
	_, err := r.db.Exec(ctx, query, tenantID)
	return err
}

// DeleteAllByUser removes all memberships for a user.
func (r *TenantMemberRepository) DeleteAllByUser(ctx context.Context, userID uuid.UUID) error {
	query := fmt.Sprintf(`DELETE FROM %s.tenant_members WHERE user_id = $1`, r.schema)
	_, err := r.db.Exec(ctx, query, userID)
	return err
}
