package author

import (
	"context"
)

// deleteAuthorUseCase implements DeleteAuthorUseCase
type deleteAuthorUseCase struct {
	authorService AuthorService
}

// NewDeleteAuthorUseCase creates a new DeleteAuthorUseCase instance
func NewDeleteAuthorUseCase(authorService AuthorService) DeleteAuthorUseCase {
	return &deleteAuthorUseCase{
		authorService: authorService,
	}
}

// Execute deletes an author
func (uc *deleteAuthorUseCase) Execute(ctx context.Context, input DeleteAuthorInput) error {
	return uc.authorService.DeleteAuthor(ctx, input.ID)
}
