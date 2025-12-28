package genre

import (
	"context"
	"system/internal/platform/database"
	"time"

	"system/internal/domain"
	ent "system/internal/ent/generated"
	"system/internal/ent/generated/genre"
	"system/internal/ent/generated/novelgenre"

	"github.com/gofrs/uuid/v5"
	"github.com/samber/lo"
)

// entGenreRepository triển khai GenreRepository sử dụng Ent
type entGenreRepository struct {
	client *ent.Client
}

// NewEntGenreRepository tạo một instance mới của entGenreRepository
func NewEntGenreRepository(client *ent.Client) domain.GenreRepository {
	return &entGenreRepository{client: client}
}

// entGenreToDomain chuyển đổi Ent Genre sang domain.Genre
func entGenreToDomain(g *ent.Genre) *domain.Genre {
	if g == nil {
		return nil
	}
	return &domain.Genre{
		ID:            g.ID,
		Name:          g.Name,
		Slug:          g.Slug,
		Description:   g.Description,
		ParentID:      g.ParentID,
		IsActive:      g.IsActive,
		NovelCount:    g.NovelCount,
		AnimeCount:    g.AnimeCount,
		MangaCount:    g.MangaCount,
		ActiveReaders: g.ActiveReaders,
		TotalViews:    g.TotalViews,
		CreatedBy:     g.CreatedBy,
		UpdatedBy:     g.UpdatedBy,
		DeletedBy:     g.DeletedBy,
		Version:       g.Version,
		CreatedAt:     g.CreatedAt,
		UpdatedAt:     g.UpdatedAt,
		DeletedAt:     g.DeletedAt,
	}
}

