package novel_volume

import (
	dtovolume "system/internal/dto/novel_volume"
)

// Re-export request types from dto package
type CreateVolumeRequest = dtovolume.CreateVolumeRequest
type UpdateVolumeRequest = dtovolume.UpdateVolumeRequest
type UpdateDisplayOrderRequest = dtovolume.UpdateDisplayOrderRequest
type ListVolumesRequest = dtovolume.ListVolumesRequest
