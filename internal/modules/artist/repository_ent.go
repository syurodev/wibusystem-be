package artist

import (
	"context"
	"system/internal/platform/database"
	"time"

	"system/internal/domain"
	ent "system/internal/ent/generated"
	"system/internal/ent/generated/artist"
	"system/internal/ent/generated/novelartist"

	"github.com/gofrs/uuid/v5"
)

// entArtistRepository triển khai ArtistRepository sử dụng Ent
type entArtistRepository struct {
	client *ent.Client
}

// NewEntArtistRepository tạo một instance mới của entArtistRepository
func NewEntArtistRepository(client *ent.Client) domain.ArtistRepository {
	return &entArtistRepository{client: client}
}

// entArtistToDomain chuyển đổi Ent Artist sang domain.Artist
func entArtistToDomain(a *ent.Artist) *domain.Artist {
	if a == nil {
		return nil
	}
	return &domain.Artist{
		ID:             a.ID,
		UserID:         a.UserID,
		Name:           a.Name,
		Slug:           a.Slug,
		Biography:      a.Biography,
		AvatarURL:      a.AvatarURL,
		SocialLinks:    a.SocialLinks,
		Specialization: a.Specialization,
		PortfolioURL:   a.PortfolioURL,
		NovelCount:     a.NovelCount,
		ArtworkCount:   a.ArtworkCount,
		FollowerCount:  a.FollowerCount,
		IsVerified:     a.IsVerified,
		Metadata:       a.Metadata,
		CreatedBy:      a.CreatedBy,
		UpdatedBy:      a.UpdatedBy,
		DeletedBy:      a.DeletedBy,
		Version:        a.Version,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
		DeletedAt:      a.DeletedAt,
	}
}

