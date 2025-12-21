// ============================================================================
// Novel Volume Repository
// ============================================================================
//
// Repository này triển khai NovelVolumeRepository interface từ domain package.
// Volume là cấp giữa trong cấu trúc phân cấp: Novel > Volume > Chapter.
//
// CRUD Operations:
//   - GetByID: Lấy volume theo ID
//   - GetByNovelIDAndNumber: Lấy volume theo novel ID và volume number
//   - GetByNovelID: Lấy danh sách volumes của một novel (hỗ trợ publishedOnly filter)
//   - Create: Tạo volume mới
//   - Update: Cập nhật thông tin volume
//   - Delete: Soft delete volume
//
// State Operations:
//   - Publish: Xuất bản volume (is_published = true)
//   - Unpublish: Ẩn volume (is_published = false)
//   - UpdateDisplayOrder: Cập nhật thứ tự hiển thị
//
// Statistics:
//   - UpdateStatistics: Cập nhật chapter_count và word_count từ chapters
//
// SQL queries được load từ thư mục queries/ sử dụng go:embed.
//
// ============================================================================

package novel_volume

import (
	"context"
	_ "embed"
	"system/internal/domain"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SQL queries embedded từ files
//
//go:embed queries/get_by_id.sql
var getByIDQuery string

//go:embed queries/get_by_novel_id_and_number.sql
var getByNovelIDAndNumberQuery string

//go:embed queries/create.sql
var createQuery string

//go:embed queries/update.sql
var updateQuery string

//go:embed queries/delete.sql
var deleteQuery string

//go:embed queries/update_display_order.sql
var updateDisplayOrderQuery string

//go:embed queries/publish.sql
var publishQuery string

//go:embed queries/unpublish.sql
var unpublishQuery string

//go:embed queries/update_statistics.sql
var updateStatisticsQuery string

// volumeRepository triển khai NovelVolumeRepository sử dụng pgx
type volumeRepository struct {
	pool *pgxpool.Pool
}

// NewVolumeRepository tạo một instance mới của volumeRepository
func NewVolumeRepository(pool *pgxpool.Pool) domain.NovelVolumeRepository {
	return &volumeRepository{pool: pool}
}

// GetByID lấy volume từ database theo ID
func (r *volumeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.NovelVolume, error) {
	rows, err := r.pool.Query(ctx, getByIDQuery, id)
	if err != nil {
		return nil, err
	}

	volume, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.NovelVolume])
	if err != nil {
		return nil, err
	}

	return &volume, nil
}

// GetByNovelIDAndNumber lấy volume theo novel ID và volume number
func (r *volumeRepository) GetByNovelIDAndNumber(ctx context.Context, novelID uuid.UUID, volumeNumber int) (*domain.NovelVolume, error) {
	rows, err := r.pool.Query(ctx, getByNovelIDAndNumberQuery, novelID, volumeNumber)
	if err != nil {
		return nil, err
	}

	volume, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.NovelVolume])
	if err != nil {
		return nil, err
	}

	return &volume, nil
}

// GetByNovelID lấy danh sách volume theo novel ID
func (r *volumeRepository) GetByNovelID(ctx context.Context, novelID uuid.UUID, publishedOnly bool) ([]*domain.NovelVolume, error) {
	// Query cơ bản từ getByIDQuery nhưng cần điều chỉnh WHERE clause
	// Không thể embed trực tiếp vì cần dynamic filter
	query := `
		SELECT v.id, v.novel_id, v.volume_number, v.title, v.slug, v.description,
		       v.cover_image_url, v.chapter_count, v.word_count, v.display_order,
		       v.is_published, v.published_at, v.created_by, v.updated_by, v.deleted_by,
		       v.version, v.created_at, v.updated_at, v.deleted_at,
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

	volumes, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[domain.NovelVolume])
	if err != nil {
		return nil, err
	}

	return volumes, nil
}

// Create tạo volume mới trong database
func (r *volumeRepository) Create(ctx context.Context, volume *domain.NovelVolume) error {
	_, err := r.pool.Exec(ctx, createQuery,
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
func (r *volumeRepository) Update(ctx context.Context, volume *domain.NovelVolume) error {
	_, err := r.pool.Exec(ctx, updateQuery,
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
	_, err := r.pool.Exec(ctx, deleteQuery, id)
	return err
}

// UpdateDisplayOrder cập nhật thứ tự hiển thị của volume
func (r *volumeRepository) UpdateDisplayOrder(ctx context.Context, id uuid.UUID, order int) error {
	_, err := r.pool.Exec(ctx, updateDisplayOrderQuery, id, order)
	return err
}

// Publish xuất bản volume
func (r *volumeRepository) Publish(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, publishQuery, id)
	return err
}

// Unpublish ẩn volume
func (r *volumeRepository) Unpublish(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, unpublishQuery, id)
	return err
}

// UpdateStatistics cập nhật chapter_count và word_count của volume dựa trên chapters
func (r *volumeRepository) UpdateStatistics(ctx context.Context, volumeID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, updateStatisticsQuery, volumeID)
	return err
}
