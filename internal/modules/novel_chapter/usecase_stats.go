package novel_chapter

import (
	"context"
)

// IncrementViewCountUseCase impl
type incrementViewCountUseCase struct {
	chapterService ChapterService
}

func NewIncrementViewCountUseCase(chapterService ChapterService) IncrementViewCountUseCase {
	return &incrementViewCountUseCase{
		chapterService: chapterService,
	}
}

func (uc *incrementViewCountUseCase) Execute(ctx context.Context, input IncrementViewCountInput) error {
	return uc.chapterService.IncrementViewCount(ctx, input.ID)
}

// UpdateStatisticsUseCase impl
type updateStatisticsUseCase struct {
	chapterService ChapterService
}

func NewUpdateStatisticsUseCase(chapterService ChapterService) UpdateStatisticsUseCase {
	return &updateStatisticsUseCase{
		chapterService: chapterService,
	}
}

func (uc *updateStatisticsUseCase) Execute(ctx context.Context, input UpdateStatisticsInput) error {
	return uc.chapterService.UpdateStatistics(ctx, input.ID, input.Stats)
}
