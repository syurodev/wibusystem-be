package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

// CreatorListFilter defines filter params for listing creators
type CreatorListFilter struct {
	Role        *string    // Filter by role (default: CREATOR)
	Search      *string    // Search display_name/username
	ViewPeriod  *string    // day, week, month, year, all
	Page        int        // Default: 1
	Limit       int        // Default: 20, max: 100
	SortBy      string     // last_content_updated_at, total_views, follower_count
	SortOrder   string     // asc, desc (default: desc)
	CreatedFrom *time.Time // Filter created_at from
	CreatedTo   *time.Time // Filter created_at to
}

// CreatorWithStats represents a creator with view statistics
type CreatorWithStats struct {
	User
	TotalViews          int64   // From ClickHouse
	PopularWorkID       *string // Novel/Manga/Anime with highest views
	PopularWorkTitle    *string
	PopularWorkCoverURL *string
	PopularWorkType     *string // novel, manga, anime
}

// CreatorListResult represents paginated creator list result
type CreatorListResult struct {
	Creators   []CreatorWithStats
	Total      int64
	Page       int
	Limit      int
	TotalPages int
}

// CreatorRepository defines the interface for creator data access
type CreatorRepository interface {
	// ListCreators returns paginated list of creators with filters
	ListCreators(ctx context.Context, filter CreatorListFilter) (*CreatorListResult, error)

	// UpdateLastContentUpdatedAt updates the last content update timestamp
	UpdateLastContentUpdatedAt(ctx context.Context, userID uuid.UUID) error

	// IncrementWorksCount increments the works_count for a user
	IncrementWorksCount(ctx context.Context, userID uuid.UUID) error

	// DecrementWorksCount decrements the works_count for a user
	DecrementWorksCount(ctx context.Context, userID uuid.UUID) error
}
