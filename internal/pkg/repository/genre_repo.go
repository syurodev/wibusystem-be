package repository

import (
	"context"
	"fmt"
	"system/internal/domain"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// genreRepository triển khai GenreRepository sử dụng pgx
type genreRepository struct {
	pool *pgxpool.Pool
}

// NewGenreRepository tạo một instance mới của genreRepository
func NewGenreRepository(pool *pgxpool.Pool) domain.GenreRepository {
	return &genreRepository{pool: pool}
}

const genreColumns = `
	id, name, slug, description, parent_id,
	display_order, is_active, novel_count, active_readers, total_views,
	created_at, updated_at
`

// GetByID lấy genre từ database theo ID
func (r *genreRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Genre, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM catalog.genres
		WHERE id = $1 AND deleted_at IS NULL
	`, genreColumns)

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}

	genre, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Genre])
	if err != nil {
		return nil, err
	}

	return &genre, nil
}

// GetBySlug lấy genre từ database theo slug
func (r *genreRepository) GetBySlug(ctx context.Context, slug string) (*domain.Genre, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM catalog.genres
		WHERE slug = $1 AND deleted_at IS NULL
	`, genreColumns)

	rows, err := r.pool.Query(ctx, query, slug)
	if err != nil {
		return nil, err
	}

	genre, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Genre])
	if err != nil {
		return nil, err
	}

	return &genre, nil
}

// GetAll lấy tất cả genres
func (r *genreRepository) GetAll(ctx context.Context, activeOnly bool) ([]*domain.Genre, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM catalog.genres
		WHERE deleted_at IS NULL
	`, genreColumns)

	if activeOnly {
		query += " AND is_active = true"
	}

	query += " ORDER BY display_order ASC, name ASC"

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	genres, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Genre])
	if err != nil {
		return nil, err
	}

	return genres, nil
}

// List lấy danh sách genres với pagination, search và sort
func (r *genreRepository) List(ctx context.Context, offset, limit int, search, sortBy, sortOrder string, activeOnly bool) ([]*domain.Genre, int, error) {
	// Build WHERE clause
	var whereClauses []string
	var args []interface{}
	argIdx := 1

	whereClauses = append(whereClauses, "deleted_at IS NULL")

	if activeOnly {
		whereClauses = append(whereClauses, "is_active = true")
	}

	// Search by name
	if search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE $%d", argIdx))
		args = append(args, "%"+search+"%")
		argIdx++
	}

	whereClause := " WHERE " + fmt.Sprintf("%s", whereClauses[0])
	for i := 1; i < len(whereClauses); i++ {
		whereClause += " AND " + whereClauses[i]
	}

	// Query to get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM catalog.genres%s", whereClause)
	var totalCount int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, err
	}

	// Build ORDER BY clause
	orderClause := "display_order ASC, name ASC" // Default sort
	if sortBy != "" {
		orderField := ""
		switch sortBy {
		case "name":
			orderField = "name"
		case "views":
			orderField = "total_views"
		case "series":
			orderField = "novel_count"
		case "created":
			orderField = "created_at"
		case "updated":
			orderField = "updated_at"
		default:
			orderField = "display_order, name"
		}

		if orderField != "" {
			direction := "ASC"
			if sortOrder == "desc" {
				direction = "DESC"
			}
			orderClause = fmt.Sprintf("%s %s", orderField, direction)
		}
	}

	// Query to get genres with pagination
	query := fmt.Sprintf(`
		SELECT %s
		FROM catalog.genres
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, genreColumns, whereClause, orderClause, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}

	genres, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Genre])
	if err != nil {
		return nil, 0, err
	}

	return genres, totalCount, nil
}

// GetByParentID lấy các genre con theo parent ID
func (r *genreRepository) GetByParentID(ctx context.Context, parentID uuid.UUID) ([]*domain.Genre, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM catalog.genres
		WHERE parent_id = $1 AND deleted_at IS NULL
		ORDER BY display_order ASC, name ASC
	`, genreColumns)

	rows, err := r.pool.Query(ctx, query, parentID)
	if err != nil {
		return nil, err
	}

	genres, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Genre])
	if err != nil {
		return nil, err
	}

	return genres, nil
}

// GetRootGenres lấy các genre gốc (không có parent)
func (r *genreRepository) GetRootGenres(ctx context.Context) ([]*domain.Genre, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM catalog.genres
		WHERE parent_id IS NULL AND deleted_at IS NULL
		ORDER BY display_order ASC, name ASC
	`, genreColumns)

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	genres, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Genre])
	if err != nil {
		return nil, err
	}

	return genres, nil
}

