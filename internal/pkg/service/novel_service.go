package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/gosimple/slug"

	"system/internal/domain"
	pkgerrors "system/pkg/errors"
)

// NovelService cung cấp business logic cho novels
type NovelService struct {
	novelRepo domain.NovelRepository
}

// NewNovelService tạo một instance mới của NovelService
func NewNovelService(novelRepo domain.NovelRepository) *NovelService {
	return &NovelService{
		novelRepo: novelRepo,
	}
}

// CreateNovel tạo novel mới
func (s *NovelService) CreateNovel(ctx context.Context, title, synopsis string, coverImageURL, thumbnailURL *string, status, originalLanguage, originalTitle *string, metadataJSON *string) (*domain.Novel, error) {
	// Validate input
	if title == "" {
		return nil, pkgerrors.ErrInvalidInput
	}

	// Validate status
	if status == nil || !isValidNovelStatus(*status) {
		return nil, pkgerrors.ErrInvalidInput
	}

	// Generate slug from title
	novelSlug := slug.Make(title)

	// Check if slug already exists
	existing, err := s.novelRepo.GetBySlug(ctx, novelSlug)
	if err == nil && existing != nil {
		return nil, pkgerrors.ErrSlugAlreadyExists
	}

	// Prepare synopsis JSON
	var synopsisJSON json.RawMessage
	if synopsis != "" {
		// Store as structured JSON
		synopsisJSON = json.RawMessage(fmt.Sprintf(`{"content": "%s"}`, synopsis))
	} else {
		synopsisJSON = json.RawMessage("{}")
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

	// Create novel
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	novel := &domain.Novel{
		ID:               id,
		Title:            title,
		Slug:             novelSlug,
		Synopsis:         synopsisJSON,
		CoverImageURL:    coverImageURL,
		ThumbnailURL:     thumbnailURL,
		Status:           domain.NovelStatus(*status),
		OriginalLanguage: originalLanguage,
		OriginalTitle:    originalTitle,
		Metadata:         metadata,
	}

	if err := s.novelRepo.Create(ctx, novel); err != nil {
		return nil, err
	}

	// Retrieve the created novel to get timestamps
	return s.novelRepo.GetByID(ctx, id)
}

// UpdateNovel cập nhật thông tin novel
func (s *NovelService) UpdateNovel(ctx context.Context, id uuid.UUID, title, synopsis string, coverImageURL, thumbnailURL *string, status, originalLanguage, originalTitle *string, metadataJSON *string) (*domain.Novel, error) {
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
	novel.OriginalLanguage = originalLanguage
	novel.OriginalTitle = originalTitle

	// Update synopsis JSON
	if synopsis != "" {
		novel.Synopsis = json.RawMessage(fmt.Sprintf(`{"content": "%s"}`, synopsis))
	} else {
		novel.Synopsis = json.RawMessage("{}")
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
	_, err := s.novelRepo.GetByID(ctx, id)
	if err != nil {
		return pkgerrors.ErrNovelNotFound
	}

	// TODO: Check if novel has volumes/chapters and prevent deletion if needed
	// For now, we allow deletion regardless

	return s.novelRepo.Delete(ctx, id)
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
func (s *NovelService) ListNovels(ctx context.Context, page, limit int, search, statusStr, originalLanguage, sortBy, sortOrder string) ([]*domain.Novel, int, error) {
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
	if search != "" {
		searchQuery = &search
	}

	var status *domain.NovelStatus
	if statusStr != "" && isValidNovelStatus(statusStr) {
		s := domain.NovelStatus(statusStr)
		status = &s
	}

	var origLang *string
	if originalLanguage != "" {
		origLang = &originalLanguage
	}

	filter := domain.NovelFilter{
		SearchQuery:      searchQuery,
		Status:           status,
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
