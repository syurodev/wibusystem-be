package novel_chapter

import (
	dtochapter "system/internal/dto/novel_chapter"
)

// Re-export types from dto package for backward compatibility
type ChapterResponse = dtochapter.ChapterResponse
type ChapterDetailResponse = dtochapter.ChapterDetailResponse
type ListChaptersResponse = dtochapter.ListChaptersResponse
type ChapterSummaryResponse = dtochapter.ChapterSummaryResponse
