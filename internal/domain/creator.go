package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

// CreatorListFilter defines filter params for listing creators
type CreatorListFilter struct {
	Role                   *string    // Filter by role (default: CREATOR)
	Search                 *string    // Search display_name/username
	ViewPeriod             *string    // day, week, month, year, all
	Page                   int        // Default: 1
	Limit                  int        // Default: 20, max: 100
	SortBy                 string     // last_content_updated_at, total_views, follower_count
	SortOrder              string     // asc, desc (default: desc)
	CreatedFrom            *time.Time // Filter created_at from
	CreatedTo              *time.Time // Filter created_at to
	FirstContentPostedFrom *time.Time // Filter by first content posted date from
	FirstContentPostedTo   *time.Time // Filter by first content posted date to
}

// CreatorWithStats represents a creator with view statistics
type CreatorWithStats struct {
	User
	// Stats from user_statistics table
	FollowerCount        int        `json:"follower_count"`
	WorksCount           int        `json:"works_count"` // Computed: novel + manga + anime
	NovelCount           int        `json:"novel_count"`
	MangaCount           int        `json:"manga_count"`
	AnimeCount           int        `json:"anime_count"`
	LastContentUpdatedAt *time.Time `json:"last_content_updated_at"`
	// From ClickHouse
	TotalViews          int64   `json:"total_views"`
	PopularWorkID       *string `json:"popular_work_id"`
	PopularWorkTitle    *string `json:"popular_work_title"`
	PopularWorkCoverURL *string `json:"popular_work_cover_url"`
	PopularWorkType     *string `json:"popular_work_type"` // novel, manga, anime
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

	// IncrementNovelCount increments the novel_count for a user
	IncrementNovelCount(ctx context.Context, userID uuid.UUID) error

	// DecrementNovelCount decrements the novel_count for a user
	DecrementNovelCount(ctx context.Context, userID uuid.UUID) error

	// EnsureUserStatisticsExists ensures a user_statistics record exists
	EnsureUserStatisticsExists(ctx context.Context, userID uuid.UUID) error
}