// GetByID lấy genre từ database theo ID
func (r *entGenreRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Genre, error) {
	g, err := database.GetClientFromContext(ctx, r.client).Genre.
		Query().
		Where(
			genre.ID(id),
			genre.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return entGenreToDomain(g), nil
}

// GetBySlug lấy genre từ database theo slug
func (r *entGenreRepository) GetBySlug(ctx context.Context, slug string) (*domain.Genre, error) {
	g, err := database.GetClientFromContext(ctx, r.client).Genre.
		Query().
		Where(
			genre.Slug(slug),
			genre.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		return nil, err
	}
	return entGenreToDomain(g), nil
}

// GetAll lấy tất cả genres
func (r *entGenreRepository) GetAll(ctx context.Context, activeOnly bool) ([]*domain.Genre, error) {
	query := database.GetClientFromContext(ctx, r.client).Genre.Query().Where(genre.DeletedAtIsNil())

	if activeOnly {
		query = query.Where(genre.IsActive(true))
	}

	genres, err := query.Order(ent.Asc(genre.FieldName)).All(ctx)
	if err != nil {
		return nil, err
	}

	result := lo.Map(genres, func(g *ent.Genre, _ int) *domain.Genre {
		return entGenreToDomain(g)
	})
	return result, nil
}

// List lấy danh sách genres với pagination, search và sort
func (r *entGenreRepository) List(ctx context.Context, offset, limit int, search, sortBy, sortOrder string, activeOnly bool) ([]*domain.Genre, int, error) {
	query := database.GetClientFromContext(ctx, r.client).Genre.Query().Where(genre.DeletedAtIsNil())

	if activeOnly {
		query = query.Where(genre.IsActive(true))
	}

	if search != "" {
		query = query.Where(genre.NameContainsFold(search))
	}

	// Đếm tổng
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Sắp xếp
	orderFunc := ent.Asc(genre.FieldName) // Default
	if sortBy != "" {
		orderField := genre.FieldName
		switch sortBy {
		case "name":
			orderField = genre.FieldName
		case "views":
			orderField = genre.FieldTotalViews
		case "created":
			orderField = genre.FieldCreatedAt
		case "updated":
			orderField = genre.FieldUpdatedAt
		case "readers":
			orderField = genre.FieldActiveReaders
		}
		if sortOrder == "desc" {
			orderFunc = ent.Desc(orderField)
		} else {
			orderFunc = ent.Asc(orderField)
		}
	}

	genres, err := query.
		Order(orderFunc).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := lo.Map(genres, func(g *ent.Genre, _ int) *domain.Genre {
		return entGenreToDomain(g)
	})
	return result, total, nil
}

// ListSelection lấy danh sách genres rút gọn (chỉ ID và Name)
func (r *entGenreRepository) ListSelection(ctx context.Context, offset, limit int, search string) ([]*domain.Genre, int, error) {
	query := database.GetClientFromContext(ctx, r.client).Genre.Query().
		Where(genre.DeletedAtIsNil(), genre.IsActive(true))

	if search != "" {
		query = query.Where(genre.NameContainsFold(search))
	}

	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	genres, err := query.
		Select(genre.FieldID, genre.FieldName).
		Order(ent.Asc(genre.FieldName)).
		Offset(offset).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := lo.Map(genres, func(g *ent.Genre, _ int) *domain.Genre {
		return &domain.Genre{ID: g.ID, Name: g.Name}
	})
	return result, total, nil
}

// GetByParentID lấy các genre con theo parent ID
func (r *entGenreRepository) GetByParentID(ctx context.Context, parentID uuid.UUID) ([]*domain.Genre, error) {
	genres, err := database.GetClientFromContext(ctx, r.client).Genre.Query().
		Where(
			genre.ParentID(parentID),
			genre.DeletedAtIsNil(),
		).
		Order(ent.Asc(genre.FieldName)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := lo.Map(genres, func(g *ent.Genre, _ int) *domain.Genre {
		return entGenreToDomain(g)
	})
	return result, nil
}

// GetRootGenres lấy các genre gốc (không có parent)
func (r *entGenreRepository) GetRootGenres(ctx context.Context) ([]*domain.Genre, error) {
	genres, err := database.GetClientFromContext(ctx, r.client).Genre.Query().
		Where(
			genre.ParentIDIsNil(),
			genre.DeletedAtIsNil(),
		).
		Order(ent.Asc(genre.FieldName)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := lo.Map(genres, func(g *ent.Genre, _ int) *domain.Genre {
		return entGenreToDomain(g)
	})
	return result, nil
}

// Create tạo genre mới
func (r *entGenreRepository) Create(ctx context.Context, g *domain.Genre) error {
	builder := database.GetClientFromContext(ctx, r.client).Genre.Create().
		SetID(g.ID).
		SetName(g.Name).
		SetSlug(g.Slug).
		SetIsActive(g.IsActive)

	if g.Description != nil {
		builder = builder.SetDescription(*g.Description)
	}
	if g.ParentID != nil {
		builder = builder.SetParentID(*g.ParentID)
	}
	if g.CreatedBy != nil {
		builder = builder.SetCreatedBy(*g.CreatedBy)
	}

	_, err := builder.Save(ctx)
	return err
}

// Update cập nhật genre
func (r *entGenreRepository) Update(ctx context.Context, g *domain.Genre) error {
	builder := database.GetClientFromContext(ctx, r.client).Genre.UpdateOneID(g.ID).
		SetName(g.Name).
		SetSlug(g.Slug).
		SetIsActive(g.IsActive)

	if g.Description != nil {
		builder = builder.SetDescription(*g.Description)
	} else {
		builder = builder.ClearDescription()
	}

	if g.ParentID != nil {
		builder = builder.SetParentID(*g.ParentID)
	} else {
		builder = builder.ClearParentID()
	}

	if g.UpdatedBy != nil {
		builder = builder.SetUpdatedBy(*g.UpdatedBy)
	}

	_, err := builder.Save(ctx)
	return err
}

// Delete xóa genre (soft delete)
func (r *entGenreRepository) Delete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	now := time.Now()
	return database.GetClientFromContext(ctx, r.client).Genre.UpdateOneID(id).
		SetDeletedAt(now).
		SetDeletedBy(deletedBy).
		Exec(ctx)
}

// GetNovelGenres lấy danh sách genres của một novel
func (r *entGenreRepository) GetNovelGenres(ctx context.Context, novelID uuid.UUID) ([]*domain.Genre, error) {
	novelGenres, err := database.GetClientFromContext(ctx, r.client).NovelGenre.Query().
		Where(novelgenre.NovelID(novelID)).
		WithGenre().
		Order(ent.Asc(novelgenre.FieldDisplayOrder)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := lo.FilterMap(novelGenres, func(ng *ent.NovelGenre, _ int) (*domain.Genre, bool) {
		if ng.Edges.Genre != nil {
			return entGenreToDomain(ng.Edges.Genre), true
		}
		return nil, false
	})
	return result, nil
}

// AddNovelGenre thêm genre cho novel
func (r *entGenreRepository) AddNovelGenre(ctx context.Context, novelID, genreID, createdBy uuid.UUID) error {
	_, err := database.GetClientFromContext(ctx, r.client).NovelGenre.Create().
		SetNovelID(novelID).
		SetGenreID(genreID).
		SetCreatedBy(createdBy).
		Save(ctx)
	return err
}

// AddNovelGenres thêm nhiều genres cho novel (Batch Insert)
func (r *entGenreRepository) AddNovelGenres(ctx context.Context, novelID uuid.UUID, genreIDs []uuid.UUID, createdBy uuid.UUID) error {
	if len(genreIDs) == 0 {
		return nil
	}

	builders := make([]*ent.NovelGenreCreate, len(genreIDs))
	for i, genreID := range genreIDs {
		builders[i] = database.GetClientFromContext(ctx, r.client).NovelGenre.Create().
			SetNovelID(novelID).
			SetGenreID(genreID).
			SetDisplayOrder(i).
			SetCreatedBy(createdBy)
	}

	_, err := database.GetClientFromContext(ctx, r.client).NovelGenre.CreateBulk(builders...).Save(ctx)
	return err
}

// RemoveNovelGenre xóa genre khỏi novel
func (r *entGenreRepository) RemoveNovelGenre(ctx context.Context, novelID, genreID uuid.UUID) error {
	_, err := database.GetClientFromContext(ctx, r.client).NovelGenre.Delete().
		Where(
			novelgenre.NovelID(novelID),
			novelgenre.GenreID(genreID),
		).
		Exec(ctx)
	return err
}

// UpdateNovelGenres cập nhật toàn bộ genres của novel
func (r *entGenreRepository) UpdateNovelGenres(ctx context.Context, novelID uuid.UUID, genreIDs []uuid.UUID, createdBy uuid.UUID) error {
	// Xóa tất cả genres hiện tại
	_, err := database.GetClientFromContext(ctx, r.client).NovelGenre.Delete().
		Where(novelgenre.NovelID(novelID)).
		Exec(ctx)
	if err != nil {
		return err
	}

	// Thêm genres mới
	return r.AddNovelGenres(ctx, novelID, genreIDs, createdBy)
}

// BatchIncrementNovelCount tăng số lượng novel cho nhiều genres
func (r *entGenreRepository) BatchIncrementNovelCount(ctx context.Context, increments map[uuid.UUID]int) error {
	if len(increments) == 0 {
		return nil
	}

	for genreID, count := range increments {
		_, err := database.GetClientFromContext(ctx, r.client).Genre.UpdateOneID(genreID).
			AddNovelCount(count).
			Save(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// BatchIncrementTotalViews tăng total views cho nhiều genres
func (r *entGenreRepository) BatchIncrementTotalViews(ctx context.Context, increments map[uuid.UUID]int64) error {
	if len(increments) == 0 {
		return nil
	}

	for genreID, count := range increments {
		_, err := database.GetClientFromContext(ctx, r.client).Genre.UpdateOneID(genreID).
			AddTotalViews(count).
			Save(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetGenresByNovelIDs lấy genre IDs cho danh sách novel IDs
func (r *entGenreRepository) GetGenresByNovelIDs(ctx context.Context, novelIDs []uuid.UUID) (map[uuid.UUID][]uuid.UUID, error) {
	if len(novelIDs) == 0 {
		return nil, nil
	}

	novelGenres, err := database.GetClientFromContext(ctx, r.client).NovelGenre.Query().
		Where(novelgenre.NovelIDIn(novelIDs...)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := lo.GroupBy(novelGenres, func(ng *ent.NovelGenre) uuid.UUID {
		return ng.NovelID
	})
	// Convert to map[uuid.UUID][]uuid.UUID
	finalResult := make(map[uuid.UUID][]uuid.UUID, len(result))
	for novelID, ngs := range result {
		finalResult[novelID] = lo.Map(ngs, func(ng *ent.NovelGenre, _ int) uuid.UUID {
			return ng.GenreID
		})
	}
	return finalResult, nil
}

// BatchUpdateActiveReaders cập nhật số lượng active readers cho nhiều genres
func (r *entGenreRepository) BatchUpdateActiveReaders(ctx context.Context, updates map[uuid.UUID]int64) error {
	if len(updates) == 0 {
		return nil
	}

	for genreID, count := range updates {
		_, err := database.GetClientFromContext(ctx, r.client).Genre.UpdateOneID(genreID).
			SetActiveReaders(count).
			Save(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// Merge gộp nhiều genres (sources) thành một genre (target)
// TODO: Implement với Ent transactions
func (r *entGenreRepository) Merge(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID, mergedBy uuid.UUID) error {
	// TODO: Cần implement với Ent transactions
	return nil
}

// GetMergePreview lấy danh sách các novel sẽ bị ảnh hưởng khi merge
// TODO: Implement với Ent queries
func (r *entGenreRepository) GetMergePreview(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID) ([]*domain.AffectedNovel, error) {
	return nil, nil
}
