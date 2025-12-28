package novel

import (
	"context"
	"encoding/json"
	"system/internal/platform/database"
	"time"

	"system/internal/domain"
	ent "system/internal/ent/generated"
	"system/internal/ent/generated/novel"
	"system/internal/ent/generated/novelartist"
	"system/internal/ent/generated/novelauthor"
	"system/internal/ent/generated/novelchapter"
	"system/internal/ent/generated/novelgenre"
	"system/internal/ent/generated/novelvolume"

	"github.com/gofrs/uuid/v5"
	"github.com/samber/lo"
)

// entNovelRepository triển khai NovelRepository sử dụng Ent
type entNovelRepository struct {
	client *ent.Client
}

// NewEntNovelRepository tạo một instance mới của entNovelRepository
func NewEntNovelRepository(client *ent.Client) domain.NovelRepository {
	return &entNovelRepository{client: client}
}

// entNovelToDomain chuyển đổi Ent Novel sang domain.Novel
func entNovelToDomain(n *ent.Novel) *domain.Novel {
	if n == nil {
		return nil
	}
	return &domain.Novel{
		ID:               n.ID,
		Title:            n.Title,
		Slug:             n.Slug,
		Synopsis:         n.Synopsis,
		CoverImageURL:    n.CoverImageURL,
		ThumbnailURL:     n.ThumbnailURL,
		Status:           domain.NovelStatus(n.Status.String()),
		IsOneshot:        n.IsOneshot,
		OriginalLanguage: n.OriginalLanguage,
		OriginalTitle:    n.OriginalTitle,
		OwnerID:          n.OwnerID,
		OwnerType:        n.OwnerType,
		TotalVolumes:     n.TotalVolumes,
		TotalChapters:    n.TotalChapters,
		TotalWords:       n.TotalWords,
		ViewCount:        n.ViewCount,
		FavoriteCount:    n.FavoriteCount,
		RatingAverage:    n.RatingAverage,
		RatingCount:      n.RatingCount,
		Metadata:         n.Metadata,
		FirstPublishedAt: n.FirstPublishedAt,
		LastChapterAt:    n.LastChapterAt,
		CompletedAt:      n.CompletedAt,
		CreatedBy:        n.CreatedBy,
		UpdatedBy:        n.UpdatedBy,
		DeletedBy:        n.DeletedBy,
		CreatedAt:        n.CreatedAt,
		UpdatedAt:        n.UpdatedAt,
		DeletedAt:        n.DeletedAt,
	}
}

