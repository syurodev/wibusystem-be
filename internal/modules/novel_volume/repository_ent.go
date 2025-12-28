package novel_volume

import (
	"context"
	"system/internal/platform/database"
	"time"

	"system/internal/domain"
	ent "system/internal/ent/generated"
	"system/internal/ent/generated/novelchapter"
	"system/internal/ent/generated/novelvolume"

	"github.com/gofrs/uuid/v5"
)

// entVolumeRepository triển khai NovelVolumeRepository sử dụng Ent
type entVolumeRepository struct {
	client *ent.Client
}

// NewEntVolumeRepository tạo một instance mới của entVolumeRepository
func NewEntVolumeRepository(client *ent.Client) domain.NovelVolumeRepository {
	return &entVolumeRepository{client: client}
}

// entVolumeToDomain chuyển đổi Ent NovelVolume sang domain.NovelVolume
func entVolumeToDomain(v *ent.NovelVolume) *domain.NovelVolume {
	if v == nil {
		return nil
	}
	return &domain.NovelVolume{
		ID:            v.ID,
		NovelID:       v.NovelID,
		VolumeNumber:  v.VolumeNumber,
		Title:         v.Title,
		Slug:          v.Slug,
		Description:   v.Description,
		CoverImageURL: v.CoverImageURL,
		ChapterCount:  v.ChapterCount,
		WordCount:     v.WordCount,
		DisplayOrder:  v.DisplayOrder,
		IsPublished:   v.IsPublished,
		PublishedAt:   v.PublishedAt,
		CreatedBy:     v.CreatedBy,
		UpdatedBy:     v.UpdatedBy,
		DeletedBy:     v.DeletedBy,
		Version:       v.Version,
		CreatedAt:     v.CreatedAt,
		UpdatedAt:     v.UpdatedAt,
		DeletedAt:     v.DeletedAt,
	}
}

// GetByID lấy volume từ database theo ID
func (r *entVolumeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.NovelVolume, error) {
	v, err := database.GetClientFromContext(ctx, r.client).NovelVolume.Query().
		Where(novelvolume.ID(id), novelvolume.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return entVolumeToDomain(v), nil
}

// GetByNovelIDAndNumber lấy volume theo novel ID và volume number
func (r *entVolumeRepository) GetByNovelIDAndNumber(ctx context.Context, novelID uuid.UUID, volumeNumber int) (*domain.NovelVolume, error) {
	v, err := database.GetClientFromContext(ctx, r.client).NovelVolume.Query().
		Where(
			novelvolume.NovelID(novelID),
			novelvolume.VolumeNumber(volumeNumber),
			novelvolume.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return entVolumeToDomain(v), nil
}

// GetByNovelID lấy danh sách volume theo novel ID
func (r *entVolumeRepository) GetByNovelID(ctx context.Context, novelID uuid.UUID, publishedOnly bool) ([]*domain.NovelVolume, error) {
	query := database.GetClientFromContext(ctx, r.client).NovelVolume.Query().
		Where(novelvolume.NovelID(novelID), novelvolume.DeletedAtIsNil())

	if publishedOnly {
		query = query.Where(novelvolume.IsPublished(true))
	}

	volumes, err := query.
		Order(ent.Asc(novelvolume.FieldDisplayOrder), ent.Asc(novelvolume.FieldVolumeNumber)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.NovelVolume, len(volumes))
	for i, v := range volumes {
		result[i] = entVolumeToDomain(v)
	}
	return result, nil
}

// Create tạo volume mới trong database
func (r *entVolumeRepository) Create(ctx context.Context, v *domain.NovelVolume) error {
	builder := database.GetClientFromContext(ctx, r.client).NovelVolume.Create().
		SetID(v.ID).
		SetNovelID(v.NovelID).
		SetVolumeNumber(v.VolumeNumber).
		SetTitle(v.Title).
		SetSlug(v.Slug).
		SetDisplayOrder(v.DisplayOrder).
		SetIsPublished(v.IsPublished).
		SetCreatedBy(v.CreatedBy)

	if v.Description != nil {
		builder = builder.SetDescription(*v.Description)
	}
	if v.CoverImageURL != nil {
		builder = builder.SetCoverImageURL(*v.CoverImageURL)
	}

	_, err := builder.Save(ctx)
	return err
}

// Update cập nhật thông tin volume
func (r *entVolumeRepository) Update(ctx context.Context, v *domain.NovelVolume) error {
	builder := database.GetClientFromContext(ctx, r.client).NovelVolume.UpdateOneID(v.ID).
		SetVolumeNumber(v.VolumeNumber).
		SetTitle(v.Title).
		SetSlug(v.Slug).
		SetDisplayOrder(v.DisplayOrder).
		SetIsPublished(v.IsPublished)

	if v.Description != nil {
		builder = builder.SetDescription(*v.Description)
	} else {
		builder = builder.ClearDescription()
	}
	if v.CoverImageURL != nil {
		builder = builder.SetCoverImageURL(*v.CoverImageURL)
	} else {
		builder = builder.ClearCoverImageURL()
	}

	_, err := builder.Save(ctx)
	return err
}

// Delete xóa mềm volume
func (r *entVolumeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return database.GetClientFromContext(ctx, r.client).NovelVolume.UpdateOneID(id).
		SetDeletedAt(now).
		Exec(ctx)
}

// UpdateDisplayOrder cập nhật thứ tự hiển thị của volume
func (r *entVolumeRepository) UpdateDisplayOrder(ctx context.Context, id uuid.UUID, order int) error {
	return database.GetClientFromContext(ctx, r.client).NovelVolume.UpdateOneID(id).
		SetDisplayOrder(order).
		Exec(ctx)
}

// Publish xuất bản volume
func (r *entVolumeRepository) Publish(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return database.GetClientFromContext(ctx, r.client).NovelVolume.UpdateOneID(id).
		SetIsPublished(true).
		SetPublishedAt(now).
		Exec(ctx)
}

// Unpublish ẩn volume
func (r *entVolumeRepository) Unpublish(ctx context.Context, id uuid.UUID) error {
	return database.GetClientFromContext(ctx, r.client).NovelVolume.UpdateOneID(id).
		SetIsPublished(false).
		ClearPublishedAt().
		Exec(ctx)
}

// UpdateStatistics cập nhật chapter_count và word_count của volume
// dựa trên các chapters thuộc volume này
func (r *entVolumeRepository) UpdateStatistics(ctx context.Context, volumeID uuid.UUID) error {
	client := database.GetClientFromContext(ctx, r.client)

	// Get all chapters belonging to this volume (excluding deleted)
	chapters, err := client.NovelChapter.Query().
		Where(
			novelchapter.VolumeID(volumeID),
			novelchapter.DeletedAtIsNil(),
		).
		All(ctx)
	if err != nil {
		return err
	}

	// Calculate statistics
	chapterCount := len(chapters)
	var wordCount int64
	for _, ch := range chapters {
		wordCount += int64(ch.WordCount)
	}

	// Update volume with calculated stats
	_, err = client.NovelVolume.UpdateOneID(volumeID).
		SetChapterCount(chapterCount).
		SetWordCount(wordCount).
		Save(ctx)

	return err
}
