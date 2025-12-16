package author

import (
	"context"

	"system/internal/domain"
)

// ============================================================================
// ListAuthorsUseCase
// ============================================================================

// listAuthorsUseCase implements ListAuthorsUseCase
type listAuthorsUseCase struct {
	authorService AuthorService
}

// NewListAuthorsUseCase creates a new ListAuthorsUseCase instance
func NewListAuthorsUseCase(authorService AuthorService) ListAuthorsUseCase {
	return &listAuthorsUseCase{
		authorService: authorService,
	}
}

// Execute lists authors with pagination
func (uc *listAuthorsUseCase) Execute(ctx context.Context, input ListAuthorsInput) ([]*domain.Author, int, error) {
	return uc.authorService.ListAuthors(
		ctx,
		input.Page,
		input.Limit,
		input.Search,
		input.SortBy,
		input.SortOrder,
		input.IsVerified,
	)
}

// ============================================================================
// ListSelectionUseCase
// ============================================================================

// listSelectionUseCase implements ListSelectionUseCase
type listSelectionUseCase struct {
	authorService AuthorService
}

// NewListSelectionUseCase creates a new ListSelectionUseCase instance
func NewListSelectionUseCase(authorService AuthorService) ListSelectionUseCase {
	return &listSelectionUseCase{
		authorService: authorService,
	}
}

// Execute lists authors for selection dropdown
func (uc *listSelectionUseCase) Execute(ctx context.Context, input ListSelectionInput) ([]*domain.Author, int64, error) {
	return uc.authorService.ListSelection(ctx, input.Page, input.Limit, input.Search)
}
