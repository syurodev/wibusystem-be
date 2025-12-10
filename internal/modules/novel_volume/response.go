package novel_volume

import (
	dtovolume "system/internal/dto/novel_volume"
)

// Re-export types from dto package for backward compatibility
type VolumeResponse = dtovolume.VolumeResponse
type VolumeDetailResponse = dtovolume.VolumeDetailResponse
type VolumeInfoResponse = dtovolume.VolumeInfoResponse
type VolumeInfoResponseWithChapters = dtovolume.VolumeInfoResponseWithChapters
