package novel

import (
	"context"
)

// incrementViewCountUseCase implements IncrementViewCountUseCase
type incrementViewCountUseCase struct {
	novelService NovelService
}

// NewIncrementViewCountUseCase creates a new IncrementViewCountUseCase instance
func NewIncrementViewCountUseCase(novelService NovelService) IncrementViewCountUseCase {
	return &incrementViewCountUseCase{
		novelService: novelService,
	}
}

// Execute increments the view count of a novel
func (uc *incrementViewCountUseCase) Execute(ctx context.Context, input IncrementViewCountInput) error {
	return uc.novelService.IncrementViewCount(ctx, input.ID)
}
