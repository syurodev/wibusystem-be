// ============================================================================
// Novel Service
// ============================================================================
//
// Service này cung cấp business logic cho Novel module.
// Novel là cấp cao nhất trong cấu trúc phân cấp: Novel > Volume > Chapter.
//
// CRUD Operations:
//   - CreateNovel: Tạo novel mới với relations (genres, authors, artists)
//   - UpdateNovel: Cập nhật thông tin novel
//   - DeleteNovel: Soft delete novel và update related stats
//   - GetNovelByID/GetNovelBySlug: Lấy chi tiết novel
//
// List Operations:
//   - ListNovels: List với filter, search, sort, pagination
//   - GetNovelsByIDs: Batch get novels theo list IDs
//
// Relations:
//   - GetNovelGenres/Authors/Artists: Lấy IDs của relations
//   - GetNovelGenresDetails/AuthorsDetails/ArtistsDetails: Lấy chi tiết relations
//
// Advanced:
//   - GetNovelFull: Lấy toàn bộ data (novel + volumes + chapters) cho detail page
//   - IncrementViewCount: Tăng view count
//
// ============================================================================

package novel

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/gosimple/slug"
	"github.com/samber/lo"

	"system/internal/domain"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/stringutil"
)

// novelServiceImpl implements NovelService interface
type novelServiceImpl struct {
	novelRepo   domain.NovelRepository
	volumeRepo  domain.NovelVolumeRepository
	genreRepo   domain.GenreRepository
	authorRepo  domain.AuthorRepository
	artistRepo  domain.ArtistRepository
	creatorRepo domain.CreatorRepository
}

// NewService tạo một instance mới của NovelService
func NewService(
	novelRepo domain.NovelRepository,
	volumeRepo domain.NovelVolumeRepository,
	genreRepo domain.GenreRepository,
	authorRepo domain.AuthorRepository,
	artistRepo domain.ArtistRepository,
	creatorRepo domain.CreatorRepository,
) NovelService {
	return &novelServiceImpl{
		novelRepo:   novelRepo,
		volumeRepo:  volumeRepo,
		genreRepo:   genreRepo,
		authorRepo:  authorRepo,
		artistRepo:  artistRepo,
		creatorRepo: creatorRepo,
	}
}

