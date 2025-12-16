package artist

import (
	"context"
)

// deleteArtistUseCase implements DeleteArtistUseCase
type deleteArtistUseCase struct {
	artistService ArtistService
}

// NewDeleteArtistUseCase creates a new DeleteArtistUseCase instance
func NewDeleteArtistUseCase(artistService ArtistService) DeleteArtistUseCase {
	return &deleteArtistUseCase{
		artistService: artistService,
	}
}

// Execute deletes an artist
func (uc *deleteArtistUseCase) Execute(ctx context.Context, input DeleteArtistInput) error {
	return uc.artistService.DeleteArtist(ctx, input.ID)
}