// GetByID lấy artist từ database theo ID
func (r *entArtistRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Artist, error) {
	a, err := database.GetClientFromContext(ctx, r.client).Artist.Query().
		Where(artist.ID(id), artist.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return entArtistToDomain(a), nil
}

// GetBySlug lấy artist từ database theo slug
func (r *entArtistRepository) GetBySlug(ctx context.Context, slug string) (*domain.Artist, error) {
	a, err := database.GetClientFromContext(ctx, r.client).Artist.Query().
		Where(artist.Slug(slug), artist.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return entArtistToDomain(a), nil
}

// GetByUserID lấy artist từ database theo user ID
func (r *entArtistRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Artist, error) {
	a, err := database.GetClientFromContext(ctx, r.client).Artist.Query().
		Where(artist.UserID(userID), artist.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return entArtistToDomain(a), nil
}

// List lấy danh sách artists với filter
func (r *entArtistRepository) List(ctx context.Context, filter domain.ArtistFilter) ([]*domain.Artist, int64, error) {
	query := database.GetClientFromContext(ctx, r.client).Artist.Query().Where(artist.DeletedAtIsNil())

	if filter.SearchQuery != nil && *filter.SearchQuery != "" {
		query = query.Where(artist.NameContainsFold(*filter.SearchQuery))
	}
	if filter.IsVerified != nil {
		query = query.Where(artist.IsVerified(*filter.IsVerified))
	}
	if filter.Specialization != nil && *filter.Specialization != "" {
		query = query.Where(artist.Specialization(*filter.Specialization))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	orderFunc := ent.Asc(artist.FieldCreatedAt)
	if filter.SortBy != "" {
		orderField := artist.FieldCreatedAt
		switch filter.SortBy {
		case "name":
			orderField = artist.FieldName
		case "novel_count":
			orderField = artist.FieldNovelCount
		case "created_at":
			orderField = artist.FieldCreatedAt
		}
		if filter.SortOrder == "desc" {
			orderFunc = ent.Desc(orderField)
		} else {
			orderFunc = ent.Asc(orderField)
		}
	}

	artists, err := query.
		Order(orderFunc).
		Offset(filter.Offset).
		Limit(filter.Limit).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*domain.Artist, len(artists))
	for i, a := range artists {
		result[i] = entArtistToDomain(a)
	}
	return result, int64(total), nil
}

// ListSelection lấy danh sách artists rút gọn (chỉ ID và Name)
func (r *entArtistRepository) ListSelection(ctx context.Context, offset, limit int, search string) ([]*domain.Artist, int64, error) {
	query := database.GetClientFromContext(ctx, r.client).Artist.Query().Where(artist.DeletedAtIsNil())

	if search != "" {
		query = query.Where(artist.NameContainsFold(search))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	artists, err := query.
		Select(artist.FieldID, artist.FieldName).
		Order(ent.Asc(artist.FieldName)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*domain.Artist, len(artists))
	for i, a := range artists {
		result[i] = &domain.Artist{
			ID:   a.ID,
			Name: a.Name,
		}
	}
	return result, int64(total), nil
}

// Create tạo artist mới
func (r *entArtistRepository) Create(ctx context.Context, a *domain.Artist) error {
	builder := database.GetClientFromContext(ctx, r.client).Artist.Create().
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
	if a.Specialization != nil {
		builder = builder.SetSpecialization(*a.Specialization)
	}
	if a.PortfolioURL != nil {
		builder = builder.SetPortfolioURL(*a.PortfolioURL)
	}
	if a.Metadata != nil {
		builder = builder.SetMetadata(a.Metadata)
	}

	_, err := builder.Save(ctx)
	return err
}

// Update cập nhật artist
func (r *entArtistRepository) Update(ctx context.Context, a *domain.Artist) error {
	builder := database.GetClientFromContext(ctx, r.client).Artist.UpdateOneID(a.ID).
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
	if a.Specialization != nil {
		builder = builder.SetSpecialization(*a.Specialization)
	} else {
		builder = builder.ClearSpecialization()
	}
	if a.UpdatedBy != nil {
		builder = builder.SetUpdatedBy(*a.UpdatedBy)
	}

	_, err := builder.Save(ctx)
	return err
}

// Delete xóa mềm artist
func (r *entArtistRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return database.GetClientFromContext(ctx, r.client).Artist.UpdateOneID(id).
		SetDeletedAt(now).
		Exec(ctx)
}

// GetNovelArtists lấy danh sách artists của một novel
func (r *entArtistRepository) GetNovelArtists(ctx context.Context, novelID uuid.UUID) ([]*domain.NovelArtist, error) {
	novelArtists, err := database.GetClientFromContext(ctx, r.client).NovelArtist.Query().
		Where(novelartist.NovelID(novelID)).
		WithArtist().
		Order(ent.Asc(novelartist.FieldDisplayOrder)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.NovelArtist, len(novelArtists))
	for i, na := range novelArtists {
		result[i] = &domain.NovelArtist{
			Role:         na.Role,
			DisplayOrder: na.DisplayOrder,
		}
		if na.Edges.Artist != nil {
			result[i].Artist = entArtistToDomain(na.Edges.Artist)
		}
	}
	return result, nil
}

// AddNovelArtist thêm artist cho novel
func (r *entArtistRepository) AddNovelArtist(ctx context.Context, novelID, artistID uuid.UUID, displayOrder int) error {
	_, err := database.GetClientFromContext(ctx, r.client).NovelArtist.Create().
		SetNovelID(novelID).
		SetArtistID(artistID).
		SetDisplayOrder(displayOrder).
		Save(ctx)
	return err
}

// RemoveNovelArtist xóa artist khỏi novel
func (r *entArtistRepository) RemoveNovelArtist(ctx context.Context, novelID, artistID uuid.UUID, role string) error {
	_, err := database.GetClientFromContext(ctx, r.client).NovelArtist.Delete().
		Where(
			novelartist.NovelID(novelID),
			novelartist.ArtistID(artistID),
		).
		Exec(ctx)
	return err
}

// UpdateStatistics cập nhật thống kê
func (r *entArtistRepository) UpdateStatistics(ctx context.Context, id uuid.UUID, stats domain.ArtistStatistics) error {
	builder := database.GetClientFromContext(ctx, r.client).Artist.UpdateOneID(id)

	if stats.NovelCount != nil {
		builder = builder.SetNovelCount(*stats.NovelCount)
	}
	if stats.ArtworkCount != nil {
		builder = builder.SetArtworkCount(*stats.ArtworkCount)
	}
	if stats.FollowerCount != nil {
		builder = builder.SetFollowerCount(*stats.FollowerCount)
	}

	_, err := builder.Save(ctx)
	return err
}

// Merge gộp nhiều artists thành một
func (r *entArtistRepository) Merge(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID, mergedBy uuid.UUID) error {
	// TODO: Implement với Ent transactions
	return nil
}

// GetMergePreview lấy danh sách các novel sẽ bị ảnh hưởng khi merge
func (r *entArtistRepository) GetMergePreview(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID) ([]*domain.Novel, error) {
	// TODO: Implement
	return nil, nil
}
