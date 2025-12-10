package analytics

import (
	"context"
	"fmt"
	"system/internal/domain"
	"system/internal/modules/novel"

	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"
)

// AnalyticsService handles analytics and trending logic.
type analyticsServiceImpl struct {
	viewAnalyticsRepo domain.ViewAnalyticsRepository
	novelService      novel.NovelService // Use Service instead of Repo
	logger            *zap.Logger
}

// NewAnalyticsService creates a new analytics service instance.
func NewService(
	viewAnalyticsRepo domain.ViewAnalyticsRepository,
	novelService novel.NovelService,
	logger *zap.Logger,
) *analyticsServiceImpl {
	return &analyticsServiceImpl{
		viewAnalyticsRepo: viewAnalyticsRepo,
		novelService:      novelService,
		logger:            logger,
	}
}

// TrendingItem represents a unified media item for trending list.
// This matches the structure expected by the frontend MediaSeriesSchema.
type TrendingItem struct {
	ID               uuid.UUID       `json:"id"`
	Title            string          `json:"title"`
	Slug             string          `json:"slug"`
	OriginalTitle    *string         `json:"original_title,omitempty"`
	OriginalLanguage *string         `json:"original_language,omitempty"`
	CoverURL         *string         `json:"cover_url,omitempty"`
	Type             string          `json:"type"`   // "novel", "manga", "anime"
	Status           string          `json:"status"` // "ongoing", "completed", etc.
	Rating           float64         `json:"rating"`
	Views            int64           `json:"views"`
	Favorites        int             `json:"favorites"`
	Genres           []*domain.Genre `json:"genres"`
	Owner            *domain.User    `json:"owner,omitempty"` // Simplified owner info
	// Add other fields as needed
}

// GetTopTrending retrieves top trending media.
// GetTopTrending retrieves top trending media.
func (s *analyticsServiceImpl) GetTopTrending(ctx context.Context, mediaType string, timeRange string, limit int) ([]map[string]any, error) {
	// 1. Get trending IDs from ClickHouse
	trendingData, err := s.viewAnalyticsRepo.GetTopTrending(ctx, mediaType, timeRange, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get trending data: %w", err)
	}

	if len(trendingData) == 0 {
		return s.getFallbackTrending(ctx, mediaType, limit)
	}

	return s.hydrateTrendingItems(ctx, trendingData)
}

func (s *analyticsServiceImpl) getFallbackTrending(ctx context.Context, mediaType string, limit int) ([]map[string]any, error) {
	// Currently only supports Novel fallback
	if mediaType == domain.MediaTypeNovel || mediaType == "all" || mediaType == "" {
		novels, _, err := s.novelService.ListNovels(ctx, 1, limit, nil, "", nil, nil, "", "views", "desc")
		if err != nil {
			return nil, fmt.Errorf("failed to get fallback novels: %w", err)
		}

		results := []map[string]any{}
		for _, n := range novels {
			results = append(results, s.mapNovelToMediaSeries(n, uint64(n.ViewCount)))
		}
		return results, nil
	}
	return []map[string]any{}, nil
}

