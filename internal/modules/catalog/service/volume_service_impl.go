package service

import (
	"context"
	"fmt"
	"wibusystem/internal/modules/catalog/domain"
	"wibusystem/internal/modules/catalog/repository"

	"github.com/google/uuid"
)

// volumeService implements the VolumeService interface.
type volumeService struct {
	volumeRepo repository.VolumeRepository
	novelRepo  repository.NovelRepository // To check if novel exists
}

// NewVolumeService creates a new instance of VolumeService.
func NewVolumeService(volumeRepo repository.VolumeRepository, novelRepo repository.NovelRepository) VolumeService {
	return &volumeService{volumeRepo: volumeRepo, novelRepo: novelRepo}
}

// CreateVolume handles the business logic for creating a new volume.
func (s *volumeService) CreateVolume(ctx context.Context, novelID uuid.UUID, title string, volumeNumber int) (*domain.Volume, error) {
	// Check if the novel exists
	_, err := s.novelRepo.GetByID(ctx, novelID)
	if err != nil {
		return nil, fmt.Errorf("cannot create volume for a non-existent novel: %w", err)
	}

	volume, err := domain.NewVolume(novelID, volumeNumber)
	if err != nil {
		return nil, err
	}
	volume.VolumeTitle = title

	if err := s.volumeRepo.Create(ctx, volume); err != nil {
		return nil, err
	}

	return volume, nil
}

// GetVolume retrieves a single volume.
func (s *volumeService) GetVolume(ctx context.Context, id uuid.UUID) (*domain.Volume, error) {
	return s.volumeRepo.GetByID(ctx, id)
}

// ListVolumesByNovel retrieves a list of volumes for a specific novel.
func (s *volumeService) ListVolumesByNovel(ctx context.Context, novelID uuid.UUID, page, pageSize int) ([]*domain.Volume, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	return s.volumeRepo.ListByNovelID(ctx, novelID, pageSize, offset)
}
