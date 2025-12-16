package artist

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
)

// getArtistUseCase implements GetArtistUseCase
type getArtistUseCase struct {
	artistService ArtistService
}

// NewGetArtistUseCase creates a new GetArtistUseCase instance
func NewGetArtistUseCase(artistService ArtistService) GetArtistUseCase {
	return &getArtistUseCase{
		artistService: artistService,
	}
}

// Execute gets an artist by ID or slug
func (uc *getArtistUseCase) Execute(ctx context.Context, input GetArtistInput) (*domain.Artist, error) {
	// Try parsing as UUID first
	id, err := uuid.FromString(input.IDOrSlug)
	if err == nil {
		return uc.artistService.GetArtistByID(ctx, id)
	}

	// Otherwise treat as slug
	return uc.artistService.GetArtistBySlug(ctx, input.IDOrSlug)
}
