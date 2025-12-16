package artist

import (
	"context"

	"system/internal/domain"
)

// createArtistUseCase implements CreateArtistUseCase
type createArtistUseCase struct {
	artistService ArtistService
}

// NewCreateArtistUseCase creates a new CreateArtistUseCase instance
func NewCreateArtistUseCase(artistService ArtistService) CreateArtistUseCase {
	return &createArtistUseCase{
		artistService: artistService,
	}
}

// Execute creates a new artist
func (uc *createArtistUseCase) Execute(ctx context.Context, input CreateArtistInput) (*domain.Artist, error) {
	return uc.artistService.CreateArtist(
		ctx,
		input.Name,
		input.Biography,
		input.AvatarURL,
		input.SocialLinksJSON,
		input.Specialization,
		input.CreatedBy,
	)
}
