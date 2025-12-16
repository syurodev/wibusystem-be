package novel_volume

import (
	"context"

	"system/internal/domain"
)

type updateVolumeUseCase struct {
	volumeService VolumeService
}

func NewUpdateVolumeUseCase(volumeService VolumeService) UpdateVolumeUseCase {
	return &updateVolumeUseCase{
		volumeService: volumeService,
	}
}

func (uc *updateVolumeUseCase) Execute(ctx context.Context, input UpdateVolumeInput) (*domain.NovelVolume, error) {
	return uc.volumeService.UpdateVolume(
		ctx,
		input.ID,
		input.VolumeNumber,
		input.Title,
		input.Description,
		input.CoverImageURL,
		input.DisplayOrder,
		input.IsPublished,
		input.ChangedBy,
		input.RequestContext,
	)
}
