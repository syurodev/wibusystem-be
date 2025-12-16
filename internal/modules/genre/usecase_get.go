package genre

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
)

// getGenreUseCase implements GetGenreUseCase
type getGenreUseCase struct {
	genreService GenreService
}

// NewGetGenreUseCase creates a new GetGenreUseCase instance
func NewGetGenreUseCase(genreService GenreService) GetGenreUseCase {
	return &getGenreUseCase{
		genreService: genreService,
	}
}

// Execute gets a genre by ID or slug
func (uc *getGenreUseCase) Execute(ctx context.Context, input GetGenreInput) (*domain.Genre, error) {
	// Try parsing as UUID first
	id, err := uuid.FromString(input.IDOrSlug)
	if err == nil {
		return uc.genreService.GetGenreByID(ctx, id)
	}

	// Otherwise treat as slug
	return uc.genreService.GetGenreBySlug(ctx, input.IDOrSlug)
}
