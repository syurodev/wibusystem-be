package novel_volume

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
)

// CreateVolumeInput represents input for creating a volume
type CreateVolumeInput struct {
	NovelID       uuid.UUID
	Title         string
	Description   *string
	CoverImageURL *string
	DisplayOrder  int
	IsPublished   bool
	CreatedBy     uuid.UUID
}

// UpdateVolumeInput represents input for updating a volume
type UpdateVolumeInput struct {
	ID             uuid.UUID
	VolumeNumber   int
	Title          string
	Description    *string
	CoverImageURL  *string
	DisplayOrder   int
	IsPublished    bool
	ChangedBy      uuid.UUID
	RequestContext map[string]any
}

// DeleteVolumeInput represents input for deleting a volume
type DeleteVolumeInput struct {
	ID uuid.UUID
}

// GetVolumeInput represents input for getting a volume
type GetVolumeInput struct {
	ID uuid.UUID
}

// ListVolumesByNovelInput represents input for listing volumes
type ListVolumesByNovelInput struct {
	NovelID       uuid.UUID
	PublishedOnly bool
}

// UpdateDisplayOrderInput represents input for updating display order
type UpdateDisplayOrderInput struct {
	ID           uuid.UUID
	DisplayOrder int
}

// PublishVolumeInput represents input for publishing a volume
type PublishVolumeInput struct {
	ID             uuid.UUID
	ChangedBy      uuid.UUID
	RequestContext map[string]any
}

// UnpublishVolumeInput represents input for unpublishing a volume
type UnpublishVolumeInput struct {
	ID             uuid.UUID
	ChangedBy      uuid.UUID
	RequestContext map[string]any
}

// CreateVolumeUseCase interface
type CreateVolumeUseCase interface {
	Execute(ctx context.Context, input CreateVolumeInput) (*domain.NovelVolume, error)
}

// UpdateVolumeUseCase interface
type UpdateVolumeUseCase interface {
	Execute(ctx context.Context, input UpdateVolumeInput) (*domain.NovelVolume, error)
}

// DeleteVolumeUseCase interface
type DeleteVolumeUseCase interface {
	Execute(ctx context.Context, input DeleteVolumeInput) error
}

// GetVolumeUseCase interface
type GetVolumeUseCase interface {
	Execute(ctx context.Context, input GetVolumeInput) (*domain.NovelVolume, error)
}

// ListVolumesByNovelUseCase interface
type ListVolumesByNovelUseCase interface {
	Execute(ctx context.Context, input ListVolumesByNovelInput) ([]*domain.NovelVolume, error)
}

// UpdateDisplayOrderUseCase interface
type UpdateDisplayOrderUseCase interface {
	Execute(ctx context.Context, input UpdateDisplayOrderInput) error
}

// PublishVolumeUseCase interface
type PublishVolumeUseCase interface {
	Execute(ctx context.Context, input PublishVolumeInput) error
}

// UnpublishVolumeUseCase interface
type UnpublishVolumeUseCase interface {
	Execute(ctx context.Context, input UnpublishVolumeInput) error
}
