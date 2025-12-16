package novel_chapter

import (
	"context"

	"system/internal/domain"
)

// ListChaptersByNovelUseCase impl
type listChaptersByNovelUseCase struct {
	chapterService ChapterService
}

func NewListChaptersByNovelUseCase(chapterService ChapterService) ListChaptersByNovelUseCase {
	return &listChaptersByNovelUseCase{
		chapterService: chapterService,
	}
}

func (uc *listChaptersByNovelUseCase) Execute(ctx context.Context, input ListChaptersByNovelInput) ([]*domain.NovelChapter, error) {
	return uc.chapterService.GetChaptersByNovelID(ctx, input.NovelID, input.Filter)
}

// ListChaptersByVolumeUseCase impl
type listChaptersByVolumeUseCase struct {
	chapterService ChapterService
}

func NewListChaptersByVolumeUseCase(chapterService ChapterService) ListChaptersByVolumeUseCase {
	return &listChaptersByVolumeUseCase{
		chapterService: chapterService,
	}
}

func (uc *listChaptersByVolumeUseCase) Execute(ctx context.Context, input ListChaptersByVolumeInput) ([]*domain.NovelChapter, error) {
	return uc.chapterService.GetChaptersByVolumeID(ctx, input.VolumeID, input.PublishedOnly)
}
