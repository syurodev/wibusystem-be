package artist

import (
	"context"

	"system/internal/domain"
)

// ============================================================================
// ListArtistsUseCase
// ============================================================================

// listArtistsUseCase implements ListArtistsUseCase
type listArtistsUseCase struct {
	artistService ArtistService
}

// NewListArtistsUseCase creates a new ListArtistsUseCase instance
func NewListArtistsUseCase(artistService ArtistService) ListArtistsUseCase {
	return &listArtistsUseCase{
		artistService: artistService,
	}
}

// Execute lists artists with pagination
func (uc *listArtistsUseCase) Execute(ctx context.Context, input ListArtistsInput) ([]*domain.Artist, int, error) {
	return uc.artistService.ListArtists(
		ctx,
		input.Page,
		input.Limit,
		input.Search,
		input.SortBy,
		input.SortOrder,
		input.Specialization,
		input.IsVerified,
	)
}

// ============================================================================
// ListSelectionUseCase
// ============================================================================

// listSelectionUseCase implements ListSelectionUseCase
type listSelectionUseCase struct {
	artistService ArtistService
}

// NewListSelectionUseCase creates a new ListSelectionUseCase instance
func NewListSelectionUseCase(artistService ArtistService) ListSelectionUseCase {
	return &listSelectionUseCase{
		artistService: artistService,
	}
}

// Execute lists artists for selection dropdown
func (uc *listSelectionUseCase) Execute(ctx context.Context, input ListSelectionInput) ([]*domain.Artist, int64, error) {
	return uc.artistService.ListSelection(ctx, input.Page, input.Limit, input.Search)
}