// CreateNovel tạo novel mới
func (s *novelServiceImpl) CreateNovel(
	ctx context.Context,
	title string,
	synopsis json.RawMessage,
	coverImageURL, thumbnailURL *string,
	status, originalLanguage, originalTitle *string,
	metadataJSON *string,
	isOneshot bool,
	ownerID uuid.UUID,
	ownerType string,
	createdBy uuid.UUID,
	genreIDs []uuid.UUID,
	authorIDs []uuid.UUID,
	artistIDs []uuid.UUID,
) (*domain.Novel, error) {
	if title == "" {
		return nil, pkgerrors.BadRequest(I18nInvalidInput, "title is required")
	}

	if status == nil || !domain.NovelStatus(*status).IsValid() {
		return nil, pkgerrors.BadRequest(I18nInvalidStatus, "invalid novel status")
	}

	novelSlug, err := stringutil.GenerateUniqueSlug(title)
	if err != nil {
		return nil, err
	}

	if len(synopsis) == 0 || string(synopsis) == "null" {
		synopsis = json.RawMessage("{}")
	} else {
		if !json.Valid(synopsis) {
			return nil, pkgerrors.BadRequest(I18nInvalidInput, "invalid synopsis JSON")
		}
	}

	var metadata json.RawMessage
	if metadataJSON != nil && *metadataJSON != "" {
		if !json.Valid([]byte(*metadataJSON)) {
			return nil, pkgerrors.BadRequest(I18nInvalidInput, "invalid metadata JSON")
		}
		metadata = json.RawMessage(*metadataJSON)
	} else {
		metadata = json.RawMessage("{}")
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, pkgerrors.Internal(I18nCreateFailed, "failed to generate ID")
	}

	novel := &domain.Novel{
		ID:               id,
		Title:            title,
		Slug:             novelSlug,
		Synopsis:         synopsis,
		CoverImageURL:    coverImageURL,
		ThumbnailURL:     thumbnailURL,
		Status:           domain.NovelStatus(*status),
		IsOneshot:        isOneshot,
		OriginalLanguage: originalLanguage,
		OriginalTitle:    originalTitle,
		Metadata:         metadata,
		OwnerID:          ownerID,
		OwnerType:        ownerType,
		CreatedBy:        createdBy,
	}

	if err := s.novelRepo.Create(ctx, novel); err != nil {
		return nil, err
	}

	if len(genreIDs) > 0 {
		if err := s.genreRepo.AddNovelGenres(ctx, id, genreIDs, ownerID); err != nil {
			fmt.Printf("Failed to add genres: %v\n", err)
		} else {
			increments := lo.SliceToMap(genreIDs, func(gid uuid.UUID) (uuid.UUID, int) {
				return gid, 1
			})
			if err := s.genreRepo.BatchIncrementNovelCount(ctx, increments); err != nil {
				fmt.Printf("Failed to increment genre novel count: %v\n", err)
			}
		}
	}

	for i, authorID := range authorIDs {
		if _, err := s.authorRepo.GetByID(ctx, authorID); err != nil {
			continue
		}
		if err := s.authorRepo.AddNovelAuthor(ctx, id, authorID, i); err != nil {
			fmt.Printf("Failed to add author %s: %v\n", authorID, err)
		}
	}

	for i, artistID := range artistIDs {
		if _, err := s.artistRepo.GetByID(ctx, artistID); err != nil {
			continue
		}
		if err := s.artistRepo.AddNovelArtist(ctx, id, artistID, i); err != nil {
			fmt.Printf("Failed to add artist %s: %v\n", artistID, err)
		}
	}

	if ownerType == "user" {
		if err := s.creatorRepo.IncrementNovelCount(ctx, ownerID); err != nil {
			fmt.Printf("Failed to increment novel count for user %s: %v\n", ownerID, err)
		}
	}

	return s.novelRepo.GetByID(ctx, id)
}

// CreateNovelEntity creates only the novel entity without relations
// This is used by CreateNovelUseCase for orchestrated creation
func (s *novelServiceImpl) CreateNovelEntity(ctx context.Context, novel *domain.Novel) error {
	return s.novelRepo.Create(ctx, novel)
}

// UpdateNovel cập nhật thông tin novel
func (s *novelServiceImpl) UpdateNovel(ctx context.Context, id uuid.UUID, title string, synopsis json.RawMessage, coverImageURL, thumbnailURL *string, status, originalLanguage, originalTitle *string, metadataJSON *string, isOneshot bool) (*domain.Novel, error) {
	if title == "" {
		return nil, pkgerrors.BadRequest(I18nInvalidInput, "title is required")
	}

	if status == nil || !domain.NovelStatus(*status).IsValid() {
		return nil, pkgerrors.BadRequest(I18nInvalidStatus, "invalid novel status")
	}

	novel, err := s.novelRepo.GetByID(ctx, id)
	if err != nil {
		return nil, pkgerrors.NotFound(I18nNotFound, "novel not found")
	}

	newSlug := slug.Make(title)
	if newSlug != novel.Slug {
		existing, err := s.novelRepo.GetBySlug(ctx, newSlug)
		if err == nil && existing != nil && existing.ID != id {
			return nil, pkgerrors.Conflict(I18nSlugAlreadyExists, "slug already exists")
		}
		novel.Slug = newSlug
	}

	novel.Title = title
	novel.CoverImageURL = coverImageURL
	novel.ThumbnailURL = thumbnailURL
	novel.Status = domain.NovelStatus(*status)
	novel.IsOneshot = isOneshot
	novel.OriginalLanguage = originalLanguage
	novel.OriginalTitle = originalTitle

	if len(synopsis) > 0 && string(synopsis) != "null" {
		if !json.Valid(synopsis) {
			return nil, pkgerrors.BadRequest(I18nInvalidInput, "invalid synopsis JSON")
		}
		novel.Synopsis = synopsis
	}

	if metadataJSON != nil && *metadataJSON != "" {
		if !json.Valid([]byte(*metadataJSON)) {
			return nil, pkgerrors.BadRequest(I18nInvalidInput, "invalid metadata JSON")
		}
		novel.Metadata = json.RawMessage(*metadataJSON)
	} else {
		novel.Metadata = json.RawMessage("{}")
	}

	if novel.Status == domain.NovelStatusCompleted && novel.CompletedAt == nil {
		now := novel.UpdatedAt
		novel.CompletedAt = &now
	}

	if err := s.novelRepo.Update(ctx, novel); err != nil {
		return nil, err
	}

	return s.novelRepo.GetByID(ctx, id)
}

