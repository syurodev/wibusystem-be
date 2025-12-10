package novel_chapter

import (
	dtochapter "system/internal/dto/novel_chapter"
)

// Re-export request types from dto package
type CreateChapterRequest = dtochapter.CreateChapterRequest
type UpdateChapterRequest = dtochapter.UpdateChapterRequest
type ScheduleChapterRequest = dtochapter.ScheduleChapterRequest
type ListChaptersRequest = dtochapter.ListChaptersRequest
type UpdateStatisticsRequest = dtochapter.UpdateStatisticsRequest
