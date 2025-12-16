package genre

import (
	"context"

	"system/internal/domain"
)

// createGenreUseCase implements CreateGenreUseCase
type createGenreUseCase struct {
	genreService GenreService
}

// NewCreateGenreUseCase creates a new CreateGenreUseCase instance
func NewCreateGenreUseCase(genreService GenreService) CreateGenreUseCase {
	return &createGenreUseCase{
		genreService: genreService,
	}
}

// Execute creates a new genre
func (uc *createGenreUseCase) Execute(ctx context.Context, input CreateGenreInput) (*domain.Genre, error) {
	return uc.genreService.CreateGenre(ctx, input.Name, input.Description, input.ParentID, input.UserID)
}
