package novel

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
)

// getNovelUseCase implements GetNovelUseCase
type getNovelUseCase struct {
	novelService NovelService
}

// NewGetNovelUseCase creates a new GetNovelUseCase instance
func NewGetNovelUseCase(novelService NovelService) GetNovelUseCase {
	return &getNovelUseCase{
		novelService: novelService,
	}
}

// Execute gets a novel by ID or slug
func (uc *getNovelUseCase) Execute(ctx context.Context, input GetNovelInput) (*domain.Novel, error) {
	// Try parsing as UUID first
	id, err := uuid.FromString(input.IDOrSlug)
	if err == nil {
		return uc.novelService.GetNovelByID(ctx, id)
	}

	// Otherwise treat as slug
	return uc.novelService.GetNovelBySlug(ctx, input.IDOrSlug)
}
