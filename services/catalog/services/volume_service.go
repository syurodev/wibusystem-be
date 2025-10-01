package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	d "wibusystem/pkg/common/dto"
	"wibusystem/services/catalog/repositories"
	"wibusystem/services/catalog/services/interfaces"
)

// VolumeService implements volume-related business logic
// This service handles all business operations for novel volumes,
// coordinating between HTTP handlers and repository layer.
type VolumeService struct {
	repos *repositories.Repositories
}

// NewVolumeService creates a new volume service instance
// Takes repositories as dependency for data access operations
func NewVolumeService(repos *repositories.Repositories) interfaces.VolumeServiceInterface {
	return &VolumeService{
		repos: repos,
	}
}

// CreateVolume creates a new volume for a specific novel
// Validates novel ID format and delegates to repository for creation
func (s *VolumeService) CreateVolume(ctx context.Context, novelID string, req d.CreateVolumeRequest) (*d.CreateVolumeResponse, error) {
	// Parse and validate novel UUID
	novelUUID, err := uuid.Parse(novelID)
	if err != nil {
		return nil, fmt.Errorf("invalid novel ID format: %w", err)
	}

	// Create volume through repository
	volume, err := s.repos.Volume.CreateVolume(ctx, novelUUID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create volume: %w", err)
	}

	// Map to response DTO
	response := &d.CreateVolumeResponse{
		ID:           volume.ID.String(),
		NovelID:      volume.NovelID.String(),
		VolumeNumber: volume.VolumeNumber,
		Title:        volume.VolumeTitle,
		CreatedAt:    volume.CreatedAt,
	}

	return response, nil
}

// GetVolumeByID retrieves detailed information about a specific volume
// Converts volume model to response DTO with all required fields
func (s *VolumeService) GetVolumeByID(ctx context.Context, id string) (*d.VolumeResponse, error) {
	// Parse and validate volume UUID
	volumeUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid volume ID format: %w", err)
	}

	// Get volume from repository
	volume, err := s.repos.Volume.GetVolumeByID(ctx, volumeUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get volume: %w", err)
	}

	// Map to response DTO
	response := &d.VolumeResponse{
		ID:           volume.ID.String(),
		NovelID:      volume.NovelID.String(),
		VolumeNumber: volume.VolumeNumber,
		Title:        volume.VolumeTitle,
		Description:  volume.Description,
		CoverImage:   volume.CoverImage,
		PublishedAt:  volume.PublishedAt,
		IsPublic:     volume.IsAvailable,
		PriceCoins:   volume.PriceCoins,
		ChapterCount: volume.ChapterCount,
		CreatedAt:    volume.CreatedAt,
		UpdatedAt:    volume.UpdatedAt,
		Chapters:     []interface{}{}, // Empty array, will be populated when chapter service is ready
	}

	return response, nil
}

// ListVolumesByNovelID retrieves all volumes for a specific novel with pagination
// Validates novel ID and returns paginated volume list
func (s *VolumeService) ListVolumesByNovelID(ctx context.Context, novelID string, req d.ListVolumesRequest) (*d.PaginatedVolumesResponse, error) {
	// Parse and validate novel UUID
	novelUUID, err := uuid.Parse(novelID)
	if err != nil {
		return nil, fmt.Errorf("invalid novel ID format: %w", err)
	}

	// Get volumes from repository
	response, err := s.repos.Volume.ListVolumesByNovelID(ctx, novelUUID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %w", err)
	}

	return response, nil
}

// UpdateVolume updates an existing volume
// Validates volume ID and request, then updates through repository
func (s *VolumeService) UpdateVolume(ctx context.Context, id string, req d.UpdateVolumeRequest) (*d.UpdateVolumeResponse, error) {
	// Parse and validate volume UUID
	volumeUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid volume ID format: %w", err)
	}

	// Update volume through repository
	volume, err := s.repos.Volume.UpdateVolume(ctx, volumeUUID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update volume: %w", err)
	}

	// Map to response DTO
	response := &d.UpdateVolumeResponse{
		ID:        volume.ID.String(),
		Title:     volume.VolumeTitle,
		UpdatedAt: volume.UpdatedAt,
	}

	return response, nil
}

// DeleteVolume soft-deletes a volume
// Validates that no users have purchased content before deletion
func (s *VolumeService) DeleteVolume(ctx context.Context, id string) error {
	// Parse and validate volume UUID
	volumeUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid volume ID format: %w", err)
	}

	// Delete volume through repository (includes purchase checks)
	err = s.repos.Volume.DeleteVolume(ctx, volumeUUID)
	if err != nil {
		return fmt.Errorf("failed to delete volume: %w", err)
	}

	return nil
}