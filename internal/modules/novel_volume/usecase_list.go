package novel_volume

import (
	"context"

	"system/internal/domain"
)

type listVolumesByNovelUseCase struct {
	volumeService VolumeService
}

func NewListVolumesByNovelUseCase(volumeService VolumeService) ListVolumesByNovelUseCase {
	return &listVolumesByNovelUseCase{
		volumeService: volumeService,
	}
}

func (uc *listVolumesByNovelUseCase) Execute(ctx context.Context, input ListVolumesByNovelInput) ([]*domain.NovelVolume, error) {
	return uc.volumeService.GetVolumesByNovelID(ctx, input.NovelID, input.PublishedOnly)
}
