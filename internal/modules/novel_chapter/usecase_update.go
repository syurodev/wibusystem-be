package novel_chapter

import (
	"context"

	"system/internal/domain"
)

type updateChapterUseCase struct {
	chapterService ChapterService
}

func NewUpdateChapterUseCase(chapterService ChapterService) UpdateChapterUseCase {
	return &updateChapterUseCase{
		chapterService: chapterService,
	}
}

func (uc *updateChapterUseCase) Execute(ctx context.Context, input UpdateChapterInput) (*domain.NovelChapter, error) {
	return uc.chapterService.UpdateChapter(
		ctx,
		input.ID,
		input.VolumeID,
		input.ChapterNumber,
		input.Title,
		input.Content,
		input.WordCount,
		input.CharacterCount,
		input.AuthorNotes,
		input.IsFree,
		input.Price,
		input.Currency,
		input.Status,
		input.DisplayOrder,
		input.ScheduledAt,
		input.ChangedBy,
		input.RequestContext,
	)
}
