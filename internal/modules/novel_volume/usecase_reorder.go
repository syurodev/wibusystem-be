package novel_volume

import (
	"context"
)

// UpdateDisplayOrderUseCase impl
type updateDisplayOrderUseCase struct {
	volumeService VolumeService
}

func NewUpdateDisplayOrderUseCase(volumeService VolumeService) UpdateDisplayOrderUseCase {
	return &updateDisplayOrderUseCase{
		volumeService: volumeService,
	}
}

func (uc *updateDisplayOrderUseCase) Execute(ctx context.Context, input UpdateDisplayOrderInput) error {
	return uc.volumeService.UpdateDisplayOrder(ctx, input.ID, input.DisplayOrder)
}
