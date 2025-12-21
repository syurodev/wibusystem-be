package analytics

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
)

// AnalyticsService interface định nghĩa business logic cho analytics
// Cung cấp API để lấy dữ liệu trending, thống kê views cho tất cả media types (anime, manga, novel)
type AnalyticsService interface {
	// GetTopTrending retrieves top trending media from all types
	GetTopTrending(ctx context.Context, mediaType string, timeRange string, limit int) ([]map[string]any, error)
	
	// GetCreatorViewStats retrieves view statistics for multiple creators
	// GetCreatorViewStats retrieves view statistics for multiple creators
	GetCreatorViewStats(ctx context.Context, ownerIDs []uuid.UUID, timeRange string) (map[uuid.UUID]domain.CreatorViewStats, error)

	// GetMostActiveCreators retrieves creators with most activity (uploads)
	GetMostActiveCreators(ctx context.Context, limit int) ([]domain.CreatorActivityStat, error)

	// GetActiveOrganizations retrieves organizations with most activity
	GetActiveOrganizations(ctx context.Context, limit int) ([]domain.OrgActivityStat, error)

	// GetRisingStars retrieves new creators with high view counts
	GetRisingStars(ctx context.Context, limit int) ([]domain.CreatorActivityStat, error)

	// GetFreshUpdates retrieves recently published chapters
	GetFreshUpdates(ctx context.Context, limit int) ([]domain.NovelChapterSummary, error)

	// GetTopGenresByViews retrieves genres with most views for a time range
	// Returns hydrated genre data similar to ListGenres response
	// offset: 0 = current period, 1 = previous period
	GetTopGenresByViews(ctx context.Context, period string, offset int, limit int) ([]*domain.Genre, error)

	// GetTopCreatorsByViews retrieves creators with most views for a time range
	// Returns User objects with TotalViews populated
	GetTopCreatorsByViews(ctx context.Context, period string, offset int, limit int) ([]*domain.User, error)

	// GetTopOrgsByViews retrieves organizations with most views for a time range
	GetTopOrgsByViews(ctx context.Context, period string, offset int, limit int) ([]*domain.Organization, error)

	// GetTopGenresWithRankComparison retrieves top genres with rank comparison (week/month/year)
	GetTopGenresWithRankComparison(ctx context.Context, period string, limit int) ([]GenreRankResponse, error)

	// GetTopCreatorsWithRankComparison retrieves top creators with rank comparison
	GetTopCreatorsWithRankComparison(ctx context.Context, period string, limit int) ([]CreatorRankResponse, error)

	// GetTopOrgsWithRankComparison retrieves top orgs with rank comparison
	GetTopOrgsWithRankComparison(ctx context.Context, period string, limit int) ([]OrgRankResponse, error)

	// GetTopMediaWithRankComparison retrieves top media (novel/manga/anime) with rank comparison
	GetTopMediaWithRankComparison(ctx context.Context, period string, mediaType string, limit int) ([]MediaRankResponse, error)
}

// GenreRankResponse wraps Genre with rank info
type GenreRankResponse struct {
	*domain.Genre
	Stats domain.RankStat
}

// CreatorRankResponse wraps User with rank info
type CreatorRankResponse struct {
	*domain.User
	Stats domain.RankStat
}

// OrgRankResponse wraps Organization with rank info
type OrgRankResponse struct {
	*domain.Organization
	Stats domain.RankStat
}

// MediaRankResponse wraps Novel/Manga/Anime with rank info
// Since we don't have a common Media struct yet (or it's complex), we'll return generic map or specific structs
// For now, let's use a simplified struct holding the Entity details
type MediaRankResponse struct {
	ID          uuid.UUID
	Title       string
	Cover       string
	Slug        string
	Type        string // novel, manga, anime
	Stats       domain.RankStat
	Authors     []*domain.Author
	Artist      []*domain.Artist
}

