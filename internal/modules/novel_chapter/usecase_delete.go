package novel_chapter

import (
	"context"
)

type deleteChapterUseCase struct {
	chapterService ChapterService
}

func NewDeleteChapterUseCase(chapterService ChapterService) DeleteChapterUseCase {
	return &deleteChapterUseCase{
		chapterService: chapterService,
	}
}

func (uc *deleteChapterUseCase) Execute(ctx context.Context, input DeleteChapterInput) error {
	return uc.chapterService.DeleteChapter(ctx, input.ID)
}
