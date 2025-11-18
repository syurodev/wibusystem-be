package repository

import (
	"context"
	"system/internal/domain"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// volumeRepository triển khai VolumeRepository sử dụng pgx
type volumeRepository struct {
	pool *pgxpool.Pool
}

// NewVolumeRepository tạo một instance mới của volumeRepository
func NewVolumeRepository(pool *pgxpool.Pool) domain.VolumeRepository {
	return &volumeRepository{pool: pool}
}

// GetByID lấy volume từ database theo ID
func (r *volumeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Volume, error) {
	query := `
		SELECT id, novel_id, volume_number, title, slug, description,
		       cover_image_url, chapter_count, word_count, display_order,
		       is_published, published_at, created_at, updated_at, deleted_at
		FROM catalog.volumes
		WHERE id = $1 AND deleted_at IS NULL
	`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}

	volume, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Volume])
	if err != nil {
		return nil, err
	}

	return &volume, nil
}

// GetByNovelIDAndNumber lấy volume theo novel ID và volume number
func (r *volumeRepository) GetByNovelIDAndNumber(ctx context.Context, novelID uuid.UUID, volumeNumber int) (*domain.Volume, error) {
	query := `
		SELECT id, novel_id, volume_number, title, slug, description,
		       cover_image_url, chapter_count, word_count, display_order,
		       is_published, published_at, created_at, updated_at, deleted_at
		FROM catalog.volumes
		WHERE novel_id = $1 AND volume_number = $2 AND deleted_at IS NULL
	`

	rows, err := r.pool.Query(ctx, query, novelID, volumeNumber)
	if err != nil {
		return nil, err
	}

	volume, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Volume])
	if err != nil {
		return nil, err
	}

	return &volume, nil
}

// GetByNovelID lấy danh sách volume theo novel ID
func (r *volumeRepository) GetByNovelID(ctx context.Context, novelID uuid.UUID, publishedOnly bool) ([]*domain.Volume, error) {
	query := `
		SELECT id, novel_id, volume_number, title, slug, description,
		       cover_image_url, chapter_count, word_count, display_order,
		       is_published, published_at, created_at, updated_at, deleted_at
		FROM catalog.volumes
		WHERE novel_id = $1 AND deleted_at IS NULL
	`

	if publishedOnly {
		query += " AND is_published = true"
	}

	query += " ORDER BY display_order ASC, volume_number ASC"

	rows, err := r.pool.Query(ctx, query, novelID)
	if err != nil {
		return nil, err
	}

	volumes, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.Volume])
	if err != nil {
		return nil, err
	}

	return volumes, nil
}

// Create tạo volume mới trong database
func (r *volumeRepository) Create(ctx context.Context, volume *domain.Volume) error {
	query := `
		INSERT INTO catalog.volumes (
			id, novel_id, volume_number, title, slug, description,
			cover_image_url, display_order, is_published
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.pool.Exec(ctx, query,
		volume.ID,
		volume.NovelID,
		volume.VolumeNumber,
		volume.Title,
		volume.Slug,
		volume.Description,
		volume.CoverImageURL,
		volume.DisplayOrder,
		volume.IsPublished,
	)

	return err
}

// Update cập nhật thông tin volume
func (r *volumeRepository) Update(ctx context.Context, volume *domain.Volume) error {
	query := `
		UPDATE catalog.volumes
		SET volume_number = $2,
		    title = $3,
		    slug = $4,
		    description = $5,
		    cover_image_url = $6,
		    display_order = $7,
		    is_published = $8
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query,
		volume.ID,
		volume.VolumeNumber,
		volume.Title,
		volume.Slug,
		volume.Description,
		volume.CoverImageURL,
		volume.DisplayOrder,
		volume.IsPublished,
	)

	return err
}

// Delete xóa mềm volume
func (r *volumeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE catalog.volumes
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// UpdateDisplayOrder cập nhật thứ tự hiển thị của volume
func (r *volumeRepository) UpdateDisplayOrder(ctx context.Context, id uuid.UUID, order int) error {
	query := `
		UPDATE catalog.volumes
		SET display_order = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, id, order)
	return err
}

// Publish xuất bản volume
func (r *volumeRepository) Publish(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE catalog.volumes
		SET is_published = true,
		    published_at = COALESCE(published_at, NOW())
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// Unpublish ẩn volume
func (r *volumeRepository) Unpublish(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE catalog.volumes
		SET is_published = false
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, id)
	return err
}
