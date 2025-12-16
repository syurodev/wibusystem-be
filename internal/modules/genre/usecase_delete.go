package genre

import (
	"context"
)

// deleteGenreUseCase implements DeleteGenreUseCase
type deleteGenreUseCase struct {
	genreService GenreService
}

// NewDeleteGenreUseCase creates a new DeleteGenreUseCase instance
func NewDeleteGenreUseCase(genreService GenreService) DeleteGenreUseCase {
	return &deleteGenreUseCase{
		genreService: genreService,
	}
}

// Execute deletes a genre
func (uc *deleteGenreUseCase) Execute(ctx context.Context, input DeleteGenreInput) error {
	return uc.genreService.DeleteGenre(ctx, input.ID, input.UserID)
}
