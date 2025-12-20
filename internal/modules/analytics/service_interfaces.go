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
}
