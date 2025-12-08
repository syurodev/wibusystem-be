package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/gosimple/slug"

	"system/internal/domain"
	"system/internal/pkg/db"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/slugutil"
)

// NovelService cung cấp business logic cho novels
type NovelService struct {
	novelRepo  domain.NovelRepository
	volumeRepo domain.VolumeRepository
	genreRepo  domain.GenreRepository
	authorRepo domain.AuthorRepository
	artistRepo domain.ArtistRepository
	creatorRepo domain.CreatorRepository
	txManager  db.TransactionManager
}

// NewNovelService tạo một instance mới của NovelService
func NewNovelService(
	novelRepo domain.NovelRepository,
	volumeRepo domain.VolumeRepository,
	genreRepo domain.GenreRepository,
	authorRepo domain.AuthorRepository,
	artistRepo domain.ArtistRepository,
	creatorRepo domain.CreatorRepository,
	txManager db.TransactionManager,
) *NovelService {
	return &NovelService{
		novelRepo:   novelRepo,
		volumeRepo:  volumeRepo,
		genreRepo:   genreRepo,
		authorRepo:  authorRepo,
		artistRepo:  artistRepo,
		creatorRepo: creatorRepo,
		txManager:   txManager,
	}
}

// CreateNovel tạo novel mới
func (s *NovelService) CreateNovel(
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
	// Validate input
	if title == "" {
		return nil, pkgerrors.ErrInvalidInput
	}

	// Validate status
	if status == nil || !isValidNovelStatus(*status) {
		return nil, pkgerrors.ErrInvalidInput
	}

	// Generate unique slug from title with random suffix
	novelSlug, err := slugutil.GenerateUniqueSlug(title)
	if err != nil {
		return nil, err
	}

	// Prepare synopsis JSON
	// If synopsis is empty or null, default to empty object
	if len(synopsis) == 0 || string(synopsis) == "null" {
		synopsis = json.RawMessage("{}")
	} else {
		// Validate JSON
		if !json.Valid(synopsis) {
			return nil, pkgerrors.ErrInvalidInput
		}
	}

	// Prepare metadata JSON
	var metadata json.RawMessage
	if metadataJSON != nil && *metadataJSON != "" {
		// Validate JSON
		if !json.Valid([]byte(*metadataJSON)) {
			return nil, pkgerrors.ErrInvalidInput
		}
		metadata = json.RawMessage(*metadataJSON)
	} else {
		metadata = json.RawMessage("{}")
	}

	// Generate ID (UUID v7)
	id, err := uuid.NewV7()
	if err != nil {
		return nil, pkgerrors.ErrInternalServer
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

	// Link Genres
	// Validate genres presence first if needed, but for performance we might skip or do a bulk check.
	// Here we just proceed with batch insert. Valid foreign keys will enforce correctness if DB constraints exist,
	// but usually better to let repo handle valid constraint errors or check beforehand.
	// For now, replacing the loop as requested.
	
	if len(genreIDs) > 0 {
		if err := s.genreRepo.AddNovelGenres(ctx, id, genreIDs, ownerID); err != nil {
			fmt.Printf("Failed to add genres: %v\n", err)
		} else {
             // For increments, we assume all succeeded.
             // We need to build the map for BatchIncrementNovelCount
             increments := make(map[uuid.UUID]int)
             for _, gid := range genreIDs {
                 increments[gid] = 1
             }
             if err := s.genreRepo.BatchIncrementNovelCount(ctx, increments); err != nil {
                 fmt.Printf("Failed to increment genre novel count: %v\n", err)
             }
        }
	}


	// Link Authors
	for i, authorID := range authorIDs {
		// Validate author existence
		if _, err := s.authorRepo.GetByID(ctx, authorID); err != nil {
			continue
		}
		if err := s.authorRepo.AddNovelAuthor(ctx, id, authorID, i); err != nil {
			fmt.Printf("Failed to add author %s: %v\n", authorID, err)
		}
	}

	// Link Artists
	for i, artistID := range artistIDs {
		// Validate artist existence
		if _, err := s.artistRepo.GetByID(ctx, artistID); err != nil {
			continue
		}
		if err := s.artistRepo.AddNovelArtist(ctx, id, artistID, i); err != nil {
			fmt.Printf("Failed to add artist %s: %v\n", artistID, err)
		}
	}

	// Increment works count for owner if it's a user
	if ownerType == "user" {
		// We assume ownerID is the user ID
		if err := s.creatorRepo.IncrementWorksCount(ctx, ownerID); err != nil {
			fmt.Printf("Failed to increment works count for user %s: %v\n", ownerID, err)
			// Non-blocking error, log and continue
		}
	}

	return s.novelRepo.GetByID(ctx, id)
}

// UpdateNovel cập nhật thông tin novel
func (s *NovelService) UpdateNovel(ctx context.Context, id uuid.UUID, title string, synopsis json.RawMessage, coverImageURL, thumbnailURL *string, status, originalLanguage, originalTitle *string, metadataJSON *string, isOneshot bool) (*domain.Novel, error) {
	// Validate input
	if title == "" {
		return nil, pkgerrors.ErrInvalidInput
	}

	// Validate status
	if status == nil || !isValidNovelStatus(*status) {
		return nil, pkgerrors.ErrInvalidInput
	}

	// Get existing novel
	novel, err := s.novelRepo.GetByID(ctx, id)
	if err != nil {
		return nil, pkgerrors.ErrNovelNotFound
	}

	// Generate new slug if title changed
	newSlug := slug.Make(title)
	if newSlug != novel.Slug {
		// Check if new slug already exists
		existing, err := s.novelRepo.GetBySlug(ctx, newSlug)
		if err == nil && existing != nil && existing.ID != id {
			return nil, pkgerrors.ErrSlugAlreadyExists
		}
		novel.Slug = newSlug
	}

	// Update fields
	novel.Title = title
	novel.CoverImageURL = coverImageURL
	novel.ThumbnailURL = thumbnailURL
	novel.Status = domain.NovelStatus(*status)
	novel.IsOneshot = isOneshot
	novel.OriginalLanguage = originalLanguage
	novel.OriginalTitle = originalTitle

	// Update synopsis JSON
	if len(synopsis) > 0 && string(synopsis) != "null" {
		if !json.Valid(synopsis) {
			return nil, pkgerrors.ErrInvalidInput
		}
		novel.Synopsis = synopsis
	}

	// Update metadata JSON
	if metadataJSON != nil && *metadataJSON != "" {
		// Validate JSON
		if !json.Valid([]byte(*metadataJSON)) {
			return nil, pkgerrors.ErrInvalidInput
		}
		novel.Metadata = json.RawMessage(*metadataJSON)
	} else {
		novel.Metadata = json.RawMessage("{}")
	}

	// Handle status changes
	if novel.Status == domain.NovelStatusCompleted && novel.CompletedAt == nil {
		now := novel.UpdatedAt // Will be updated by repository
		novel.CompletedAt = &now
	}

	if err := s.novelRepo.Update(ctx, novel); err != nil {
		return nil, err
	}

	// Retrieve updated novel
	return s.novelRepo.GetByID(ctx, id)
}

// DeleteNovel xóa novel (soft delete)
func (s *NovelService) DeleteNovel(ctx context.Context, id uuid.UUID) error {
	// Check if novel exists
	novel, err := s.novelRepo.GetByID(ctx, id)
	if err != nil {
		return pkgerrors.ErrNovelNotFound
	}

	// TODO: Check if novel has volumes/chapters and prevent deletion if needed
	// For now, we allow deletion regardless

	// Get novel genres to decrement count
	genres, err := s.genreRepo.GetNovelGenres(ctx, id)
	if err != nil {
		// Log error but proceed with deletion
		fmt.Printf("Failed to get novel genres for deletion: %v\n", err)
	}

	if err := s.novelRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Decrement novel count for genres
	if len(genres) > 0 {
		increments := make(map[uuid.UUID]int)
		for _, genre := range genres {
			increments[genre.ID] = -1
		}
		if err := s.genreRepo.BatchIncrementNovelCount(ctx, increments); err != nil {
			fmt.Printf("Failed to decrement genre novel count: %v\n", err)
		}
	}

	// Decrement works count for owner if it's a user
	if novel.OwnerType == "user" {
		if err := s.creatorRepo.DecrementWorksCount(ctx, novel.OwnerID); err != nil {
			fmt.Printf("Failed to decrement works count for user %s: %v\n", novel.OwnerID, err)
		}
	}

	return nil
}

// GetNovelByID lấy thông tin novel theo ID
func (s *NovelService) GetNovelByID(ctx context.Context, id uuid.UUID) (*domain.Novel, error) {
	return s.novelRepo.GetByID(ctx, id)
}

// GetNovelBySlug lấy thông tin novel theo slug
func (s *NovelService) GetNovelBySlug(ctx context.Context, slug string) (*domain.Novel, error) {
	return s.novelRepo.GetBySlug(ctx, slug)
}

// ListNovels lấy danh sách novels với pagination, search và sort
func (s *NovelService) ListNovels(ctx context.Context, page, limit int, ownerID *uuid.UUID, keySearch string, genreIDs []uuid.UUID, statusStrs []string, originalLanguage, sortBy, sortOrder string) ([]*domain.Novel, int, error) {
	// Validate and set defaults
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Validate sortBy
	validSortFields := map[string]string{
		"created_at":   "created_at",
		"rating":       "rating_average",
		"views":        "view_count",
		"last_chapter": "last_chapter_at",
	}

	sortField := "created_at" // default
	if sortBy != "" {
		if field, ok := validSortFields[sortBy]; ok {
			sortField = field
		}
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
	var statuses []domain.NovelStatus
	for _, statusStr := range statusStrs {
		if isValidNovelStatus(statusStr) {
			statuses = append(statuses, domain.NovelStatus(statusStr))
		}
	}

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
		SortBy:           sortField,
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

// IncrementViewCount tăng view count của novel
func (s *NovelService) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	return s.novelRepo.IncrementViewCount(ctx, id)
}

// Helper function to validate novel status
func isValidNovelStatus(status string) bool {
	validStatuses := map[string]bool{
		"draft":     true,
		"ongoing":   true,
		"completed": true,
		"hiatus":    true,
		"dropped":   true,
	}
	return validStatuses[status]
}

// GetNovelGenres lấy danh sách genre IDs của novel
func (s *NovelService) GetNovelGenres(ctx context.Context, novelID uuid.UUID) ([]uuid.UUID, error) {
	return s.novelRepo.GetGenres(ctx, novelID)
}

// GetNovelAuthors lấy danh sách author IDs của novel
func (s *NovelService) GetNovelAuthors(ctx context.Context, novelID uuid.UUID) ([]uuid.UUID, error) {
	return s.novelRepo.GetAuthors(ctx, novelID)
}

// GetNovelArtists lấy danh sách artist IDs của novel
func (s *NovelService) GetNovelArtists(ctx context.Context, novelID uuid.UUID) ([]uuid.UUID, error) {
	return s.novelRepo.GetArtists(ctx, novelID)
}

// GetNovelGenresDetails lấy danh sách genre (chi tiết) của novel
func (s *NovelService) GetNovelGenresDetails(ctx context.Context, novelID uuid.UUID) ([]*domain.Genre, error) {
	return s.genreRepo.GetNovelGenres(ctx, novelID)
}

// GetNovelAuthorsDetails lấy danh sách author (chi tiết) của novel
func (s *NovelService) GetNovelAuthorsDetails(ctx context.Context, novelID uuid.UUID) ([]*domain.NovelAuthor, error) {
	return s.authorRepo.GetNovelAuthors(ctx, novelID)
}

// GetNovelArtistsDetails lấy danh sách artist (chi tiết) của novel
func (s *NovelService) GetNovelArtistsDetails(ctx context.Context, novelID uuid.UUID) ([]*domain.NovelArtist, error) {
	return s.artistRepo.GetNovelArtists(ctx, novelID)
}
