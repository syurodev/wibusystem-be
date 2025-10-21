package postgres

import (
	"context"
	"fmt"
	"wibusystem/internal/modules/catalog/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ChapterRepositoryPG is a PostgreSQL implementation of ChapterRepository.
type ChapterRepositoryPG struct {
	DB *pgxpool.Pool
}

// NewChapterRepositoryPG creates a new instance of ChapterRepositoryPG.
func NewChapterRepositoryPG(db *pgxpool.Pool) *ChapterRepositoryPG {
	return &ChapterRepositoryPG{DB: db}
}

// Create inserts a new chapter into the database.
func (r *ChapterRepositoryPG) Create(ctx context.Context, chapter *domain.Chapter) error {
	query := `
		INSERT INTO chapters (
			id, volume_id, title, content, chapter_number, version, published_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.DB.Exec(ctx, query,
		chapter.ID, chapter.VolumeID, chapter.Title, chapter.Content, chapter.ChapterNumber,
		chapter.Version, chapter.PublishedAt, chapter.CreatedAt, chapter.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("postgres.ChapterRepositoryPG.Create: %w", err)
	}

	return nil
}

// GetByID retrieves a chapter by its ID.
func (r *ChapterRepositoryPG) GetByID(ctx context.Context, id uuid.UUID) (*domain.Chapter, error) {
	query := `SELECT id, volume_id, title, content, chapter_number, version, published_at, created_at, updated_at FROM chapters WHERE id = $1`
	chapter := &domain.Chapter{}
	row := r.DB.QueryRow(ctx, query, id)

	err := row.Scan(
		&chapter.ID, &chapter.VolumeID, &chapter.Title, &chapter.Content, &chapter.ChapterNumber,
		&chapter.Version, &chapter.PublishedAt, &chapter.CreatedAt, &chapter.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("chapter with id %s not found", id)
		}
		return nil, fmt.Errorf("postgres.ChapterRepositoryPG.GetByID: %w", err)
	}

	return chapter, nil
}

// Update modifies an existing chapter in the database.
func (r *ChapterRepositoryPG) Update(ctx context.Context, chapter *domain.Chapter) error {
	query := `
		UPDATE chapters SET
			volume_id = $2, title = $3, content = $4, chapter_number = $5,
			version = $6, published_at = $7, updated_at = $8
		WHERE id = $1`

	_, err := r.DB.Exec(ctx, query,
		chapter.ID, chapter.VolumeID, chapter.Title, chapter.Content, chapter.ChapterNumber,
		chapter.Version, chapter.PublishedAt, chapter.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("postgres.ChapterRepositoryPG.Update: %w", err)
	}

	return nil
}

// Delete removes a chapter from the database.
func (r *ChapterRepositoryPG) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM chapters WHERE id = $1`
	_, err := r.DB.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("postgres.ChapterRepositoryPG.Delete: %w", err)
	}
	return nil
}

// ListByVolumeID retrieves all chapters for a given volume.
func (r *ChapterRepositoryPG) ListByVolumeID(ctx context.Context, volumeID uuid.UUID, limit, offset int) ([]*domain.Chapter, int64, error) {
	query := `
		SELECT id, volume_id, title, content, chapter_number, version, published_at, created_at, updated_at
		FROM chapters
		WHERE volume_id = $1
		ORDER BY chapter_number ASC
		LIMIT $2 OFFSET $3`

	rows, err := r.DB.Query(ctx, query, volumeID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("postgres.ChapterRepositoryPG.ListByVolumeID: %w", err)
	}
	defer rows.Close()

	chapters := []*domain.Chapter{}
	for rows.Next() {
		chapter := &domain.Chapter{}
		err := rows.Scan(
			&chapter.ID, &chapter.VolumeID, &chapter.Title, &chapter.Content, &chapter.ChapterNumber,
			&chapter.Version, &chapter.PublishedAt, &chapter.CreatedAt, &chapter.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("postgres.ChapterRepositoryPG.ListByVolumeID: failed to scan chapter: %w", err)
		}
		chapters = append(chapters, chapter)
	}

	countQuery := `SELECT COUNT(*) FROM chapters WHERE volume_id = $1`
	var total int64
	err = r.DB.QueryRow(ctx, countQuery, volumeID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("postgres.ChapterRepositoryPG.ListByVolumeID: failed to count chapters: %w", err)
	}

	return chapters, total, nil
}
