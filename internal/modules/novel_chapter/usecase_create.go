package novel_chapter

import (
	"context"

	"system/internal/domain"
)

type createChapterUseCase struct {
	chapterService ChapterService
}

func NewCreateChapterUseCase(chapterService ChapterService) CreateChapterUseCase {
	return &createChapterUseCase{
		chapterService: chapterService,
	}
}

func (uc *createChapterUseCase) Execute(ctx context.Context, input CreateChapterInput) (*domain.NovelChapter, error) {
	return uc.chapterService.CreateChapter(
		ctx,
		input.NovelID,
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
		input.CreatedBy,
	)
}