// GetByID lấy novel từ database theo ID
func (r *entNovelRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Novel, error) {
	n, err := database.GetClientFromContext(ctx, r.client).Novel.Query().
		Where(novel.ID(id), novel.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return entNovelToDomain(n), nil
}

// GetBySlug lấy novel từ database theo slug
func (r *entNovelRepository) GetBySlug(ctx context.Context, slug string) (*domain.Novel, error) {
	n, err := database.GetClientFromContext(ctx, r.client).Novel.Query().
		Where(novel.Slug(slug), novel.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return entNovelToDomain(n), nil
}

// GetByAuthorID lấy danh sách novel theo author ID
func (r *entNovelRepository) GetByAuthorID(ctx context.Context, authorID uuid.UUID, limit, offset int) ([]*domain.Novel, error) {
	// Get novel IDs từ junction table
	novelAuthors, err := database.GetClientFromContext(ctx, r.client).NovelAuthor.Query().
		Where(novelauthor.AuthorID(authorID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	if len(novelAuthors) == 0 {
		return []*domain.Novel{}, nil
	}

	novelIDs := lo.Map(novelAuthors, func(na *ent.NovelAuthor, _ int) uuid.UUID {
		return na.NovelID
	})

	novels, err := database.GetClientFromContext(ctx, r.client).Novel.Query().
		Where(novel.IDIn(novelIDs...), novel.DeletedAtIsNil()).
		Limit(limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := lo.Map(novels, func(n *ent.Novel, _ int) *domain.Novel {
		return entNovelToDomain(n)
	})
	return result, nil
}

// Create tạo novel mới trong database
func (r *entNovelRepository) Create(ctx context.Context, n *domain.Novel) error {
	// Đảm bảo metadata và synopsis không null
	metadata := n.Metadata
	if metadata == nil {
		metadata = json.RawMessage("{}")
	}
	synopsis := n.Synopsis
	if synopsis == nil {
		synopsis = json.RawMessage("{}")
	}

	builder := database.GetClientFromContext(ctx, r.client).Novel.Create().
		SetID(n.ID).
		SetTitle(n.Title).
		SetSlug(n.Slug).
		SetSynopsis(synopsis).
		SetStatus(novel.Status(n.Status)).
		SetIsOneshot(n.IsOneshot).
		SetOwnerID(n.OwnerID).
		SetOwnerType(n.OwnerType).
		SetMetadata(metadata).
		SetCreatedBy(n.CreatedBy)

	if n.CoverImageURL != nil {
		builder = builder.SetCoverImageURL(*n.CoverImageURL)
	}
	if n.ThumbnailURL != nil {
		builder = builder.SetThumbnailURL(*n.ThumbnailURL)
	}
	if n.OriginalLanguage != nil {
		builder = builder.SetOriginalLanguage(*n.OriginalLanguage)
	}
	if n.OriginalTitle != nil {
		builder = builder.SetOriginalTitle(*n.OriginalTitle)
	}

	_, err := builder.Save(ctx)
	return err
}

// Update cập nhật thông tin novel
func (r *entNovelRepository) Update(ctx context.Context, n *domain.Novel) error {
	builder := database.GetClientFromContext(ctx, r.client).Novel.UpdateOneID(n.ID).
		SetTitle(n.Title).
		SetSlug(n.Slug).
		SetStatus(novel.Status(n.Status)).
		SetIsOneshot(n.IsOneshot)

	if n.Synopsis != nil {
		builder = builder.SetSynopsis(n.Synopsis)
	}
	if n.CoverImageURL != nil {
		builder = builder.SetCoverImageURL(*n.CoverImageURL)
	} else {
		builder = builder.ClearCoverImageURL()
	}
	if n.ThumbnailURL != nil {
		builder = builder.SetThumbnailURL(*n.ThumbnailURL)
	} else {
		builder = builder.ClearThumbnailURL()
	}
	if n.OriginalLanguage != nil {
		builder = builder.SetOriginalLanguage(*n.OriginalLanguage)
	} else {
		builder = builder.ClearOriginalLanguage()
	}
	if n.OriginalTitle != nil {
		builder = builder.SetOriginalTitle(*n.OriginalTitle)
	} else {
		builder = builder.ClearOriginalTitle()
	}
	if n.Metadata != nil {
		builder = builder.SetMetadata(n.Metadata)
	}
	if n.UpdatedBy != nil {
		builder = builder.SetUpdatedBy(*n.UpdatedBy)
	}
	if n.FirstPublishedAt != nil {
		builder = builder.SetFirstPublishedAt(*n.FirstPublishedAt)
	}
	if n.CompletedAt != nil {
		builder = builder.SetCompletedAt(*n.CompletedAt)
	}

	_, err := builder.Save(ctx)
	return err
}

// Delete xóa mềm novel
func (r *entNovelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return database.GetClientFromContext(ctx, r.client).Novel.UpdateOneID(id).
		SetDeletedAt(now).
		Exec(ctx)
}

// List lấy danh sách novel với filter và pagination
func (r *entNovelRepository) List(ctx context.Context, filter domain.NovelFilter) ([]*domain.Novel, int64, error) {
	query := database.GetClientFromContext(ctx, r.client).Novel.Query().Where(novel.DeletedAtIsNil())

	// Filter by IDs
	if len(filter.IDs) > 0 {
		query = query.Where(novel.IDIn(filter.IDs...))
	}

	// Filter by owner
	if filter.OwnerID != nil {
		query = query.Where(novel.OwnerID(*filter.OwnerID))
	}

	// Filter by statuses
	if len(filter.Statuses) > 0 {
		statuses := make([]novel.Status, len(filter.Statuses))
		for i, s := range filter.Statuses {
			statuses[i] = novel.Status(s)
		}
		query = query.Where(novel.StatusIn(statuses...))
	}

	// Filter by original language
	if filter.OriginalLanguage != nil {
		query = query.Where(novel.OriginalLanguage(*filter.OriginalLanguage))
	}

	// Search query
	if filter.SearchQuery != nil && *filter.SearchQuery != "" {
		query = query.Where(novel.Or(
			novel.TitleContainsFold(*filter.SearchQuery),
			novel.OriginalTitleContainsFold(*filter.SearchQuery),
		))
	}

	// Count total
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Sort
	orderFunc := ent.Desc(novel.FieldCreatedAt)
	if filter.SortBy != "" {
		orderField := novel.FieldCreatedAt
		switch filter.SortBy {
		case "rating":
			orderField = novel.FieldRatingAverage
		case "views":
			orderField = novel.FieldViewCount
		case "last_chapter":
			orderField = novel.FieldLastChapterAt
		case "created_at":
			orderField = novel.FieldCreatedAt
		}
		if filter.SortOrder == "asc" {
			orderFunc = ent.Asc(orderField)
		} else {
			orderFunc = ent.Desc(orderField)
		}
	}

	novels, err := query.
		Order(orderFunc).
		Limit(filter.Limit).
		Offset(filter.Offset).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := lo.Map(novels, func(n *ent.Novel, _ int) *domain.Novel {
		return entNovelToDomain(n)
	})
	return result, int64(total), nil
}

// UpdateStatistics cập nhật thống kê của novel
func (r *entNovelRepository) UpdateStatistics(ctx context.Context, id uuid.UUID, stats domain.NovelStatistics) error {
	builder := database.GetClientFromContext(ctx, r.client).Novel.UpdateOneID(id)

	if stats.ViewCount != nil {
		builder = builder.SetViewCount(*stats.ViewCount)
	}
	if stats.FavoriteCount != nil {
		builder = builder.SetFavoriteCount(*stats.FavoriteCount)
	}
	if stats.RatingAverage != nil {
		builder = builder.SetRatingAverage(*stats.RatingAverage)
	}
	if stats.RatingCount != nil {
		builder = builder.SetRatingCount(*stats.RatingCount)
	}

	_, err := builder.Save(ctx)
	return err
}

// IncrementViewCount tăng view count
func (r *entNovelRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	return database.GetClientFromContext(ctx, r.client).Novel.UpdateOneID(id).
		AddViewCount(1).
		Exec(ctx)
}

// BatchIncrementViewCount tăng view count cho nhiều novels
func (r *entNovelRepository) BatchIncrementViewCount(ctx context.Context, increments map[uuid.UUID]int64) error {
	if len(increments) == 0 {
		return nil
	}

	for id, count := range increments {
		_, err := database.GetClientFromContext(ctx, r.client).Novel.UpdateOneID(id).
			AddViewCount(count).
			Save(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetAuthors lấy danh sách author IDs của novel
func (r *entNovelRepository) GetAuthors(ctx context.Context, novelID uuid.UUID) ([]uuid.UUID, error) {
	novelAuthors, err := database.GetClientFromContext(ctx, r.client).NovelAuthor.Query().
		Where(novelauthor.NovelID(novelID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := lo.Map(novelAuthors, func(na *ent.NovelAuthor, _ int) uuid.UUID {
		return na.AuthorID
	})
	return result, nil
}

// GetGenres lấy danh sách genre IDs của novel
func (r *entNovelRepository) GetGenres(ctx context.Context, novelID uuid.UUID) ([]uuid.UUID, error) {
	novelGenres, err := database.GetClientFromContext(ctx, r.client).NovelGenre.Query().
		Where(novelgenre.NovelID(novelID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := lo.Map(novelGenres, func(ng *ent.NovelGenre, _ int) uuid.UUID {
		return ng.GenreID
	})
	return result, nil
}

// GetArtists lấy danh sách artist IDs của novel
func (r *entNovelRepository) GetArtists(ctx context.Context, novelID uuid.UUID) ([]uuid.UUID, error) {
	novelArtists, err := database.GetClientFromContext(ctx, r.client).NovelArtist.Query().
		Where(novelartist.NovelID(novelID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := lo.Map(novelArtists, func(na *ent.NovelArtist, _ int) uuid.UUID {
		return na.ArtistID
	})
	return result, nil
}

// GetOrganizationAssignments lấy danh sách organization IDs được assign cho novel
// TODO: Cần tạo junction table NovelOrganization
func (r *entNovelRepository) GetOrganizationAssignments(ctx context.Context, novelID uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

// GetNovelFullBySlug lấy toàn bộ dữ liệu novel trong một transaction
func (r *entNovelRepository) GetNovelFullBySlug(ctx context.Context, slug string) (*NovelFullData, error) {
	client := database.GetClientFromContext(ctx, r.client)

	// 1. Get novel by slug
	n, err := client.Novel.Query().
		Where(novel.Slug(slug), novel.DeletedAtIsNil()).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, err
		}
		return nil, err
	}

	novelDomain := entNovelToDomain(n)

	// 2. Get genres
	novelGenres, err := client.NovelGenre.Query().
		Where(novelgenre.NovelID(n.ID)).
		WithGenre().
		All(ctx)
	if err != nil {
		return nil, err
	}
	genres := make([]*domain.Genre, 0, len(novelGenres))
	for _, ng := range novelGenres {
		if ng.Edges.Genre != nil {
			g := ng.Edges.Genre
			genres = append(genres, &domain.Genre{
				ID:   g.ID,
				Name: g.Name,
				Slug: g.Slug,
			})
		}
	}

	// 3. Get authors
	novelAuthors, err := client.NovelAuthor.Query().
		Where(novelauthor.NovelID(n.ID)).
		WithAuthor().
		Order(ent.Asc(novelauthor.FieldDisplayOrder)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	authors := make([]*domain.NovelAuthor, 0, len(novelAuthors))
	for _, na := range novelAuthors {
		author := &domain.NovelAuthor{
			DisplayOrder: na.DisplayOrder,
		}
		if na.Edges.Author != nil {
			a := na.Edges.Author
			author.Author = &domain.Author{
				ID:   a.ID,
				Name: a.Name,
				Slug: a.Slug,
			}
		}
		authors = append(authors, author)
	}

	// 4. Get artists
	novelArtists, err := client.NovelArtist.Query().
		Where(novelartist.NovelID(n.ID)).
		WithArtist().
		Order(ent.Asc(novelartist.FieldDisplayOrder)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	artists := make([]*domain.NovelArtist, 0, len(novelArtists))
	for _, na := range novelArtists {
		artist := &domain.NovelArtist{
			DisplayOrder: na.DisplayOrder,
		}
		if na.Edges.Artist != nil {
			a := na.Edges.Artist
			artist.Artist = &domain.Artist{
				ID:   a.ID,
				Name: a.Name,
				Slug: a.Slug,
			}
		}
		artists = append(artists, artist)
	}

	// 5. Get published volumes
	entVolumes, err := client.NovelVolume.Query().
		Where(
			novelvolume.NovelID(n.ID),
			novelvolume.IsPublished(true),
			novelvolume.DeletedAtIsNil(),
		).
		Order(ent.Asc(novelvolume.FieldDisplayOrder)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// 6. Get published chapters
	entChapters, err := client.NovelChapter.Query().
		Where(
			novelchapter.NovelID(n.ID),
			novelchapter.StatusEQ(novelchapter.StatusPublished),
			novelchapter.DeletedAtIsNil(),
		).
		Order(ent.Asc(novelchapter.FieldDisplayOrder)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	// Build volume ID map
	volumeMap := make(map[uuid.UUID]*domain.NovelVolumeWithChapters)
	volumes := make([]*domain.NovelVolume, 0, len(entVolumes))
	volumesWithChapters := make([]*domain.NovelVolumeWithChapters, 0, len(entVolumes))

	for _, v := range entVolumes {
		vol := &domain.NovelVolume{
			ID:            v.ID,
			NovelID:       v.NovelID,
			VolumeNumber:  v.VolumeNumber,
			Title:         v.Title,
			Slug:          v.Slug,
			CoverImageURL: v.CoverImageURL,
			DisplayOrder:  v.DisplayOrder,
			IsPublished:   v.IsPublished,
			PublishedAt:   v.PublishedAt,
			CreatedAt:     v.CreatedAt,
			UpdatedAt:     v.UpdatedAt,
		}
		volumes = append(volumes, vol)

		vwc := &domain.NovelVolumeWithChapters{
			Volume:   vol,
			Chapters: make([]*domain.NovelChapter, 0),
		}
		volumeMap[v.ID] = vwc
		volumesWithChapters = append(volumesWithChapters, vwc)
	}

	// Map chapters to volumes or chaptersWithoutVol
	chaptersWithoutVol := make([]*domain.NovelChapter, 0)

	for _, c := range entChapters {
		ch := &domain.NovelChapter{
			ID:            c.ID,
			NovelID:       c.NovelID,
			VolumeID:      c.VolumeID,
			ChapterNumber: c.ChapterNumber,
			Title:         c.Title,
			Slug:          c.Slug,
			DisplayOrder:  c.DisplayOrder,
			Status:        domain.NovelChapterStatus(c.Status.String()),
			PublishedAt:   c.PublishedAt,
			CreatedAt:     c.CreatedAt,
			UpdatedAt:     c.UpdatedAt,
		}

		if c.VolumeID != nil {
			if vwc, ok := volumeMap[*c.VolumeID]; ok {
				vwc.Chapters = append(vwc.Chapters, ch)
			} else {
				chaptersWithoutVol = append(chaptersWithoutVol, ch)
			}
		} else {
			chaptersWithoutVol = append(chaptersWithoutVol, ch)
		}
	}

	// Owner info will be set from OwnerID if needed by the handler
	// The domain.Novel already has OwnerID, OwnerType from entNovelToDomain

	return &NovelFullData{
		Novel:               novelDomain,
		Genres:              genres,
		Authors:             authors,
		Artists:             artists,
		Volumes:             volumes,
		VolumesWithChapters: volumesWithChapters,
		ChaptersWithoutVol:  chaptersWithoutVol,
	}, nil
}

// UpdateContentStatistics recalculates and updates content statistics
// (total_volumes, total_chapters, total_words) based on published content
func (r *entNovelRepository) UpdateContentStatistics(ctx context.Context, novelID uuid.UUID) error {
	client := database.GetClientFromContext(ctx, r.client)

	// Count published volumes
	totalVolumes, err := client.NovelVolume.Query().
		Where(
			novelvolume.NovelID(novelID),
			novelvolume.IsPublished(true),
			novelvolume.DeletedAtIsNil(),
		).
		Count(ctx)
	if err != nil {
		return err
	}

	// Count published chapters and sum word counts
	chapters, err := client.NovelChapter.Query().
		Where(
			novelchapter.NovelID(novelID),
			novelchapter.StatusEQ(novelchapter.StatusPublished),
			novelchapter.DeletedAtIsNil(),
		).
		All(ctx)
	if err != nil {
		return err
	}

	totalChapters := len(chapters)
	totalWords := lo.SumBy(chapters, func(ch *ent.NovelChapter) int64 {
		return int64(ch.WordCount)
	})

	// Update novel with calculated stats
	_, err = client.Novel.UpdateOneID(novelID).
		SetTotalVolumes(totalVolumes).
		SetTotalChapters(totalChapters).
		SetTotalWords(totalWords).
		Save(ctx)

	return err
}
