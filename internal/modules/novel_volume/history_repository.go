package novel_volume

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
)

// VolumeHistoryRepository interface for logging volume history
type VolumeHistoryRepository interface {
	LogUpdate(ctx context.Context, volumeID, novelID uuid.UUID, oldVolume, newVolume *domain.NovelVolume, changedBy uuid.UUID, requestContext map[string]any) error
	LogPublish(ctx context.Context, volumeID, novelID uuid.UUID, changedBy uuid.UUID, requestContext map[string]any) error
	LogUnpublish(ctx context.Context, volumeID, novelID uuid.UUID, changedBy uuid.UUID, requestContext map[string]any) error
	GetLatestVersion(ctx context.Context, volumeID uuid.UUID) (int, error)
}
