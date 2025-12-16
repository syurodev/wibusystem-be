package artist

import (
	"context"

	"system/internal/domain"
)

// ============================================================================
// MergeArtistsUseCase
// ============================================================================

// mergeArtistsUseCase implements MergeArtistsUseCase
type mergeArtistsUseCase struct {
	artistService ArtistService
}

// NewMergeArtistsUseCase creates a new MergeArtistsUseCase instance
func NewMergeArtistsUseCase(artistService ArtistService) MergeArtistsUseCase {
	return &mergeArtistsUseCase{
		artistService: artistService,
	}
}

// Execute merges source artists into target artist
func (uc *mergeArtistsUseCase) Execute(ctx context.Context, input MergeArtistsInput) error {
	return uc.artistService.MergeArtists(ctx, input.TargetID, input.SourceIDs, input.MergedBy)
}

// ============================================================================
// PreviewMergeUseCase
// ============================================================================

// previewMergeUseCase implements PreviewMergeUseCase
type previewMergeUseCase struct {
	artistService ArtistService
}

// NewPreviewMergeUseCase creates a new PreviewMergeUseCase instance
func NewPreviewMergeUseCase(artistService ArtistService) PreviewMergeUseCase {
	return &previewMergeUseCase{
		artistService: artistService,
	}
}

// Execute previews the result of merging artists
func (uc *previewMergeUseCase) Execute(ctx context.Context, input PreviewMergeInput) ([]*domain.Novel, error) {
	return uc.artistService.PreviewMergeArtists(ctx, input.TargetID, input.SourceIDs)
}
