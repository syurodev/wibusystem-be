package postgres

import (
	"context"
	"fmt"
	"wibusystem/internal/modules/catalog/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// VolumeRepositoryPG is a PostgreSQL implementation of VolumeRepository.
type VolumeRepositoryPG struct {
	DB *pgxpool.Pool
}

// NewVolumeRepositoryPG creates a new instance of VolumeRepositoryPG.
func NewVolumeRepositoryPG(db *pgxpool.Pool) *VolumeRepositoryPG {
	return &VolumeRepositoryPG{DB: db}
}

// Create inserts a new volume into the database.
func (r *VolumeRepositoryPG) Create(ctx context.Context, volume *domain.Volume) error {
	query := `
		INSERT INTO volumes (
			id, novel_id, volume_title, description, cover_image, volume_number, published_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.DB.Exec(ctx, query,
		volume.ID, volume.NovelID, volume.VolumeTitle, volume.Description, volume.CoverImage,
		volume.VolumeNumber, volume.PublishedAt, volume.CreatedAt, volume.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("postgres.VolumeRepositoryPG.Create: %w", err)
	}

	return nil
}

// GetByID retrieves a volume by its ID.
func (r *VolumeRepositoryPG) GetByID(ctx context.Context, id uuid.UUID) (*domain.Volume, error) {
	query := `SELECT id, novel_id, volume_title, description, cover_image, volume_number, published_at, created_at, updated_at FROM volumes WHERE id = $1`
	volume := &domain.Volume{}
	row := r.DB.QueryRow(ctx, query, id)

	err := row.Scan(
		&volume.ID, &volume.NovelID, &volume.VolumeTitle, &volume.Description, &volume.CoverImage,
		&volume.VolumeNumber, &volume.PublishedAt, &volume.CreatedAt, &volume.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("volume with id %s not found", id)
		}
		return nil, fmt.Errorf("postgres.VolumeRepositoryPG.GetByID: %w", err)
	}

	return volume, nil
}

// Update modifies an existing volume in the database.
func (r *VolumeRepositoryPG) Update(ctx context.Context, volume *domain.Volume) error {
	query := `
		UPDATE volumes SET
			novel_id = $2, volume_title = $3, description = $4, cover_image = $5,
			volume_number = $6, published_at = $7, updated_at = $8
		WHERE id = $1`

	_, err := r.DB.Exec(ctx, query,
		volume.ID, volume.NovelID, volume.VolumeTitle, volume.Description, volume.CoverImage,
		volume.VolumeNumber, volume.PublishedAt, volume.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("postgres.VolumeRepositoryPG.Update: %w", err)
	}

	return nil
}

// Delete removes a volume from the database.
func (r *VolumeRepositoryPG) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM volumes WHERE id = $1`
	_, err := r.DB.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("postgres.VolumeRepositoryPG.Delete: %w", err)
	}
	return nil
}

// ListByNovelID retrieves all volumes for a given novel.
func (r *VolumeRepositoryPG) ListByNovelID(ctx context.Context, novelID uuid.UUID, limit, offset int) ([]*domain.Volume, int64, error) {
	query := `
		SELECT id, novel_id, volume_title, description, cover_image, volume_number, published_at, created_at, updated_at
		FROM volumes
		WHERE novel_id = $1
		ORDER BY volume_number ASC
		LIMIT $2 OFFSET $3`

	rows, err := r.DB.Query(ctx, query, novelID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("postgres.VolumeRepositoryPG.ListByNovelID: %w", err)
	}
	defer rows.Close()

	volumes := []*domain.Volume{}
	for rows.Next() {
		volume := &domain.Volume{}
		err := rows.Scan(
			&volume.ID, &volume.NovelID, &volume.VolumeTitle, &volume.Description, &volume.CoverImage,
			&volume.VolumeNumber, &volume.PublishedAt, &volume.CreatedAt, &volume.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("postgres.VolumeRepositoryPG.ListByNovelID: failed to scan volume: %w", err)
		}
		volumes = append(volumes, volume)
	}

	countQuery := `SELECT COUNT(*) FROM volumes WHERE novel_id = $1`
	var total int64
	err = r.DB.QueryRow(ctx, countQuery, novelID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("postgres.VolumeRepositoryPG.ListByNovelID: failed to count volumes: %w", err)
	}

	return volumes, total, nil
}