func (s *analyticsServiceImpl) hydrateTrendingItems(ctx context.Context, trendingData []map[string]any) ([]map[string]any, error) {
	// 2. Collect IDs by type
	novelIDs := []uuid.UUID{}
	// mangaIDs := []uuid.UUID{}
	// animeIDs := []uuid.UUID{}

	// Map to preserve order and views
	type ItemMeta struct {
		Type  string
		Views uint64
	}
	itemMap := make(map[uuid.UUID]ItemMeta)
	orderedIDs := []uuid.UUID{}

	for _, item := range trendingData {
		mType, ok := item["media_type"].(string)
		if !ok {
			continue
		}
		mID, ok := item["media_id"].(uuid.UUID)
		if !ok {
			continue
		}
		views, _ := item["views"].(uint64)

		itemMap[mID] = ItemMeta{Type: mType, Views: views}
		orderedIDs = append(orderedIDs, mID)

		switch mType {
		case domain.MediaTypeNovel:
			novelIDs = append(novelIDs, mID)
		case domain.MediaTypeManga:
			// mangaIDs = append(mangaIDs, mID)
		case domain.MediaTypeAnime:
			// animeIDs = append(animeIDs, mID)
		}
	}

	// 3. Fetch details from NovelService
	novelMap := make(map[uuid.UUID]*domain.Novel)
	if len(novelIDs) > 0 {
		novels, err := s.novelService.GetNovelsByIDs(ctx, novelIDs)
		if err != nil {
			s.logger.Error("Failed to fetch trending novels", zap.Error(err))
			// Continue partial?
		} else {
			for _, n := range novels {
				novelMap[n.ID] = n
			}
		}
	}

	// 4. Assemble result
	results := []map[string]any{}
	for _, id := range orderedIDs {
		meta := itemMap[id]

		if meta.Type == domain.MediaTypeNovel {
			novel, found := novelMap[id]
			if found {
				results = append(results, s.mapNovelToMediaSeries(novel, meta.Views))
			}
		}
		// Handle other types when repos are available
	}

	return results, nil
}

// GetCreatorViewStats retrieves view statistics for multiple creators
func (s *analyticsServiceImpl) GetCreatorViewStats(ctx context.Context, ownerIDs []uuid.UUID, timeRange string) (map[uuid.UUID]domain.CreatorViewStats, error) {
	return s.viewAnalyticsRepo.GetCreatorViewStats(ctx, ownerIDs, timeRange)
}

func (s *analyticsServiceImpl) mapNovelToMediaSeries(n *domain.Novel, views uint64) map[string]any {
	// Map domain.Novel to MediaSeriesSchema format
	// Note: Owner info needs to be structured correctly
	owner := map[string]any{
		"id":           n.OwnerID,
		"display_name": "",
		"username":     "",
	}
	if n.OwnerDisplayName != nil {
		owner["display_name"] = *n.OwnerDisplayName
	}
	if n.OwnerUsername != nil {
		owner["username"] = *n.OwnerUsername
	}
	if n.OwnerAvatarURL != nil {
		owner["avatar_url"] = *n.OwnerAvatarURL
	}

	genres := []map[string]any{}
	for _, g := range n.Genres {
		genres = append(genres, map[string]any{
			"id":   g.ID,
			"name": g.Name,
		})
	}

	return map[string]any{
		"id":                n.ID,
		"title":             n.Title,
		"slug":              n.Slug,
		"original_title":    n.OriginalTitle,
		"original_language": n.OriginalLanguage,
		"cover_url":         n.CoverImageURL,
		"type":              domain.MediaTypeNovel,
		"status":            n.Status,
		"rating":            n.RatingAverage,

		// User asked for "Trending API", usually we show the views that made it trending (e.g. "10k views this week").
		// But MediaSeriesSchema has `views` field which is usually total views.
		// Let's use Total Views from Postgres for consistency with the object definition,
		// OR use the trending views if the UI expects to show "Trending Views".
		// Given `MediaSeriesSchema` is a generic schema, `views` likely means total views.
		// But for a trending list, showing total views might be misleading if sorting by weekly views.
		// However, `n.ViewCount` is total views.
		// Let's use `n.ViewCount` (total) to match the schema's semantic, but the list is ordered by trending views.
		// Wait, if I use `views` from arg, it's `uint64`. `n.ViewCount` is `int64`.
		// Let's use `n.ViewCount` for now.
		// "views":             n.ViewCount, 
		
		// Actually, let's use the trending views because that's what makes it interesting in this context?
		// No, let's stick to the entity's persistent data.
		"views":     n.ViewCount,
		"favorites": n.FavoriteCount,
		"genres":    genres,
		"owner":     owner,
		"created_at": n.CreatedAt,
		"updated_at": n.UpdatedAt,
		
		// Synopsis - JSONB content from database
		"synopsis": n.Synopsis,
	}
}
