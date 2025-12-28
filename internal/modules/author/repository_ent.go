package author

import (
	"context"
	"system/internal/platform/database"
	"time"

	"system/internal/domain"
	ent "system/internal/ent/generated"
	"system/internal/ent/generated/author"
	"system/internal/ent/generated/novelauthor"

	"github.com/gofrs/uuid/v5"
)

// entAuthorRepository triển khai AuthorRepository sử dụng Ent
type entAuthorRepository struct {
	client *ent.Client
}

// NewEntAuthorRepository tạo một instance mới của entAuthorRepository
func NewEntAuthorRepository(client *ent.Client) domain.AuthorRepository {
	return &entAuthorRepository{client: client}
}

// entAuthorToDomain chuyển đổi Ent Author sang domain.Author
func entAuthorToDomain(a *ent.Author) *domain.Author {
	if a == nil {
		return nil
	}
	return &domain.Author{
		ID:            a.ID,
		UserID:        a.UserID,
		Name:          a.Name,
		Slug:          a.Slug,
		Biography:     a.Biography,
		AvatarURL:     a.AvatarURL,
		SocialLinks:   a.SocialLinks,
		NovelCount:    a.NovelCount,
		TotalChapters: a.TotalChapters,
		TotalViews:    a.TotalViews,
		FollowerCount: a.FollowerCount,
		IsVerified:    a.IsVerified,
		Metadata:      a.Metadata,
		CreatedBy:     a.CreatedBy,
		UpdatedBy:     a.UpdatedBy,
		DeletedBy:     a.DeletedBy,
		Version:       a.Version,
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
		DeletedAt:     a.DeletedAt,
	}
}

