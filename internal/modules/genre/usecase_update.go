package genre

import (
	"context"

	"system/internal/domain"
)

// updateGenreUseCase implements UpdateGenreUseCase
type updateGenreUseCase struct {
	genreService GenreService
}

// NewUpdateGenreUseCase creates a new UpdateGenreUseCase instance
func NewUpdateGenreUseCase(genreService GenreService) UpdateGenreUseCase {
	return &updateGenreUseCase{
		genreService: genreService,
	}
}

// Execute updates an existing genre
func (uc *updateGenreUseCase) Execute(ctx context.Context, input UpdateGenreInput) (*domain.Genre, error) {
	return uc.genreService.UpdateGenre(
		ctx,
		input.ID,
		input.Name,
		input.Description,
		input.ParentID,
		input.IsActive,
		input.UserID,
	)
}
