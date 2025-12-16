package novel

import (
	"context"

	"system/internal/domain"
)

// listNovelsUseCase implements ListNovelsUseCase
type listNovelsUseCase struct {
	novelService NovelService
}

// NewListNovelsUseCase creates a new ListNovelsUseCase instance
func NewListNovelsUseCase(novelService NovelService) ListNovelsUseCase {
	return &listNovelsUseCase{
		novelService: novelService,
	}
}

// Execute lists novels with pagination
func (uc *listNovelsUseCase) Execute(ctx context.Context, input ListNovelsInput) ([]*domain.Novel, int, error) {
	return uc.novelService.ListNovels(
		ctx,
		input.Page,
		input.Limit,
		input.OwnerID,
		input.KeySearch,
		input.GenreIDs,
		input.StatusStrs,
		input.OriginalLanguage,
		input.SortBy,
		input.SortOrder,
	)
}
