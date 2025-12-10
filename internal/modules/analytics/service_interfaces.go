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
	GetCreatorViewStats(ctx context.Context, ownerIDs []uuid.UUID, timeRange string) (map[uuid.UUID]domain.CreatorViewStats, error)
}
