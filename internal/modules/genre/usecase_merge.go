package genre

import (
	"context"

	"system/internal/domain"
)

// ============================================================================
// MergeGenresUseCase
// ============================================================================

// mergeGenresUseCase implements MergeGenresUseCase
type mergeGenresUseCase struct {
	genreService GenreService
}

// NewMergeGenresUseCase creates a new MergeGenresUseCase instance
func NewMergeGenresUseCase(genreService GenreService) MergeGenresUseCase {
	return &mergeGenresUseCase{
		genreService: genreService,
	}
}

// Execute merges source genres into target genre
func (uc *mergeGenresUseCase) Execute(ctx context.Context, input MergeGenresInput) error {
	return uc.genreService.MergeGenres(ctx, input.TargetID, input.SourceIDs, input.MergedBy)
}

// ============================================================================
// PreviewMergeUseCase
// ============================================================================

// previewMergeUseCase implements PreviewMergeUseCase
type previewMergeUseCase struct {
	genreService GenreService
}

// NewPreviewMergeUseCase creates a new PreviewMergeUseCase instance
func NewPreviewMergeUseCase(genreService GenreService) PreviewMergeUseCase {
	return &previewMergeUseCase{
		genreService: genreService,
	}
}

// Execute previews the result of merging genres
func (uc *previewMergeUseCase) Execute(ctx context.Context, input PreviewMergeInput) ([]*domain.AffectedNovel, error) {
	return uc.genreService.PreviewMergeGenres(ctx, input.TargetID, input.SourceIDs)
}
