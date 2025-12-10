package media

import (
	dtomedia "system/internal/dto/media"
)

// Re-export types from dto package for backward compatibility
type MediaSeriesResponse = dtomedia.MediaSeriesResponse
type GenreInfo = dtomedia.GenreInfo
type OwnerInfo = dtomedia.OwnerInfo
type LatestChapterInfo = dtomedia.LatestChapterInfo
type HomeData = dtomedia.HomeData
