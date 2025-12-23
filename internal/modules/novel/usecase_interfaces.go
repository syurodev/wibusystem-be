package novel

import (
	"context"
	"encoding/json"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
)

// ============================================================================
// Input Types
// ============================================================================

// CreateNovelInput là input cho CreateNovelUseCase
type CreateNovelInput struct {
	Title            string
	Synopsis         json.RawMessage
	CoverImageURL    *string
	ThumbnailURL     *string
	Status           string
	IsOneshot        bool
	OriginalLanguage *string
	OriginalTitle    *string
	MetadataJSON     *string
	OwnerID          uuid.UUID
	OwnerType        string
	CreatedBy        uuid.UUID
	GenreIDs         []uuid.UUID
	AuthorIDs        []uuid.UUID
	ArtistIDs        []uuid.UUID
}

// UpdateNovelInput là input cho UpdateNovelUseCase
type UpdateNovelInput struct {
	ID               uuid.UUID
	Title            string
	Synopsis         json.RawMessage
	CoverImageURL    *string
	ThumbnailURL     *string
	Status           string
	IsOneshot        bool
	OriginalLanguage *string
	OriginalTitle    *string
	MetadataJSON     *string
	GenreIDs         []uuid.UUID
	AuthorIDs        []uuid.UUID
	ArtistIDs        []uuid.UUID
}

// DeleteNovelInput là input cho DeleteNovelUseCase
type DeleteNovelInput struct {
	ID uuid.UUID
}

// GetNovelInput là input cho GetNovelUseCase
type GetNovelInput struct {
	IDOrSlug string
}

// ListNovelsInput là input cho ListNovelsUseCase
type ListNovelsInput struct {
	Page             int
	Limit            int
	OwnerID          *uuid.UUID
	KeySearch        string
	GenreIDs         []uuid.UUID
	StatusStrs       []string
	OriginalLanguage string
	SortBy           string
	SortOrder        string
}

// IncrementViewCountInput là input cho IncrementViewCountUseCase
type IncrementViewCountInput struct {
	ID uuid.UUID
}

// GetNovelFullInput là input cho GetNovelFullUseCase
type GetNovelFullInput struct {
	Slug string
}

// ============================================================================
// UseCase Interfaces
// ============================================================================

// CreateNovelUseCase interface for creating novels
type CreateNovelUseCase interface {
	Execute(ctx context.Context, input CreateNovelInput) (*domain.Novel, error)
}

// UpdateNovelUseCase interface for updating novels
type UpdateNovelUseCase interface {
	Execute(ctx context.Context, input UpdateNovelInput) (*domain.Novel, error)
}

// DeleteNovelUseCase interface for deleting novels
type DeleteNovelUseCase interface {
	Execute(ctx context.Context, input DeleteNovelInput) error
}

// GetNovelUseCase interface for getting a single novel
type GetNovelUseCase interface {
	Execute(ctx context.Context, input GetNovelInput) (*domain.Novel, error)
}

// ListNovelsUseCase interface for listing novels with pagination
type ListNovelsUseCase interface {
	Execute(ctx context.Context, input ListNovelsInput) ([]*domain.Novel, int, error)
}

// IncrementViewCountUseCase interface for incrementing view count
type IncrementViewCountUseCase interface {
	Execute(ctx context.Context, input IncrementViewCountInput) error
}

// GetNovelFullUseCase interface for getting full novel details
type GetNovelFullUseCase interface {
	Execute(ctx context.Context, input GetNovelFullInput) (*NovelFullData, error)
}

// MediaRankStat represents rank statistics for a media item
// Matches domain.RankStat structure
type MediaRankStat struct {
	EntityID     uuid.UUID
	TotalViews   int
	CurrentRank  int
	PreviousRank *int
	RankChange   *int
}

// MediaRankResult represents a novel with its rank statistics
type MediaRankResult struct {
	Novel *domain.Novel
	Stats MediaRankStat
}

// TopNovelService defines interface for getting top novels
// Used to avoid import cycle with analytics module
type TopNovelService interface {
	GetTopNovelsWithRank(ctx context.Context, period string, offset int, limit int) ([]MediaRankResult, error)
}