// DeleteNovel xóa novel (soft delete)
func (s *novelServiceImpl) DeleteNovel(ctx context.Context, id uuid.UUID) error {
	novel, err := s.novelRepo.GetByID(ctx, id)
	if err != nil {
		return pkgerrors.NotFound(I18nNotFound, "novel not found")
	}

	genres, err := s.genreRepo.GetNovelGenres(ctx, id)
	if err != nil {
		fmt.Printf("Failed to get novel genres for deletion: %v\n", err)
	}

	if err := s.novelRepo.Delete(ctx, id); err != nil {
		return err
	}

	if len(genres) > 0 {
		increments := make(map[uuid.UUID]int)
		for _, genre := range genres {
			increments[genre.ID] = -1
		}
		if err := s.genreRepo.BatchIncrementNovelCount(ctx, increments); err != nil {
			fmt.Printf("Failed to decrement genre novel count: %v\n", err)
		}
	}

	if novel.OwnerType == "user" {
		if err := s.creatorRepo.DecrementNovelCount(ctx, novel.OwnerID); err != nil {
			fmt.Printf("Failed to decrement novel count for user %s: %v\n", novel.OwnerID, err)
		}
	}

	return nil
}

// DeleteNovelEntity deletes only the novel entity without relations logic
// This is used by DeleteNovelUseCase for orchestrated deletion
func (s *novelServiceImpl) DeleteNovelEntity(ctx context.Context, id uuid.UUID) error {
	return s.novelRepo.Delete(ctx, id)
}

// GetNovelByID lấy thông tin novel theo ID
func (s *novelServiceImpl) GetNovelByID(ctx context.Context, id uuid.UUID) (*domain.Novel, error) {
	return s.novelRepo.GetByID(ctx, id)
}

// GetNovelBySlug lấy thông tin novel theo slug
func (s *novelServiceImpl) GetNovelBySlug(ctx context.Context, slug string) (*domain.Novel, error) {
	return s.novelRepo.GetBySlug(ctx, slug)
}

// ListNovels lấy danh sách novels với pagination, search và sort
func (s *novelServiceImpl) ListNovels(ctx context.Context, page, limit int, ownerID *uuid.UUID, keySearch string, genreIDs []uuid.UUID, statusStrs []string, originalLanguage, sortBy, sortOrder string) ([]*domain.Novel, int, error) {
	// Validate and set defaults
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Use NovelSortField enum for type-safe sorting
	sortField := domain.NovelSortField(sortBy)
	if !sortField.IsValid() {
		sortField = domain.NovelSortCreatedAt
	}

	// Validate sortOrder
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// Build filter
	offset := (page - 1) * limit
	var searchQuery *string
	if keySearch != "" {
		searchQuery = &keySearch
	}

	// Convert string statuses to domain.NovelStatus slice
	statuses := lo.FilterMap(statusStrs, func(s string, _ int) (domain.NovelStatus, bool) {
		status := domain.NovelStatus(s)
		return status, status.IsValid()
	})

	var origLang *string
	if originalLanguage != "" {
		origLang = &originalLanguage
	}

	filter := domain.NovelFilter{
		OwnerID:          ownerID,
		SearchQuery:      searchQuery,
		GenreIDs:         genreIDs,
		Statuses:         statuses,
		OriginalLanguage: origLang,
		SortBy:           sortField.ToDBColumn(),
		SortOrder:        sortOrder,
		Limit:            limit,
		Offset:           offset,
	}

	novels, total, err := s.novelRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return novels, int(total), nil
}

