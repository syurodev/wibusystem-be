package author

import (
	"context"

	"system/internal/domain"
)

// ============================================================================
// MergeAuthorsUseCase
// ============================================================================

// mergeAuthorsUseCase implements MergeAuthorsUseCase
type mergeAuthorsUseCase struct {
	authorService AuthorService
}

// NewMergeAuthorsUseCase creates a new MergeAuthorsUseCase instance
func NewMergeAuthorsUseCase(authorService AuthorService) MergeAuthorsUseCase {
	return &mergeAuthorsUseCase{
		authorService: authorService,
	}
}

// Execute merges source authors into target author
func (uc *mergeAuthorsUseCase) Execute(ctx context.Context, input MergeAuthorsInput) error {
	return uc.authorService.MergeAuthors(ctx, input.TargetID, input.SourceIDs, input.MergedBy)
}

// ============================================================================
// PreviewMergeUseCase
// ============================================================================

// previewMergeUseCase implements PreviewMergeUseCase
type previewMergeUseCase struct {
	authorService AuthorService
}

// NewPreviewMergeUseCase creates a new PreviewMergeUseCase instance
func NewPreviewMergeUseCase(authorService AuthorService) PreviewMergeUseCase {
	return &previewMergeUseCase{
		authorService: authorService,
	}
}

// Execute previews the result of merging authors
func (uc *previewMergeUseCase) Execute(ctx context.Context, input PreviewMergeInput) ([]*domain.Novel, error) {
	return uc.authorService.PreviewMergeAuthors(ctx, input.TargetID, input.SourceIDs)
}
