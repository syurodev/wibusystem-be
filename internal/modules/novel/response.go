package novel

import (
	dtonovel "system/internal/dto/novel"
)

// Re-export types from dto package for backward compatibility
type OwnerInfo = dtonovel.OwnerInfo
type GenreInfo = dtonovel.GenreInfo
type LatestChapterInfo = dtonovel.LatestChapterInfo
type NovelResponse = dtonovel.NovelResponse
type NovelDetailResponse = dtonovel.NovelDetailResponse
type ChapterSummaryResponse = dtonovel.ChapterSummaryResponse
type VolumeInfoResponse = dtonovel.VolumeInfoResponseWithChapters
type NovelFullResponse = dtonovel.NovelFullResponse
type AffectedNovel = dtonovel.AffectedNovel
