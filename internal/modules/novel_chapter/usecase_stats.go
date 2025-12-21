package novel_chapter

import (
	"context"
)

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
