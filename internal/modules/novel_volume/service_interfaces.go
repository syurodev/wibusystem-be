package novel_volume

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
)

// VolumeService interface định nghĩa business logic cho volumes
type VolumeService interface {
	// CreateVolume creates a new volume with auto-calculated volume number
	CreateVolume(ctx context.Context, novelID uuid.UUID, title string, description, coverImageURL *string, displayOrder int, isPublished bool, createdBy uuid.UUID) (*domain.NovelVolume, error)

	// UpdateVolume updates volume information with history tracking
	UpdateVolume(ctx context.Context, id uuid.UUID, volumeNumber int, title string, description, coverImageURL *string, displayOrder int, isPublished bool, changedBy uuid.UUID, requestContext map[string]any) (*domain.NovelVolume, error)

	// DeleteVolume deletes a volume (soft delete)
	DeleteVolume(ctx context.Context, id uuid.UUID) error

	// GetVolumeByID retrieves a volume by ID
	GetVolumeByID(ctx context.Context, id uuid.UUID) (*domain.NovelVolume, error)

	// GetVolumesByNovelID retrieves all volumes for a novel
	GetVolumesByNovelID(ctx context.Context, novelID uuid.UUID, publishedOnly bool) ([]*domain.NovelVolume, error)

	// UpdateDisplayOrder updates the display order of a volume
	UpdateDisplayOrder(ctx context.Context, id uuid.UUID, order int) error

	// PublishVolume publishes a volume with history tracking
	PublishVolume(ctx context.Context, id uuid.UUID, changedBy uuid.UUID, requestContext map[string]any) error

	// UnpublishVolume unpublishes a volume with history tracking
	UnpublishVolume(ctx context.Context, id uuid.UUID, changedBy uuid.UUID, requestContext map[string]any) error
}
