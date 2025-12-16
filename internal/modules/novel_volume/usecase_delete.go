package novel_volume

import (
	"context"
)

type deleteVolumeUseCase struct {
	volumeService VolumeService
}

func NewDeleteVolumeUseCase(volumeService VolumeService) DeleteVolumeUseCase {
	return &deleteVolumeUseCase{
		volumeService: volumeService,
	}
}

func (uc *deleteVolumeUseCase) Execute(ctx context.Context, input DeleteVolumeInput) error {
	return uc.volumeService.DeleteVolume(ctx, input.ID)
}
