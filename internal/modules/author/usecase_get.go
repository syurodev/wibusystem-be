package author

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
)

// getAuthorUseCase implements GetAuthorUseCase
type getAuthorUseCase struct {
	authorService AuthorService
}

// NewGetAuthorUseCase creates a new GetAuthorUseCase instance
func NewGetAuthorUseCase(authorService AuthorService) GetAuthorUseCase {
	return &getAuthorUseCase{
		authorService: authorService,
	}
}

// Execute gets an author by ID or slug
func (uc *getAuthorUseCase) Execute(ctx context.Context, input GetAuthorInput) (*domain.Author, error) {
	// Try parsing as UUID first
	id, err := uuid.FromString(input.IDOrSlug)
	if err == nil {
		return uc.authorService.GetAuthorByID(ctx, id)
	}

	// Otherwise treat as slug
	return uc.authorService.GetAuthorBySlug(ctx, input.IDOrSlug)
}
