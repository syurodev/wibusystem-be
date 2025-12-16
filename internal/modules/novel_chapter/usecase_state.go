package novel_chapter

import (
	"context"
)

// PublishChapterUseCase impl
type publishChapterUseCase struct {
	chapterService ChapterService
}

func NewPublishChapterUseCase(chapterService ChapterService) PublishChapterUseCase {
	return &publishChapterUseCase{
		chapterService: chapterService,
	}
}

func (uc *publishChapterUseCase) Execute(ctx context.Context, input PublishChapterInput) error {
	return uc.chapterService.PublishChapter(ctx, input.ID, input.ChangedBy, input.RequestContext)
}

// ScheduleChapterUseCase impl
type scheduleChapterUseCase struct {
	chapterService ChapterService
}

func NewScheduleChapterUseCase(chapterService ChapterService) ScheduleChapterUseCase {
	return &scheduleChapterUseCase{
		chapterService: chapterService,
	}
}

func (uc *scheduleChapterUseCase) Execute(ctx context.Context, input ScheduleChapterInput) error {
	return uc.chapterService.ScheduleChapter(ctx, input.ID, input.ScheduledAt)
}