// Create tạo genre mới
func (r *genreRepository) Create(ctx context.Context, genre *domain.Genre) error {
	query := `
		INSERT INTO catalog.genres (
			id, name, slug, description, parent_id, display_order,
			is_active, created_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
	`

	_, err := r.pool.Exec(ctx, query,
		genre.ID,
		genre.Name,
		genre.Slug,
		genre.Description,
		genre.ParentID,
		genre.DisplayOrder,
		genre.IsActive,
		genre.CreatedBy,
	)

	return err
}

// Update cập nhật genre
func (r *genreRepository) Update(ctx context.Context, genre *domain.Genre) error {
	query := `
		UPDATE catalog.genres
		SET name = $2,
		    slug = $3,
		    description = $4,
		    parent_id = $5,
		    display_order = $6,
		    is_active = $7,
		    updated_by = $8,
		    updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query,
		genre.ID,
		genre.Name,
		genre.Slug,
		genre.Description,
		genre.ParentID,
		genre.DisplayOrder,
		genre.IsActive,
		genre.UpdatedBy,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// Delete xóa genre (soft delete)
func (r *genreRepository) Delete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	query := `
		UPDATE catalog.genres
		SET deleted_at = NOW(),
		    deleted_by = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query, id, deletedBy)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// GetNovelGenres lấy danh sách genres của một novel
func (r *genreRepository) GetNovelGenres(ctx context.Context, novelID uuid.UUID) ([]*domain.Genre, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM catalog.genres g
		INNER JOIN catalog.novel_genres ng ON g.id = ng.genre_id
		WHERE ng.novel_id = $1 AND g.deleted_at IS NULL
		ORDER BY ng.display_order ASC, g.name ASC
	`, "g."+genreColumns)

	rows, err := r.pool.Query(ctx, query, novelID)
	if err != nil {
		return nil, err
	}

	genres, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Genre])
	if err != nil {
		return nil, err
	}

	return genres, nil
}

// AddNovelGenre thêm genre cho novel
func (r *genreRepository) AddNovelGenre(ctx context.Context, novelID, genreID, createdBy uuid.UUID, displayOrder int) error {
	query := `
		INSERT INTO catalog.novel_genres (id, novel_id, genre_id, display_order, created_by, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW())
		ON CONFLICT (novel_id, genre_id) DO UPDATE
		SET display_order = EXCLUDED.display_order
	`

	_, err := r.pool.Exec(ctx, query, novelID, genreID, displayOrder, createdBy)
	return err
}

// RemoveNovelGenre xóa genre khỏi novel
func (r *genreRepository) RemoveNovelGenre(ctx context.Context, novelID, genreID uuid.UUID) error {
	query := `
		DELETE FROM catalog.novel_genres
		WHERE novel_id = $1 AND genre_id = $2
	`

	_, err := r.pool.Exec(ctx, query, novelID, genreID)
	return err
}

// UpdateNovelGenres cập nhật toàn bộ genres của novel
func (r *genreRepository) UpdateNovelGenres(ctx context.Context, novelID uuid.UUID, genreIDs []uuid.UUID, createdBy uuid.UUID) error {
	// Start transaction
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Delete existing genres
	_, err = tx.Exec(ctx, "DELETE FROM catalog.novel_genres WHERE novel_id = $1", novelID)
	if err != nil {
		return err
	}

	// Insert new genres
	if len(genreIDs) > 0 {
		for i, genreID := range genreIDs {
			_, err = tx.Exec(ctx,
				"INSERT INTO catalog.novel_genres (id, novel_id, genre_id, display_order, created_by, created_at) VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW())",
				novelID, genreID, i, createdBy,
			)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}
