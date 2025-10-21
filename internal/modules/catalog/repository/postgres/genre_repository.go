package postgres

import (
	"context"
	"fmt"
	"wibusystem/internal/modules/catalog/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GenreRepositoryPG is a PostgreSQL implementation of GenreRepository.
type GenreRepositoryPG struct {
	DB *pgxpool.Pool
}

// NewGenreRepositoryPG creates a new instance of GenreRepositoryPG.
func NewGenreRepositoryPG(db *pgxpool.Pool) *GenreRepositoryPG {
	return &GenreRepositoryPG{DB: db}
}

// Create inserts a new genre into the database.
func (r *GenreRepositoryPG) Create(ctx context.Context, genre *domain.Genre) error {
	query := `
		INSERT INTO genres (id, name, description, slug, parent_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.DB.Exec(ctx, query,
		genre.ID, genre.Name, genre.Description, genre.Slug, genre.ParentID, genre.CreatedAt, genre.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("postgres.GenreRepositoryPG.Create: %w", err)
	}

	return nil
}

// GetByID retrieves a genre by its ID.
func (r *GenreRepositoryPG) GetByID(ctx context.Context, id uuid.UUID) (*domain.Genre, error) {
	query := `SELECT id, name, description, slug, parent_id, novel_count, created_at, updated_at FROM genres WHERE id = $1`
	genre := &domain.Genre{}
	row := r.DB.QueryRow(ctx, query, id)

	err := row.Scan(
		&genre.ID, &genre.Name, &genre.Description, &genre.Slug, &genre.ParentID,
		&genre.NovelCount, &genre.CreatedAt, &genre.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("genre with id %s not found", id)
		}
		return nil, fmt.Errorf("postgres.GenreRepositoryPG.GetByID: %w", err)
	}

	return genre, nil
}

// Update modifies an existing genre in the database.
func (r *GenreRepositoryPG) Update(ctx context.Context, genre *domain.Genre) error {
	query := `
		UPDATE genres SET
			name = $2, description = $3, slug = $4, parent_id = $5, novel_count = $6, updated_at = $7
		WHERE id = $1`

	_, err := r.DB.Exec(ctx, query,
		genre.ID, genre.Name, genre.Description, genre.Slug, genre.ParentID,
		genre.NovelCount, genre.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("postgres.GenreRepositoryPG.Update: %w", err)
	}

	return nil
}

// Delete removes a genre from the database.
func (r *GenreRepositoryPG) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM genres WHERE id = $1`
	_, err := r.DB.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("postgres.GenreRepositoryPG.Delete: %w", err)
	}
	return nil
}

// List retrieves a paginated list of genres.
func (r *GenreRepositoryPG) List(ctx context.Context, limit, offset int) ([]*domain.Genre, int64, error) {
	query := `SELECT id, name, description, slug, parent_id, novel_count, created_at, updated_at FROM genres ORDER BY name ASC LIMIT $1 OFFSET $2`
	rows, err := r.DB.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("postgres.GenreRepositoryPG.List: %w", err)
	}
	defer rows.Close()

	genres := []*domain.Genre{}
	for rows.Next() {
		genre := &domain.Genre{}
		err := rows.Scan(
			&genre.ID, &genre.Name, &genre.Description, &genre.Slug, &genre.ParentID,
			&genre.NovelCount, &genre.CreatedAt, &genre.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("postgres.GenreRepositoryPG.List: failed to scan genre: %w", err)
		}
		genres = append(genres, genre)
	}

	countQuery := `SELECT COUNT(*) FROM genres`
	var total int64
	err = r.DB.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("postgres.GenreRepositoryPG.List: failed to count genres: %w", err)
	}

	return genres, total, nil
}
