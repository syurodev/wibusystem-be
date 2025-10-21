package repository

import (
	"context"
	"wibusystem/internal/modules/catalog/domain"

	"github.com/google/uuid"
)

// ChapterRepository defines the interface for interacting with chapter data.
type ChapterRepository interface {
	Create(ctx context.Context, chapter *domain.Chapter) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Chapter, error)
	Update(ctx context.Context, chapter *domain.Chapter) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByVolumeID(ctx context.Context, volumeID uuid.UUID, limit, offset int) ([]*domain.Chapter, int64, error)
}