// GetByID lấy author từ database theo ID
func (r *entAuthorRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Author, error) {
	a, err := database.GetClientFromContext(ctx, r.client).Author.Query().
		Where(author.ID(id), author.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return entAuthorToDomain(a), nil
}

// GetBySlug lấy author từ database theo slug
func (r *entAuthorRepository) GetBySlug(ctx context.Context, slug string) (*domain.Author, error) {
	a, err := database.GetClientFromContext(ctx, r.client).Author.Query().
		Where(author.Slug(slug), author.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return entAuthorToDomain(a), nil
}

// GetByUserID lấy author từ database theo user ID
func (r *entAuthorRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Author, error) {
	a, err := database.GetClientFromContext(ctx, r.client).Author.Query().
		Where(author.UserID(userID), author.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return entAuthorToDomain(a), nil
}

// List lấy danh sách authors với filter
func (r *entAuthorRepository) List(ctx context.Context, filter domain.AuthorFilter) ([]*domain.Author, int64, error) {
	query := database.GetClientFromContext(ctx, r.client).Author.Query().Where(author.DeletedAtIsNil())

	if filter.SearchQuery != nil && *filter.SearchQuery != "" {
		query = query.Where(author.NameContainsFold(*filter.SearchQuery))
	}

	if filter.IsVerified != nil {
		query = query.Where(author.IsVerified(*filter.IsVerified))
	}

	// Count total
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Sort
	orderFunc := ent.Asc(author.FieldCreatedAt)
	if filter.SortBy != "" {
		orderField := author.FieldCreatedAt
		switch filter.SortBy {
		case "name":
			orderField = author.FieldName
		case "novel_count":
			orderField = author.FieldNovelCount
		case "created_at":
			orderField = author.FieldCreatedAt
		}
		if filter.SortOrder == "desc" {
			orderFunc = ent.Desc(orderField)
		} else {
			orderFunc = ent.Asc(orderField)
		}
	}

	authors, err := query.
		Order(orderFunc).
		Offset(filter.Offset).
		Limit(filter.Limit).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*domain.Author, len(authors))
	for i, a := range authors {
		result[i] = entAuthorToDomain(a)
	}
	return result, int64(total), nil
}

// ListSelection lấy danh sách authors rút gọn (chỉ ID và Name)
func (r *entAuthorRepository) ListSelection(ctx context.Context, offset, limit int, search string) ([]*domain.Author, int64, error) {
	query := database.GetClientFromContext(ctx, r.client).Author.Query().Where(author.DeletedAtIsNil())

	if search != "" {
		query = query.Where(author.NameContainsFold(search))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	authors, err := query.
		Select(author.FieldID, author.FieldName).
		Order(ent.Asc(author.FieldName)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*domain.Author, len(authors))
	for i, a := range authors {
		result[i] = &domain.Author{
			ID:   a.ID,
			Name: a.Name,
		}
	}
	return result, int64(total), nil
}

// Create tạo author mới
func (r *entAuthorRepository) Create(ctx context.Context, a *domain.Author) error {
	builder := database.GetClientFromContext(ctx, r.client).Author.Create().
		SetID(a.ID).
		SetName(a.Name).
		SetSlug(a.Slug).
		SetCreatedBy(a.CreatedBy)

	if a.UserID != nil {
		builder = builder.SetUserID(*a.UserID)
	}
	if a.Biography != nil {
		builder = builder.SetBiography(a.Biography)
	}
	if a.AvatarURL != nil {
		builder = builder.SetAvatarURL(*a.AvatarURL)
	}
	if a.SocialLinks != nil {
		builder = builder.SetSocialLinks(a.SocialLinks)
	}
	if a.Metadata != nil {
		builder = builder.SetMetadata(a.Metadata)
	}

	_, err := builder.Save(ctx)
	return err
}

// Update cập nhật author
func (r *entAuthorRepository) Update(ctx context.Context, a *domain.Author) error {
	builder := database.GetClientFromContext(ctx, r.client).Author.UpdateOneID(a.ID).
		SetName(a.Name).
		SetSlug(a.Slug).
		SetIsVerified(a.IsVerified)

	if a.UserID != nil {
		builder = builder.SetUserID(*a.UserID)
	} else {
		builder = builder.ClearUserID()
	}
	if a.Biography != nil {
		builder = builder.SetBiography(a.Biography)
	} else {
		builder = builder.ClearBiography()
	}
	if a.AvatarURL != nil {
		builder = builder.SetAvatarURL(*a.AvatarURL)
	} else {
		builder = builder.ClearAvatarURL()
	}
	if a.SocialLinks != nil {
		builder = builder.SetSocialLinks(a.SocialLinks)
	} else {
		builder = builder.ClearSocialLinks()
	}
	if a.UpdatedBy != nil {
		builder = builder.SetUpdatedBy(*a.UpdatedBy)
	}

	_, err := builder.Save(ctx)
	return err
}

// Delete xóa mềm author
func (r *entAuthorRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return database.GetClientFromContext(ctx, r.client).Author.UpdateOneID(id).
		SetDeletedAt(now).
		Exec(ctx)
}

// GetNovelAuthors lấy danh sách authors của một novel
func (r *entAuthorRepository) GetNovelAuthors(ctx context.Context, novelID uuid.UUID) ([]*domain.NovelAuthor, error) {
	novelAuthors, err := database.GetClientFromContext(ctx, r.client).NovelAuthor.Query().
		Where(novelauthor.NovelID(novelID)).
		WithAuthor().
		Order(ent.Asc(novelauthor.FieldDisplayOrder)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.NovelAuthor, len(novelAuthors))
	for i, na := range novelAuthors {
		result[i] = &domain.NovelAuthor{
			Role:         na.Role,
			DisplayOrder: na.DisplayOrder,
		}
		if na.Edges.Author != nil {
			result[i].Author = entAuthorToDomain(na.Edges.Author)
		}
	}
	return result, nil
}

// AddNovelAuthor thêm author cho novel
func (r *entAuthorRepository) AddNovelAuthor(ctx context.Context, novelID, authorID uuid.UUID, displayOrder int) error {
	_, err := database.GetClientFromContext(ctx, r.client).NovelAuthor.Create().
		SetNovelID(novelID).
		SetAuthorID(authorID).
		SetDisplayOrder(displayOrder).
		Save(ctx)
	return err
}

// RemoveNovelAuthor xóa author khỏi novel
func (r *entAuthorRepository) RemoveNovelAuthor(ctx context.Context, novelID, authorID uuid.UUID) error {
	_, err := database.GetClientFromContext(ctx, r.client).NovelAuthor.Delete().
		Where(
			novelauthor.NovelID(novelID),
			novelauthor.AuthorID(authorID),
		).
		Exec(ctx)
	return err
}

// UpdateStatistics cập nhật thống kê
func (r *entAuthorRepository) UpdateStatistics(ctx context.Context, id uuid.UUID, stats domain.AuthorStatistics) error {
	builder := database.GetClientFromContext(ctx, r.client).Author.UpdateOneID(id)

	if stats.NovelCount != nil {
		builder = builder.SetNovelCount(*stats.NovelCount)
	}
	if stats.TotalChapters != nil {
		builder = builder.SetTotalChapters(*stats.TotalChapters)
	}
	if stats.TotalViews != nil {
		builder = builder.SetTotalViews(*stats.TotalViews)
	}
	if stats.FollowerCount != nil {
		builder = builder.SetFollowerCount(*stats.FollowerCount)
	}

	_, err := builder.Save(ctx)
	return err
}

// Merge gộp nhiều authors thành một
func (r *entAuthorRepository) Merge(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID, mergedBy uuid.UUID) error {
	// TODO: Implement với Ent transactions
	return nil
}

// GetMergePreview lấy danh sách các novel sẽ bị ảnh hưởng khi merge
func (r *entAuthorRepository) GetMergePreview(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID) ([]*domain.Novel, error) {
	// TODO: Implement
	return nil, nil
}
