package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	m "wibusystem/pkg/common/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GenreRepository defines CRUD and listing operations for genres
type GenreRepository interface {
	Create(ctx context.Context, genre *m.Genre) error
	GetByID(ctx context.Context, id uuid.UUID) (*m.Genre, error)
	GetByName(ctx context.Context, name string) (*m.Genre, error)
	Update(ctx context.Context, genre *m.Genre) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int, search string) ([]*m.Genre, int64, error)
}

type genreRepository struct {
	pool *pgxpool.Pool
}

// NewGenreRepository creates a Postgres-backed GenreRepository
func NewGenreRepository(pool *pgxpool.Pool) GenreRepository {
	return &genreRepository{pool: pool}
}

// Create inserts a new genre and returns the assigned ID and timestamps
func (r *genreRepository) Create(ctx context.Context, genre *m.Genre) error {
	query := `
		INSERT INTO genre (name)
		VALUES ($1)
		RETURNING id, created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query, genre.Name).Scan(
		&genre.ID,
		&genre.CreatedAt,
		&genre.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create genre: %w", err)
	}

	return nil
}

// GetByID retrieves a genre by its ID
func (r *genreRepository) GetByID(ctx context.Context, id uuid.UUID) (*m.Genre, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM genre
		WHERE id = $1
	`

	var genre m.Genre
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&genre.ID,
		&genre.Name,
		&genre.CreatedAt,
		&genre.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get genre by ID: %w", err)
	}

	return &genre, nil
}

// GetByName retrieves a genre by its name
func (r *genreRepository) GetByName(ctx context.Context, name string) (*m.Genre, error) {
	query := `
		SELECT id, name, created_at, updated_at
		FROM genre
		WHERE LOWER(name) = LOWER($1)
	`

	var genre m.Genre
	err := r.pool.QueryRow(ctx, query, name).Scan(
		&genre.ID,
		&genre.Name,
		&genre.CreatedAt,
		&genre.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get genre by name: %w", err)
	}

	return &genre, nil
}

// Update modifies an existing genre
func (r *genreRepository) Update(ctx context.Context, genre *m.Genre) error {
	query := `
		UPDATE genre
		SET name = $2, updated_at = $3
		WHERE id = $1
		RETURNING updated_at
	`

	genre.UpdatedAt = time.Now()

	err := r.pool.QueryRow(ctx, query, genre.ID, genre.Name, genre.UpdatedAt).Scan(
		&genre.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update genre: %w", err)
	}

	return nil
}

// Delete removes a genre by ID
func (r *genreRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM genre WHERE id = $1`

	commandTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete genre: %w", err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("genre not found")
	}

	return nil
}

// List retrieves paginated genres with optional search
func (r *genreRepository) List(ctx context.Context, limit, offset int, search string) ([]*m.Genre, int64, error) {
	var genres []*m.Genre
	var total int64

	// Base query with search condition
	whereClause := ""
	args := []interface{}{limit, offset}

	if search != "" {
		whereClause = "WHERE LOWER(name) LIKE LOWER($3)"
		args = append(args, "%"+strings.ToLower(search)+"%")
	}

	// Count query
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM genre %s", whereClause)
	var countArgs []interface{}
	if search != "" {
		countArgs = []interface{}{"%"+strings.ToLower(search)+"%"}
	}

	err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count genres: %w", err)
	}

	// Data query
	query := fmt.Sprintf(`
		SELECT id, name, created_at, updated_at
		FROM genre
		%s
		ORDER BY name ASC
		LIMIT $1 OFFSET $2
	`, whereClause)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list genres: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var genre m.Genre
		err := rows.Scan(
			&genre.ID,
			&genre.Name,
			&genre.CreatedAt,
			&genre.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan genre: %w", err)
		}
		genres = append(genres, &genre)
	}

	if rows.Err() != nil {
		return nil, 0, fmt.Errorf("failed to iterate genres: %w", rows.Err())
	}

	return genres, total, nil
}