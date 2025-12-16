package author

import (
	"context"

	"system/internal/domain"
)

// updateAuthorUseCase implements UpdateAuthorUseCase
type updateAuthorUseCase struct {
	authorService AuthorService
}

// NewUpdateAuthorUseCase creates a new UpdateAuthorUseCase instance
func NewUpdateAuthorUseCase(authorService AuthorService) UpdateAuthorUseCase {
	return &updateAuthorUseCase{
		authorService: authorService,
	}
}

// Execute updates an existing author
func (uc *updateAuthorUseCase) Execute(ctx context.Context, input UpdateAuthorInput) (*domain.Author, error) {
	return uc.authorService.UpdateAuthor(
		ctx,
		input.ID,
		input.Name,
		input.Biography,
		input.AvatarURL,
		input.SocialLinksJSON,
	)
}
