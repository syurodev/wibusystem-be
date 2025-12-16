package artist

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
)

// ============================================================================
// Input Types
// ============================================================================

// CreateArtistInput là input cho CreateArtistUseCase
type CreateArtistInput struct {
	Name            string
	Biography       string
	AvatarURL       *string
	SocialLinksJSON *string
	Specialization  *string
	CreatedBy       uuid.UUID
}

// UpdateArtistInput là input cho UpdateArtistUseCase
type UpdateArtistInput struct {
	ID              uuid.UUID
	Name            string
	Biography       string
	AvatarURL       *string
	SocialLinksJSON *string
	Specialization  *string
}

// DeleteArtistInput là input cho DeleteArtistUseCase
type DeleteArtistInput struct {
	ID uuid.UUID
}

// GetArtistInput là input cho GetArtistUseCase
type GetArtistInput struct {
	IDOrSlug string
}

// ListArtistsInput là input cho ListArtistsUseCase
type ListArtistsInput struct {
	Page           int
	Limit          int
	Search         string
	SortBy         string
	SortOrder      string
	Specialization *string
	IsVerified     *bool
}

// ListSelectionInput là input cho ListSelectionUseCase
type ListSelectionInput struct {
	Page   int
	Limit  int
	Search string
}

// MergeArtistsInput là input cho MergeArtistsUseCase
type MergeArtistsInput struct {
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

// CreateArtistUseCase interface for creating artists
type CreateArtistUseCase interface {
	Execute(ctx context.Context, input CreateArtistInput) (*domain.Artist, error)
}

// UpdateArtistUseCase interface for updating artists
type UpdateArtistUseCase interface {
	Execute(ctx context.Context, input UpdateArtistInput) (*domain.Artist, error)
}

// DeleteArtistUseCase interface for deleting artists
type DeleteArtistUseCase interface {
	Execute(ctx context.Context, input DeleteArtistInput) error
}

// GetArtistUseCase interface for getting a single artist
type GetArtistUseCase interface {
	Execute(ctx context.Context, input GetArtistInput) (*domain.Artist, error)
}

// ListArtistsUseCase interface for listing artists with pagination
type ListArtistsUseCase interface {
	Execute(ctx context.Context, input ListArtistsInput) ([]*domain.Artist, int, error)
}

// ListSelectionUseCase interface for listing artists for selection dropdown
type ListSelectionUseCase interface {
	Execute(ctx context.Context, input ListSelectionInput) ([]*domain.Artist, int64, error)
}

// MergeArtistsUseCase interface for merging artists
type MergeArtistsUseCase interface {
	Execute(ctx context.Context, input MergeArtistsInput) error
}

// PreviewMergeUseCase interface for previewing merge results
type PreviewMergeUseCase interface {
	Execute(ctx context.Context, input PreviewMergeInput) ([]*domain.Novel, error)
}
