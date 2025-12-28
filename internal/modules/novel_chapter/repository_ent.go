package novel_chapter

import (
	"context"
	"system/internal/platform/database"
	"time"

	"system/internal/domain"
	ent "system/internal/ent/generated"
	"system/internal/ent/generated/novelchapter"

	"github.com/gofrs/uuid/v5"
)

// entChapterRepository triển khai NovelChapterRepository sử dụng Ent
type entChapterRepository struct {
	client *ent.Client
}

// NewEntChapterRepository tạo một instance mới của entChapterRepository
func NewEntChapterRepository(client *ent.Client) domain.NovelChapterRepository {
	return &entChapterRepository{client: client}
}

// entChapterToDomain chuyển đổi Ent NovelChapter sang domain.NovelChapter
func entChapterToDomain(c *ent.NovelChapter) *domain.NovelChapter {
	if c == nil {
		return nil
	}
	return &domain.NovelChapter{
		ID:             c.ID,
		NovelID:        c.NovelID,
		VolumeID:       c.VolumeID,
		ChapterNumber:  c.ChapterNumber,
		Title:          c.Title,
		Slug:           c.Slug,
		Content:        c.Content,
		WordCount:      c.WordCount,
		CharacterCount: c.CharacterCount,
		IsFree:         c.IsFree,
		Price:          c.Price,
		Currency:       c.Currency,
		Status:         domain.NovelChapterStatus(c.Status.String()),
		ViewCount:      c.ViewCount,
		LikeCount:      c.LikeCount,
		CommentCount:   c.CommentCount,
		DisplayOrder:   c.DisplayOrder,
		AuthorNotes:    c.AuthorNotes,
		PublishedAt:    c.PublishedAt,
		ScheduledAt:    c.ScheduledAt,
		CreatedBy:      c.CreatedBy,
		UpdatedBy:      c.UpdatedBy,
		DeletedBy:      c.DeletedBy,
		Version:        c.Version,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
		DeletedAt:      c.DeletedAt,
	}
}

