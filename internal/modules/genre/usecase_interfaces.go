package genre

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
)

// ============================================================================
// Input Types
// ============================================================================

// CreateGenreInput là input cho CreateGenreUseCase
type CreateGenreInput struct {
	Name        string
	Description string
	ParentID    *uuid.UUID
	UserID      uuid.UUID
}

// UpdateGenreInput là input cho UpdateGenreUseCase
type UpdateGenreInput struct {
	ID          uuid.UUID
	Name        string
	Description string
	ParentID    *uuid.UUID
	IsActive    bool
	UserID      uuid.UUID
}

// DeleteGenreInput là input cho DeleteGenreUseCase
type DeleteGenreInput struct {
	ID     uuid.UUID
	UserID uuid.UUID
}

// GetGenreInput là input cho GetGenreUseCase
type GetGenreInput struct {
	IDOrSlug string
}

// ListGenresInput là input cho ListGenresUseCase
type ListGenresInput struct {
	Page       int
	Limit      int
	Search     string
	SortBy     string
	SortOrder  string
	ActiveOnly bool
}

// ListSelectionInput là input cho ListSelectionUseCase
type ListSelectionInput struct {
	Page   int
	Limit  int
	Search string
}

// MergeGenresInput là input cho MergeGenresUseCase
type MergeGenresInput struct {
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

// CreateGenreUseCase interface for creating genres
type CreateGenreUseCase interface {
	Execute(ctx context.Context, input CreateGenreInput) (*domain.Genre, error)
}

// UpdateGenreUseCase interface for updating genres
type UpdateGenreUseCase interface {
	Execute(ctx context.Context, input UpdateGenreInput) (*domain.Genre, error)
}

// DeleteGenreUseCase interface for deleting genres
type DeleteGenreUseCase interface {
	Execute(ctx context.Context, input DeleteGenreInput) error
}

// GetGenreUseCase interface for getting a single genre
type GetGenreUseCase interface {
	Execute(ctx context.Context, input GetGenreInput) (*domain.Genre, error)
}

// ListGenresUseCase interface for listing genres with pagination
type ListGenresUseCase interface {
	Execute(ctx context.Context, input ListGenresInput) ([]*GenreWithTrend, int, error)
}

// ListSelectionUseCase interface for listing genres for selection dropdown
type ListSelectionUseCase interface {
	Execute(ctx context.Context, input ListSelectionInput) ([]*domain.Genre, int, error)
}

// MergeGenresUseCase interface for merging genres
type MergeGenresUseCase interface {
	Execute(ctx context.Context, input MergeGenresInput) error
}

// PreviewMergeUseCase interface for previewing merge results
type PreviewMergeUseCase interface {
	Execute(ctx context.Context, input PreviewMergeInput) ([]*domain.AffectedNovel, error)
}
