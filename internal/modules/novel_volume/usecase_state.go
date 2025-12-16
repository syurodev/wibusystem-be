package novel_volume

import (
	"context"
)

// PublishVolumeUseCase impl
type publishVolumeUseCase struct {
	volumeService VolumeService
}

func NewPublishVolumeUseCase(volumeService VolumeService) PublishVolumeUseCase {
	return &publishVolumeUseCase{
		volumeService: volumeService,
	}
}

func (uc *publishVolumeUseCase) Execute(ctx context.Context, input PublishVolumeInput) error {
	return uc.volumeService.PublishVolume(ctx, input.ID, input.ChangedBy, input.RequestContext)
}

// UnpublishVolumeUseCase impl
type unpublishVolumeUseCase struct {
	volumeService VolumeService
}

func NewUnpublishVolumeUseCase(volumeService VolumeService) UnpublishVolumeUseCase {
	return &unpublishVolumeUseCase{
		volumeService: volumeService,
	}
}

func (uc *unpublishVolumeUseCase) Execute(ctx context.Context, input UnpublishVolumeInput) error {
	return uc.volumeService.UnpublishVolume(ctx, input.ID, input.ChangedBy, input.RequestContext)
}
