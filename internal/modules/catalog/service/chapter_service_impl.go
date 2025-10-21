package service

import (
	"context"
	"fmt"
	"wibusystem/internal/modules/catalog/domain"
	"wibusystem/internal/modules/catalog/repository"

	"github.com/google/uuid"
)

// chapterService implements the ChapterService interface.
type chapterService struct {
	chapterRepo repository.ChapterRepository
	volumeRepo  repository.VolumeRepository // To check if volume exists
}

// NewChapterService creates a new instance of ChapterService.
func NewChapterService(chapterRepo repository.ChapterRepository, volumeRepo repository.VolumeRepository) ChapterService {
	return &chapterService{chapterRepo: chapterRepo, volumeRepo: volumeRepo}
}

// CreateChapter handles the business logic for creating a new chapter.
func (s *chapterService) CreateChapter(ctx context.Context, volumeID uuid.UUID, title, content string, chapterNumber int) (*domain.Chapter, error) {
	// Check if the volume exists
	_, err := s.volumeRepo.GetByID(ctx, volumeID)
	if err != nil {
		return nil, fmt.Errorf("cannot create chapter for a non-existent volume: %w", err)
	}

	chapter, err := domain.NewChapter(volumeID, chapterNumber)
	if err != nil {
		return nil, err
	}
	chapter.Title = title
	chapter.Content = []byte(content)

	if err := s.chapterRepo.Create(ctx, chapter); err != nil {
		return nil, err
	}

	// In a real implementation, we might update the chapter count in the volume here.

	return chapter, nil
}

// GetChapter retrieves a single chapter.
func (s *chapterService) GetChapter(ctx context.Context, id uuid.UUID) (*domain.Chapter, error) {
	return s.chapterRepo.GetByID(ctx, id)
}

// ListChaptersByVolume retrieves a list of chapters for a specific volume.
func (s *chapterService) ListChaptersByVolume(ctx context.Context, volumeID uuid.UUID, page, pageSize int) ([]*domain.Chapter, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50 // Chapters might be more numerous
	}
	offset := (page - 1) * pageSize

	return s.chapterRepo.ListByVolumeID(ctx, volumeID, pageSize, offset)
}
