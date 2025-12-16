package novel_volume

import (
	"context"

	"system/internal/domain"
)

type createVolumeUseCase struct {
	volumeService VolumeService
}

func NewCreateVolumeUseCase(volumeService VolumeService) CreateVolumeUseCase {
	return &createVolumeUseCase{
		volumeService: volumeService,
	}
}

func (uc *createVolumeUseCase) Execute(ctx context.Context, input CreateVolumeInput) (*domain.NovelVolume, error) {
	return uc.volumeService.CreateVolume(
		ctx,
		input.NovelID,
		input.Title,
		input.Description,
		input.CoverImageURL,
		input.DisplayOrder,
		input.IsPublished,
		input.CreatedBy,
	)
}
