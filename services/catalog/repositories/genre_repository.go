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

	// Query methods for novel relations
	GetGenresByNovelID(ctx context.Context, novelID uuid.UUID) ([]m.Genre, error)
	GetGenresByNovelIDs(ctx context.Context, novelIDs []uuid.UUID) (map[uuid.UUID][]m.Genre, error)
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

// GetByID retrieves a genre by its ID with content counts
func (r *genreRepository) GetByID(ctx context.Context, id uuid.UUID) (*m.Genre, error) {
	query := `
		SELECT
			g.id,
			g.name,
			g.created_at,
			g.updated_at,
			COALESCE(ac.anime_count, 0) as anime_count,
			COALESCE(mc.manga_count, 0) as manga_count,
			COALESCE(nc.novel_count, 0) as novel_count
		FROM genre g
		LEFT JOIN (
			SELECT genre_id, COUNT(*) as anime_count
			FROM anime_genre
			WHERE genre_id = $1
			GROUP BY genre_id
		) ac ON g.id = ac.genre_id
		LEFT JOIN (
			SELECT genre_id, COUNT(*) as manga_count
			FROM manga_genre
			WHERE genre_id = $1
			GROUP BY genre_id
		) mc ON g.id = mc.genre_id
		LEFT JOIN (
			SELECT genre_id, COUNT(*) as novel_count
			FROM novel_genre
			WHERE genre_id = $1
			GROUP BY genre_id
		) nc ON g.id = nc.genre_id
		WHERE g.id = $1
	`

	var genre m.Genre
	var animeCount, mangaCount, novelCount int
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&genre.ID,
		&genre.Name,
		&genre.CreatedAt,
		&genre.UpdatedAt,
		&animeCount,
		&mangaCount,
		&novelCount,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get genre by ID: %w", err)
	}

	// Set counts
	genre.AnimeCount = animeCount
	genre.MangaCount = mangaCount
	genre.NovelCount = novelCount

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

	// Data query with content counts
	query := fmt.Sprintf(`
		SELECT
			g.id,
			g.name,
			g.created_at,
			g.updated_at,
			COALESCE(ac.anime_count, 0) as anime_count,
			COALESCE(mc.manga_count, 0) as manga_count,
			COALESCE(nc.novel_count, 0) as novel_count
		FROM genre g
		LEFT JOIN (
			SELECT genre_id, COUNT(*) as anime_count
			FROM anime_genre
			GROUP BY genre_id
		) ac ON g.id = ac.genre_id
		LEFT JOIN (
			SELECT genre_id, COUNT(*) as manga_count
			FROM manga_genre
			GROUP BY genre_id
		) mc ON g.id = mc.genre_id
		LEFT JOIN (
			SELECT genre_id, COUNT(*) as novel_count
			FROM novel_genre
			GROUP BY genre_id
		) nc ON g.id = nc.genre_id
		%s
		ORDER BY g.name ASC
		LIMIT $1 OFFSET $2
	`, whereClause)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list genres: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var genre m.Genre
		var animeCount, mangaCount, novelCount int
		err := rows.Scan(
			&genre.ID,
			&genre.Name,
			&genre.CreatedAt,
			&genre.UpdatedAt,
			&animeCount,
			&mangaCount,
			&novelCount,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan genre: %w", err)
		}

		// Set counts (assuming they're fields in the model)
		genre.AnimeCount = animeCount
		genre.MangaCount = mangaCount
		genre.NovelCount = novelCount

		genres = append(genres, &genre)
	}

	if rows.Err() != nil {
		return nil, 0, fmt.Errorf("failed to iterate genres: %w", rows.Err())
	}

	return genres, total, nil
}

// GetGenresByNovelID retrieves all genres associated with a specific novel
func (r *genreRepository) GetGenresByNovelID(ctx context.Context, novelID uuid.UUID) ([]m.Genre, error) {
	query := `
		SELECT g.id, g.name, g.created_at, g.updated_at
		FROM genre g
		INNER JOIN novel_genre ng ON g.id = ng.genre_id
		WHERE ng.novel_id = $1
		ORDER BY g.name ASC
	`

	rows, err := r.pool.Query(ctx, query, novelID)
	if err != nil {
		return nil, fmt.Errorf("failed to get genres by novel ID: %w", err)
	}
	defer rows.Close()

	var genres []m.Genre
	for rows.Next() {
		var genre m.Genre
		err := rows.Scan(
			&genre.ID,
			&genre.Name,
			&genre.CreatedAt,
			&genre.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan genre: %w", err)
		}
		genres = append(genres, genre)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to iterate genres: %w", rows.Err())
	}

	return genres, nil
}

// GetGenresByNovelIDs retrieves genres for multiple novels in a single query (batch loading)
// Returns a map where key is novel_id and value is the list of genres for that novel
func (r *genreRepository) GetGenresByNovelIDs(ctx context.Context, novelIDs []uuid.UUID) (map[uuid.UUID][]m.Genre, error) {
	if len(novelIDs) == 0 {
		return make(map[uuid.UUID][]m.Genre), nil
	}

	query := `
		SELECT ng.novel_id, g.id, g.name, g.created_at, g.updated_at
		FROM genre g
		INNER JOIN novel_genre ng ON g.id = ng.genre_id
		WHERE ng.novel_id = ANY($1)
		ORDER BY ng.novel_id, g.name ASC
	`

	rows, err := r.pool.Query(ctx, query, novelIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get genres by novel IDs: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID][]m.Genre)
	for rows.Next() {
		var novelID uuid.UUID
		var genre m.Genre
		err := rows.Scan(
			&novelID,
			&genre.ID,
			&genre.Name,
			&genre.CreatedAt,
			&genre.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan genre: %w", err)
		}
		result[novelID] = append(result[novelID], genre)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to iterate genres: %w", rows.Err())
	}

	return result, nil
}