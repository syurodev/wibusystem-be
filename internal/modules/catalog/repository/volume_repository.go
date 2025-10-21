package repository

import (
	"context"
	"wibusystem/internal/modules/catalog/domain"

	"github.com/google/uuid"
)

// VolumeRepository defines the interface for interacting with volume data.
type VolumeRepository interface {
	Create(ctx context.Context, volume *domain.Volume) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Volume, error)
	Update(ctx context.Context, volume *domain.Volume) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByNovelID(ctx context.Context, novelID uuid.UUID, limit, offset int) ([]*domain.Volume, int64, error)
}