// GetByID lấy chapter từ database theo ID
func (r *entChapterRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.NovelChapter, error) {
	c, err := database.GetClientFromContext(ctx, r.client).NovelChapter.Query().
		Where(novelchapter.ID(id), novelchapter.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return entChapterToDomain(c), nil
}

// GetByNovelIDAndNumber lấy chapter theo novel ID và chapter number
func (r *entChapterRepository) GetByNovelIDAndNumber(ctx context.Context, novelID uuid.UUID, chapterNumber int) (*domain.NovelChapter, error) {
	c, err := database.GetClientFromContext(ctx, r.client).NovelChapter.Query().
		Where(
			novelchapter.NovelID(novelID),
			novelchapter.ChapterNumber(chapterNumber),
			novelchapter.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return entChapterToDomain(c), nil
}

// GetBySlug lấy chapter theo slug
func (r *entChapterRepository) GetBySlug(ctx context.Context, slug string) (*domain.NovelChapter, error) {
	c, err := database.GetClientFromContext(ctx, r.client).NovelChapter.Query().
		Where(novelchapter.Slug(slug), novelchapter.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return entChapterToDomain(c), nil
}

// GetByNovelID lấy danh sách chapter theo novel ID
func (r *entChapterRepository) GetByNovelID(ctx context.Context, novelID uuid.UUID, filter domain.NovelChapterFilter) ([]*domain.NovelChapter, error) {
	query := database.GetClientFromContext(ctx, r.client).NovelChapter.Query().
		Where(novelchapter.NovelID(novelID), novelchapter.DeletedAtIsNil())

	// Filter by status
	if filter.Status != nil && *filter.Status != "" {
		query = query.Where(novelchapter.StatusEQ(novelchapter.Status(*filter.Status)))
	}

	// Filter publishedOnly
	if filter.PublishedOnly {
		query = query.Where(novelchapter.StatusEQ(novelchapter.StatusPublished))
	}

	// Order
	orderFunc := ent.Asc(novelchapter.FieldChapterNumber)
	if filter.SortOrder == "desc" {
		orderFunc = ent.Desc(novelchapter.FieldChapterNumber)
	}

	// Pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	chapters, err := query.Order(orderFunc).All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.NovelChapter, len(chapters))
	for i, c := range chapters {
		result[i] = entChapterToDomain(c)
	}
	return result, nil
}

// GetByVolumeID lấy danh sách chapter theo volume ID
func (r *entChapterRepository) GetByVolumeID(ctx context.Context, volumeID uuid.UUID, publishedOnly bool) ([]*domain.NovelChapter, error) {
	query := database.GetClientFromContext(ctx, r.client).NovelChapter.Query().
		Where(novelchapter.VolumeID(volumeID), novelchapter.DeletedAtIsNil())

	if publishedOnly {
		query = query.Where(novelchapter.StatusEQ(novelchapter.StatusPublished))
	}

	chapters, err := query.
		Order(ent.Asc(novelchapter.FieldChapterNumber)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.NovelChapter, len(chapters))
	for i, c := range chapters {
		result[i] = entChapterToDomain(c)
	}
	return result, nil
}

// Create tạo chapter mới trong database
func (r *entChapterRepository) Create(ctx context.Context, c *domain.NovelChapter) error {
	builder := database.GetClientFromContext(ctx, r.client).NovelChapter.Create().
		SetID(c.ID).
		SetNovelID(c.NovelID).
		SetChapterNumber(c.ChapterNumber).
		SetTitle(c.Title).
		SetSlug(c.Slug).
		SetWordCount(c.WordCount).
		SetCharacterCount(c.CharacterCount).
		SetIsFree(c.IsFree).
		SetStatus(novelchapter.Status(c.Status)).
		SetDisplayOrder(c.DisplayOrder).
		SetCreatedBy(c.CreatedBy)

	if c.VolumeID != nil {
		builder = builder.SetVolumeID(*c.VolumeID)
	}
	if c.Content != nil {
		builder = builder.SetContent(c.Content)
	}
	if c.Price != nil {
		builder = builder.SetPrice(*c.Price)
	}
	if c.Currency != nil {
		builder = builder.SetCurrency(*c.Currency)
	}
	if c.AuthorNotes != nil {
		builder = builder.SetAuthorNotes(c.AuthorNotes)
	}

	_, err := builder.Save(ctx)
	return err
}

// Update cập nhật thông tin chapter
func (r *entChapterRepository) Update(ctx context.Context, c *domain.NovelChapter) error {
	builder := database.GetClientFromContext(ctx, r.client).NovelChapter.UpdateOneID(c.ID).
		SetChapterNumber(c.ChapterNumber).
		SetTitle(c.Title).
		SetSlug(c.Slug).
		SetWordCount(c.WordCount).
		SetCharacterCount(c.CharacterCount).
		SetIsFree(c.IsFree).
		SetDisplayOrder(c.DisplayOrder)

	if c.VolumeID != nil {
		builder = builder.SetVolumeID(*c.VolumeID)
	} else {
		builder = builder.ClearVolumeID()
	}
	if c.Content != nil {
		builder = builder.SetContent(c.Content)
	}
	if c.UpdatedBy != nil {
		builder = builder.SetUpdatedBy(*c.UpdatedBy)
	}

	_, err := builder.Save(ctx)
	return err
}

// Delete xóa mềm chapter
func (r *entChapterRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return database.GetClientFromContext(ctx, r.client).NovelChapter.UpdateOneID(id).
		SetDeletedAt(now).
		Exec(ctx)
}

// Publish xuất bản chapter ngay lập tức
func (r *entChapterRepository) Publish(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return database.GetClientFromContext(ctx, r.client).NovelChapter.UpdateOneID(id).
		SetStatus(novelchapter.StatusPublished).
		SetPublishedAt(now).
		Exec(ctx)
}

// Schedule đặt lịch xuất bản chapter
func (r *entChapterRepository) Schedule(ctx context.Context, id uuid.UUID, scheduledAt time.Time) error {
	return database.GetClientFromContext(ctx, r.client).NovelChapter.UpdateOneID(id).
		SetStatus(novelchapter.StatusScheduled).
		SetScheduledAt(scheduledAt).
		Exec(ctx)
}

// GetScheduledChapters lấy danh sách chapter cần xuất bản
func (r *entChapterRepository) GetScheduledChapters(ctx context.Context, before time.Time) ([]*domain.NovelChapter, error) {
	chapters, err := database.GetClientFromContext(ctx, r.client).NovelChapter.Query().
		Where(
			novelchapter.StatusEQ(novelchapter.StatusScheduled),
			novelchapter.ScheduledAtLTE(before),
			novelchapter.DeletedAtIsNil(),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.NovelChapter, len(chapters))
	for i, c := range chapters {
		result[i] = entChapterToDomain(c)
	}
	return result, nil
}

// IncrementViewCount tăng view count
func (r *entChapterRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	return database.GetClientFromContext(ctx, r.client).NovelChapter.UpdateOneID(id).
		AddViewCount(1).
		Exec(ctx)
}

// UpdateStatistics cập nhật thống kê của chapter
func (r *entChapterRepository) UpdateStatistics(ctx context.Context, id uuid.UUID, stats domain.NovelChapterStatistics) error {
	builder := database.GetClientFromContext(ctx, r.client).NovelChapter.UpdateOneID(id)

	if stats.ViewCount != nil {
		builder = builder.SetViewCount(*stats.ViewCount)
	}
	if stats.LikeCount != nil {
		builder = builder.SetLikeCount(*stats.LikeCount)
	}
	if stats.CommentCount != nil {
		builder = builder.SetCommentCount(*stats.CommentCount)
	}

	_, err := builder.Save(ctx)
	return err
}

// BatchIncrementViewCount tăng view count cho nhiều chapters
func (r *entChapterRepository) BatchIncrementViewCount(ctx context.Context, increments map[uuid.UUID]int64) error {
	if len(increments) == 0 {
		return nil
	}

	for id, count := range increments {
		_, err := database.GetClientFromContext(ctx, r.client).NovelChapter.UpdateOneID(id).
			AddViewCount(count).
			Save(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetRecentChapters retrieves recently published chapters across all novels
func (r *entChapterRepository) GetRecentChapters(ctx context.Context, limit int) ([]*domain.NovelChapter, error) {
	chapters, err := database.GetClientFromContext(ctx, r.client).NovelChapter.Query().
		Where(
			novelchapter.StatusEQ(novelchapter.StatusPublished),
			novelchapter.DeletedAtIsNil(),
		).
		Order(ent.Desc(novelchapter.FieldPublishedAt)).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.NovelChapter, len(chapters))
	for i, c := range chapters {
		result[i] = entChapterToDomain(c)
	}
	return result, nil
}
