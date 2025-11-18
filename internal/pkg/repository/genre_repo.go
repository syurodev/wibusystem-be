package repository

import (
	"context"
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

// GetByID lấy genre từ database theo ID
func (r *genreRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Genre, error) {
	query := `
		SELECT id, name, slug, description, parent_id,
		       display_order, is_active, created_at, updated_at
		FROM catalog.genres
		WHERE id = $1
	`

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
	query := `
		SELECT id, name, slug, description, parent_id,
		       display_order, is_active, created_at, updated_at
		FROM catalog.genres
		WHERE slug = $1
	`

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
	query := `
		SELECT id, name, slug, description, parent_id,
		       display_order, is_active, created_at, updated_at
		FROM catalog.genres
	`

	if activeOnly {
		query += " WHERE is_active = true"
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

// GetByParentID lấy các genre con theo parent ID
func (r *genreRepository) GetByParentID(ctx context.Context, parentID uuid.UUID) ([]*domain.Genre, error) {
	query := `
		SELECT id, name, slug, description, parent_id,
		       display_order, is_active, created_at, updated_at
		FROM catalog.genres
		WHERE parent_id = $1
		ORDER BY display_order ASC, name ASC
	`

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
	query := `
		SELECT id, name, slug, description, parent_id,
		       display_order, is_active, created_at, updated_at
		FROM catalog.genres
		WHERE parent_id IS NULL
		ORDER BY display_order ASC, name ASC
	`

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
			id, name, slug, description, parent_id, display_order, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.pool.Exec(ctx, query,
		genre.ID,
		genre.Name,
		genre.Slug,
		genre.Description,
		genre.ParentID,
		genre.DisplayOrder,
		genre.IsActive,
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
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, query,
		genre.ID,
		genre.Name,
		genre.Slug,
		genre.Description,
		genre.ParentID,
		genre.DisplayOrder,
		genre.IsActive,
	)

	return err
}

// Delete xóa genre
func (r *genreRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM catalog.genres WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// GetNovelGenres lấy danh sách genres của một novel
func (r *genreRepository) GetNovelGenres(ctx context.Context, novelID uuid.UUID) ([]*domain.Genre, error) {
	query := `
		SELECT g.id, g.name, g.slug, g.description, g.parent_id,
		       g.display_order, g.is_active, g.created_at, g.updated_at
		FROM catalog.genres g
		INNER JOIN catalog.novel_genres ng ON g.id = ng.genre_id
		WHERE ng.novel_id = $1
		ORDER BY ng.display_order ASC, g.name ASC
	`

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
func (r *genreRepository) AddNovelGenre(ctx context.Context, novelID, genreID uuid.UUID, displayOrder int) error {
	query := `
		INSERT INTO catalog.novel_genres (novel_id, genre_id, display_order)
		VALUES ($1, $2, $3)
		ON CONFLICT (novel_id, genre_id) DO UPDATE
		SET display_order = EXCLUDED.display_order
	`

	_, err := r.pool.Exec(ctx, query, novelID, genreID, displayOrder)
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
func (r *genreRepository) UpdateNovelGenres(ctx context.Context, novelID uuid.UUID, genreIDs []uuid.UUID) error {
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
				"INSERT INTO catalog.novel_genres (novel_id, genre_id, display_order) VALUES ($1, $2, $3)",
				novelID, genreID, i,
			)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit(ctx)
}