// GetNovelsByIDs lấy danh sách novels theo list IDs
func (s *novelServiceImpl) GetNovelsByIDs(ctx context.Context, ids []uuid.UUID) ([]*domain.Novel, error) {
	if len(ids) == 0 {
		return []*domain.Novel{}, nil
	}
	// Use repo List with IDs filter
	novels, _, err := s.novelRepo.List(ctx, domain.NovelFilter{
		IDs:   ids,
		Limit: len(ids),
	})
	return novels, err
}

// IncrementViewCount tăng view count của novel
func (s *novelServiceImpl) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	return s.novelRepo.IncrementViewCount(ctx, id)
}

// GetNovelGenres lấy danh sách genre IDs của novel
func (s *novelServiceImpl) GetNovelGenres(ctx context.Context, novelID uuid.UUID) ([]uuid.UUID, error) {
	return s.novelRepo.GetGenres(ctx, novelID)
}

// GetNovelAuthors lấy danh sách author IDs của novel
func (s *novelServiceImpl) GetNovelAuthors(ctx context.Context, novelID uuid.UUID) ([]uuid.UUID, error) {
	return s.novelRepo.GetAuthors(ctx, novelID)
}

// GetNovelArtists lấy danh sách artist IDs của novel
func (s *novelServiceImpl) GetNovelArtists(ctx context.Context, novelID uuid.UUID) ([]uuid.UUID, error) {
	return s.novelRepo.GetArtists(ctx, novelID)
}

// GetNovelGenresDetails lấy danh sách genre (chi tiết) của novel
func (s *novelServiceImpl) GetNovelGenresDetails(ctx context.Context, novelID uuid.UUID) ([]*domain.Genre, error) {
	return s.genreRepo.GetNovelGenres(ctx, novelID)
}

// GetNovelAuthorsDetails lấy danh sách author (chi tiết) của novel
func (s *novelServiceImpl) GetNovelAuthorsDetails(ctx context.Context, novelID uuid.UUID) ([]*domain.NovelAuthor, error) {
	return s.authorRepo.GetNovelAuthors(ctx, novelID)
}

// GetNovelArtistsDetails lấy danh sách artist (chi tiết) của novel
func (s *novelServiceImpl) GetNovelArtistsDetails(ctx context.Context, novelID uuid.UUID) ([]*domain.NovelArtist, error) {
	return s.artistRepo.GetNovelArtists(ctx, novelID)
}

// NovelFullRepository là interface mở rộng cho hàm GetNovelFullBySlug
type NovelFullRepository interface {
	GetNovelFullBySlug(ctx context.Context, slug string) (*NovelFullData, error)
}

// GetNovelFull lấy toàn bộ dữ liệu novel cho trang chi tiết (public API)
func (s *novelServiceImpl) GetNovelFull(ctx context.Context, slug string) (*NovelFullData, error) {
	// Type assertion để truy cập method GetNovelFullBySlug
	if fullRepo, ok := s.novelRepo.(NovelFullRepository); ok {
		return fullRepo.GetNovelFullBySlug(ctx, slug)
	}
	// Fallback: không hỗ trợ, trả về error
	return nil, pkgerrors.Internal("novel.get_full_not_supported", "GetNovelFull not supported by repository")
}
