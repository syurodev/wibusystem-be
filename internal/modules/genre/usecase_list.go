package genre

import (
	"context"

	"system/internal/domain"
)

// ============================================================================
// ListGenresUseCase
// ============================================================================

// listGenresUseCase implements ListGenresUseCase
type listGenresUseCase struct {
	genreService GenreService
}

// NewListGenresUseCase creates a new ListGenresUseCase instance
func NewListGenresUseCase(genreService GenreService) ListGenresUseCase {
	return &listGenresUseCase{
		genreService: genreService,
	}
}

// Execute lists genres with pagination
func (uc *listGenresUseCase) Execute(ctx context.Context, input ListGenresInput) ([]*GenreWithTrend, int, error) {
	return uc.genreService.ListGenres(
		ctx,
		input.Page,
		input.Limit,
		input.Search,
		input.SortBy,
		input.SortOrder,
		input.ActiveOnly,
	)
}

// ============================================================================
// ListSelectionUseCase
// ============================================================================

// listSelectionUseCase implements ListSelectionUseCase
type listSelectionUseCase struct {
	genreService GenreService
}

// NewListSelectionUseCase creates a new ListSelectionUseCase instance
func NewListSelectionUseCase(genreService GenreService) ListSelectionUseCase {
	return &listSelectionUseCase{
		genreService: genreService,
	}
}

// Execute lists genres for selection dropdown
func (uc *listSelectionUseCase) Execute(ctx context.Context, input ListSelectionInput) ([]*domain.Genre, int, error) {
	return uc.genreService.ListSelection(ctx, input.Page, input.Limit, input.Search)
}
