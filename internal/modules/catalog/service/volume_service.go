package service

import (
	"context"
	"wibusystem/internal/modules/catalog/domain"

	"github.com/google/uuid"
)

// VolumeService defines the interface for volume-related business logic.
type VolumeService interface {
	CreateVolume(ctx context.Context, novelID uuid.UUID, title string, volumeNumber int) (*domain.Volume, error)
	GetVolume(ctx context.Context, id uuid.UUID) (*domain.Volume, error)
	ListVolumesByNovel(ctx context.Context, novelID uuid.UUID, page, pageSize int) ([]*domain.Volume, int64, error)
}
