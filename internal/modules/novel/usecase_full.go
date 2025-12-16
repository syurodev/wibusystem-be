package novel

import (
	"context"
)

// getNovelFullUseCase implements GetNovelFullUseCase
type getNovelFullUseCase struct {
	novelService NovelService
}

// NewGetNovelFullUseCase creates a new GetNovelFullUseCase instance
func NewGetNovelFullUseCase(novelService NovelService) GetNovelFullUseCase {
	return &getNovelFullUseCase{
		novelService: novelService,
	}
}

// Execute gets full novel details including volumes and chapters
func (uc *getNovelFullUseCase) Execute(ctx context.Context, input GetNovelFullInput) (*NovelFullData, error) {
	return uc.novelService.GetNovelFull(ctx, input.Slug)
}
