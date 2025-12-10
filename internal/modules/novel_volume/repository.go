package novel_volume

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
		SELECT v.id, v.novel_id, v.volume_number, v.title, v.slug, v.description,
		       v.cover_image_url, v.chapter_count, v.word_count, v.display_order,
		       v.is_published, v.published_at, v.created_at, v.updated_at, v.deleted_at, v.created_by,
		       n.title as novel_title
		FROM catalog.novel_volumes v
		LEFT JOIN catalog.novels n ON v.novel_id = n.id
		WHERE v.id = $1 AND v.deleted_at IS NULL
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
		SELECT v.id, v.novel_id, v.volume_number, v.title, v.slug, v.description,
		       v.cover_image_url, v.chapter_count, v.word_count, v.display_order,
		       v.is_published, v.published_at, v.created_at, v.updated_at, v.deleted_at, v.created_by,
		       n.title as novel_title
		FROM catalog.novel_volumes v
		LEFT JOIN catalog.novels n ON v.novel_id = n.id
		WHERE v.novel_id = $1 AND v.volume_number = $2 AND v.deleted_at IS NULL
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
		SELECT v.id, v.novel_id, v.volume_number, v.title, v.slug, v.description,
		       v.cover_image_url, v.chapter_count, v.word_count, v.display_order,
		       v.is_published, v.published_at, v.created_at, v.updated_at, v.deleted_at, v.created_by,
		       n.title as novel_title
		FROM catalog.novel_volumes v
		LEFT JOIN catalog.novels n ON v.novel_id = n.id
		WHERE v.novel_id = $1 AND v.deleted_at IS NULL
	`

	if publishedOnly {
		query += " AND v.is_published = true"
	}

	query += " ORDER BY v.display_order ASC, v.volume_number ASC"

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
		INSERT INTO catalog.novel_volumes (
			id, novel_id, volume_number, title, slug, description,
			cover_image_url, display_order, is_published, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
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
		volume.CreatedBy,
	)

	return err
}

// Update cập nhật thông tin volume
func (r *volumeRepository) Update(ctx context.Context, volume *domain.Volume) error {
	query := `
		UPDATE catalog.novel_volumes
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
		UPDATE catalog.novel_volumes
		SET deleted_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// UpdateDisplayOrder cập nhật thứ tự hiển thị của volume
func (r *volumeRepository) UpdateDisplayOrder(ctx context.Context, id uuid.UUID, order int) error {
	query := `
		UPDATE catalog.novel_volumes
		SET display_order = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, id, order)
	return err
}

// Publish xuất bản volume
func (r *volumeRepository) Publish(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE catalog.novel_volumes
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
		UPDATE catalog.novel_volumes
		SET is_published = false
		WHERE id = $1 AND deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// UpdateStatistics cập nhật chapter_count và word_count của volume dựa trên chapters
func (r *volumeRepository) UpdateStatistics(ctx context.Context, volumeID uuid.UUID) error {
	query := `
		UPDATE catalog.novel_volumes v
		SET 
			chapter_count = COALESCE((
				SELECT COUNT(*)
				FROM catalog.novel_chapters c
				WHERE c.volume_id = v.id AND c.deleted_at IS NULL
			), 0),
			word_count = COALESCE((
				SELECT SUM(word_count)
				FROM catalog.novel_chapters c
				WHERE c.volume_id = v.id AND c.deleted_at IS NULL
			), 0)
		WHERE v.id = $1 AND v.deleted_at IS NULL
	`

	_, err := r.pool.Exec(ctx, query, volumeID)
	return err
}
