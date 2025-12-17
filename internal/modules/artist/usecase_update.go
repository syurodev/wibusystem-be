package artist

import (
	"context"

	"system/internal/domain"
)

// updateArtistUseCase implements UpdateArtistUseCase
type updateArtistUseCase struct {
	artistService ArtistService
}

// NewUpdateArtistUseCase creates a new UpdateArtistUseCase instance
func NewUpdateArtistUseCase(artistService ArtistService) UpdateArtistUseCase {
	return &updateArtistUseCase{
		artistService: artistService,
	}
}

// Execute updates an existing artist
func (uc *updateArtistUseCase) Execute(ctx context.Context, input UpdateArtistInput) (*domain.Artist, error) {
	return uc.artistService.UpdateArtist(
		ctx,
		input.ID,
		input.Name,
		input.Biography,
		input.AvatarURL,
		input.SocialLinksJSON,
		input.Specialization,
		input.PortfolioURL,
	)
}
