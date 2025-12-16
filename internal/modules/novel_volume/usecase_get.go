package novel_volume

import (
	"context"

	"system/internal/domain"
)

type getVolumeUseCase struct {
	volumeService VolumeService
}

func NewGetVolumeUseCase(volumeService VolumeService) GetVolumeUseCase {
	return &getVolumeUseCase{
		volumeService: volumeService,
	}
}

func (uc *getVolumeUseCase) Execute(ctx context.Context, input GetVolumeInput) (*domain.NovelVolume, error) {
	return uc.volumeService.GetVolumeByID(ctx, input.ID)
}
