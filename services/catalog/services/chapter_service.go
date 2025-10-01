package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	d "wibusystem/pkg/common/dto"
	"wibusystem/services/catalog/repositories"
)

// ChapterService implements business logic for chapter management operations.
// This service handles UUID validation, delegates to the repository layer,
// and maps domain models to response DTOs.
type ChapterService struct {
	repos *repositories.Repositories
}

// NewChapterService creates a new ChapterService instance with the given repository dependencies.
func NewChapterService(repos *repositories.Repositories) *ChapterService {
	return &ChapterService{
		repos: repos,
	}
}

// CreateChapter creates a new chapter within a volume.
// Validates volume UUID format before delegating to the repository.
func (s *ChapterService) CreateChapter(ctx context.Context, volumeID string, req d.CreateChapterRequest) (*d.CreateChapterResponse, error) {
	// Parse and validate volume UUID
	volumeUUID, err := uuid.Parse(volumeID)
	if err != nil {
		return nil, fmt.Errorf("invalid volume ID format: %w", err)
	}

	// Delegate to repository
	chapter, err := s.repos.Chapter.CreateChapter(ctx, volumeUUID, req)
	if err != nil {
		return nil, err
	}

	// Map to response DTO
	return &d.CreateChapterResponse{
		ID:                 chapter.ID.String(),
		VolumeID:           chapter.VolumeID.String(),
		ChapterNumber:      chapter.ChapterNumber,
		Title:              chapter.Title,
		IsDraft:            chapter.IsDraft,
		WordCount:          chapter.WordCount,
		CharacterCount:     chapter.CharacterCount,
		ReadingTimeMinutes: chapter.ReadingTimeMinutes,
		CreatedAt:          chapter.CreatedAt,
	}, nil
}

// GetChapterByID retrieves a specific chapter by its ID.
// Validates chapter UUID format and controls content inclusion based on the flag.
func (s *ChapterService) GetChapterByID(ctx context.Context, id string, includeContent bool) (*d.ChapterResponse, error) {
	// Parse and validate chapter UUID
	chapterUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid chapter ID format: %w", err)
	}

	// Delegate to repository
	chapter, err := s.repos.Chapter.GetChapterByID(ctx, chapterUUID, includeContent)
	if err != nil {
		return nil, err
	}

	// Map to response DTO
	response := &d.ChapterResponse{
		ID:                 chapter.ID.String(),
		VolumeID:           chapter.VolumeID.String(),
		ChapterNumber:      chapter.ChapterNumber,
		Title:              chapter.Title,
		PublishedAt:        chapter.PublishedAt,
		IsPublic:           chapter.IsPublic,
		IsDraft:            chapter.IsDraft,
		PriceCoins:         chapter.PriceCoins,
		WordCount:          chapter.WordCount,
		CharacterCount:     chapter.CharacterCount,
		ReadingTimeMinutes: chapter.ReadingTimeMinutes,
		ViewCount:          chapter.ViewCount,
		LikeCount:          chapter.LikeCount,
		CommentCount:       chapter.CommentCount,
		HasMatureContent:   chapter.HasMatureContent,
		Version:            chapter.Version,
		CreatedAt:          chapter.CreatedAt,
		UpdatedAt:          chapter.UpdatedAt,
	}

	// Include content only if requested
	if includeContent && chapter.Content != nil {
		response.Content = chapter.Content
	}

	// Parse content warnings JSONB array
	if chapter.ContentWarnings != nil {
		var warnings []string
		if err := json.Unmarshal(*chapter.ContentWarnings, &warnings); err == nil {
			response.ContentWarnings = warnings
		}
	}

	return response, nil
}

// ListChaptersByVolumeID retrieves a paginated list of chapters in a volume.
// Validates volume UUID format and applies pagination settings.
func (s *ChapterService) ListChaptersByVolumeID(ctx context.Context, volumeID string, req d.ListChaptersRequest) (*d.PaginatedChaptersResponse, error) {
	// Parse and validate volume UUID
	volumeUUID, err := uuid.Parse(volumeID)
	if err != nil {
		return nil, fmt.Errorf("invalid volume ID format: %w", err)
	}

	// Set default pagination values
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 50
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	// Delegate to repository
	response, err := s.repos.Chapter.ListChaptersByVolumeID(ctx, volumeUUID, req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// UpdateChapter updates an existing chapter's information.
// Validates chapter UUID format before delegating to the repository.
func (s *ChapterService) UpdateChapter(ctx context.Context, id string, req d.UpdateChapterRequest) (*d.UpdateChapterResponse, error) {
	// Parse and validate chapter UUID
	chapterUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid chapter ID format: %w", err)
	}

	// Delegate to repository
	chapter, err := s.repos.Chapter.UpdateChapter(ctx, chapterUUID, req)
	if err != nil {
		return nil, err
	}

	// Map to response DTO
	return &d.UpdateChapterResponse{
		ID:                 chapter.ID.String(),
		Title:              chapter.Title,
		WordCount:          chapter.WordCount,
		CharacterCount:     chapter.CharacterCount,
		ReadingTimeMinutes: chapter.ReadingTimeMinutes,
		Version:            chapter.Version,
		UpdatedAt:          chapter.UpdatedAt,
	}, nil
}

// DeleteChapter soft-deletes a chapter.
// Validates chapter UUID format and checks for existing purchases before deletion.
func (s *ChapterService) DeleteChapter(ctx context.Context, id string) error {
	// Parse and validate chapter UUID
	chapterUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid chapter ID format: %w", err)
	}

	// Delegate to repository
	return s.repos.Chapter.DeleteChapter(ctx, chapterUUID)
}

// PublishChapter publishes a chapter, making it publicly accessible.
// Validates chapter UUID format and sets the publication time.
func (s *ChapterService) PublishChapter(ctx context.Context, id string, req d.PublishChapterRequest) (*d.PublishChapterResponse, error) {
	// Parse and validate chapter UUID
	chapterUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid chapter ID format: %w", err)
	}

	// Delegate to repository
	chapter, err := s.repos.Chapter.PublishChapter(ctx, chapterUUID, req.PublishAt)
	if err != nil {
		return nil, err
	}

	// Map to response DTO
	return &d.PublishChapterResponse{
		ID:          chapter.ID.String(),
		IsPublic:    chapter.IsPublic,
		IsDraft:     chapter.IsDraft,
		PublishedAt: chapter.PublishedAt,
	}, nil
}

// UnpublishChapter unpublishes a chapter, removing it from public access.
// Validates chapter UUID format before delegating to the repository.
func (s *ChapterService) UnpublishChapter(ctx context.Context, id string) (*d.PublishChapterResponse, error) {
	// Parse and validate chapter UUID
	chapterUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid chapter ID format: %w", err)
	}

	// Delegate to repository
	chapter, err := s.repos.Chapter.UnpublishChapter(ctx, chapterUUID)
	if err != nil {
		return nil, err
	}

	// Map to response DTO
	return &d.PublishChapterResponse{
		ID:          chapter.ID.String(),
		IsPublic:    chapter.IsPublic,
		IsDraft:     chapter.IsDraft,
		PublishedAt: chapter.PublishedAt,
	}, nil
}