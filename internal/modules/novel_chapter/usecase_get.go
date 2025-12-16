package novel_chapter

import (
	"context"

	"system/internal/domain"
)

type getChapterUseCase struct {
	chapterService ChapterService
}

func NewGetChapterUseCase(chapterService ChapterService) GetChapterUseCase {
	return &getChapterUseCase{
		chapterService: chapterService,
	}
}

func (uc *getChapterUseCase) Execute(ctx context.Context, input GetChapterInput) (*domain.NovelChapter, error) {
	return uc.chapterService.GetChapterByID(ctx, input.ID)
}
