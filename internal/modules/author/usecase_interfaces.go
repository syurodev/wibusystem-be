package author

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
)

// ============================================================================
// Input Types
// ============================================================================

// CreateAuthorInput là input cho CreateAuthorUseCase
type CreateAuthorInput struct {
	Name            string
	Biography       string
	AvatarURL       *string
	SocialLinksJSON *string
	CreatedBy       uuid.UUID
}

// UpdateAuthorInput là input cho UpdateAuthorUseCase
type UpdateAuthorInput struct {
	ID              uuid.UUID
	Name            string
	Biography       string
	AvatarURL       *string
	SocialLinksJSON *string
}

// DeleteAuthorInput là input cho DeleteAuthorUseCase
type DeleteAuthorInput struct {
	ID uuid.UUID
}

// GetAuthorInput là input cho GetAuthorUseCase
type GetAuthorInput struct {
	IDOrSlug string
}

// ListAuthorsInput là input cho ListAuthorsUseCase
type ListAuthorsInput struct {
	Page       int
	Limit      int
	Search     string
	SortBy     string
	SortOrder  string
	IsVerified *bool
}

// ListSelectionInput là input cho ListSelectionUseCase
type ListSelectionInput struct {
	Page   int
	Limit  int
	Search string
}

// MergeAuthorsInput là input cho MergeAuthorsUseCase
type MergeAuthorsInput struct {
	TargetID  uuid.UUID
	SourceIDs []uuid.UUID
	MergedBy  uuid.UUID
}

// PreviewMergeInput là input cho PreviewMergeUseCase
type PreviewMergeInput struct {
	TargetID  uuid.UUID
	SourceIDs []uuid.UUID
}

// ============================================================================
// UseCase Interfaces
// ============================================================================

// CreateAuthorUseCase interface for creating authors
type CreateAuthorUseCase interface {
	Execute(ctx context.Context, input CreateAuthorInput) (*domain.Author, error)
}

// UpdateAuthorUseCase interface for updating authors
type UpdateAuthorUseCase interface {
	Execute(ctx context.Context, input UpdateAuthorInput) (*domain.Author, error)
}

// DeleteAuthorUseCase interface for deleting authors
type DeleteAuthorUseCase interface {
	Execute(ctx context.Context, input DeleteAuthorInput) error
}

// GetAuthorUseCase interface for getting a single author
type GetAuthorUseCase interface {
	Execute(ctx context.Context, input GetAuthorInput) (*domain.Author, error)
}

// ListAuthorsUseCase interface for listing authors with pagination
type ListAuthorsUseCase interface {
	Execute(ctx context.Context, input ListAuthorsInput) ([]*domain.Author, int, error)
}

// ListSelectionUseCase interface for listing authors for selection dropdown
type ListSelectionUseCase interface {
	Execute(ctx context.Context, input ListSelectionInput) ([]*domain.Author, int64, error)
}

// MergeAuthorsUseCase interface for merging authors
type MergeAuthorsUseCase interface {
	Execute(ctx context.Context, input MergeAuthorsInput) error
}

// PreviewMergeUseCase interface for previewing merge results
type PreviewMergeUseCase interface {
	Execute(ctx context.Context, input PreviewMergeInput) ([]*domain.Novel, error)
}
