package author

import (
	"context"

	"system/internal/domain"
)

// createAuthorUseCase implements CreateAuthorUseCase
type createAuthorUseCase struct {
	authorService AuthorService
}

// NewCreateAuthorUseCase creates a new CreateAuthorUseCase instance
func NewCreateAuthorUseCase(authorService AuthorService) CreateAuthorUseCase {
	return &createAuthorUseCase{
		authorService: authorService,
	}
}

// Execute creates a new author
func (uc *createAuthorUseCase) Execute(ctx context.Context, input CreateAuthorInput) (*domain.Author, error) {
	return uc.authorService.CreateAuthor(
		ctx,
		input.Name,
		input.Biography,
		input.AvatarURL,
		input.SocialLinksJSON,
		input.CreatedBy,
	)
}
