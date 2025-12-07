package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"system/internal/domain"

	"go.uber.org/zap"
)

const (
	// HomeDataCacheKey is the Redis key for home page data
	HomeDataCacheKey = "public:home:data"
	// HomeDataCacheTTL is the TTL for home page data cache (10 minutes)
	HomeDataCacheTTL = 10 * time.Minute
)

// HomeData represents aggregated data for the home page
type HomeData struct {
	Hero     []map[string]any `json:"hero"`     // Top 6 trending for hero section
	Trending []map[string]any `json:"trending"` // Top 20 trending for trending section
	Creators []any            `json:"creators"` // Top 8 creators
	Genres   []any            `json:"genres"`   // Top 12 genres
}

// PublicService handles public-facing data aggregation with caching.
type PublicService struct {
	analyticsService *AnalyticsService
	creatorService   *CreatorService
	genreService     *GenreService
	cacheService     *CacheService
	logger           *zap.Logger
}

// NewPublicService creates a new PublicService instance.
func NewPublicService(
	analyticsService *AnalyticsService,
	creatorService *CreatorService,
	genreService *GenreService,
	cacheService *CacheService,
	logger *zap.Logger,
) *PublicService {
	return &PublicService{
		analyticsService: analyticsService,
		creatorService:   creatorService,
		genreService:     genreService,
		cacheService:     cacheService,
		logger:           logger,
	}
}

// GetHomeData retrieves all data needed for the home page.
// Uses caching to reduce load on ClickHouse.
func (s *PublicService) GetHomeData(ctx context.Context) (*HomeData, error) {
	// Try to get from cache first
	return GetOrSet(ctx, s.cacheService, HomeDataCacheKey, HomeDataCacheTTL, func() (*HomeData, error) {
		return s.fetchHomeData(ctx)
	})
}

// fetchHomeData fetches all home data using goroutines for parallel execution.
func (s *PublicService) fetchHomeData(ctx context.Context) (*HomeData, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	result := &HomeData{
		Hero:     []map[string]any{},
		Trending: []map[string]any{},
		Creators: []any{},
		Genres:   []any{},
	}

	// Error collection
	var errs []error

	// Fetch trending data (for both hero and trending sections)
	wg.Add(1)
	go func() {
		defer wg.Done()
		trending, err := s.analyticsService.GetTopTrending(ctx, "all", "week", 20)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			s.logger.Warn("failed to fetch trending data", zap.Error(err))
			errs = append(errs, fmt.Errorf("trending: %w", err))
			return
		}
		result.Trending = trending
		// Take first 6 for hero section
		if len(trending) > 6 {
			result.Hero = trending[:6]
		} else {
			result.Hero = trending
		}
	}()

	// Fetch creators
	wg.Add(1)
	go func() {
		defer wg.Done()
		defaultRole := "CREATOR"
		viewPeriod := "week"
		filter := domain.CreatorListFilter{
			Page:       1,
			Limit:      8,
			Role:       &defaultRole,
			ViewPeriod: &viewPeriod,
			SortBy:     "total_views",
			SortOrder:  "desc",
		}
		creatorsResult, err := s.creatorService.ListCreators(ctx, filter)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			s.logger.Warn("failed to fetch creators data", zap.Error(err))
			errs = append(errs, fmt.Errorf("creators: %w", err))
			return
		}
		// Convert to any slice for JSON serialization
		creators := make([]any, len(creatorsResult.Creators))
		for i, c := range creatorsResult.Creators {
			creators[i] = map[string]any{
				"id":                     c.User.ID,
				"display_name":           getValue(c.User.DisplayName),
				"username":               getValue(c.User.Username),
				"avatar_url":             getValue(c.User.AvatarURL),
				"follower_count":         c.User.FollowerCount,
				"works_count":            c.User.WorksCount,
				"total_views":            c.TotalViews,
				"is_verified":            c.User.IsVerified,
				"popular_work_id":        c.PopularWorkID,
				"popular_work_title":     c.PopularWorkTitle,
				"popular_work_cover_url": c.PopularWorkCoverURL,
			}
		}
		result.Creators = creators
	}()

	// Fetch genres
	wg.Add(1)
	go func() {
		defer wg.Done()
		genres, _, err := s.genreService.ListGenres(ctx, 1, 12, "", "series", "desc", true)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			s.logger.Warn("failed to fetch genres data", zap.Error(err))
			errs = append(errs, fmt.Errorf("genres: %w", err))
			return
		}
		// Convert to any slice for JSON serialization
		genresData := make([]any, len(genres))
		for i, g := range genres {
			genresData[i] = map[string]any{
				"id":             g.ID,
				"name":           g.Name,
				"slug":           g.Slug,
				"description":    g.Description,
				"novel_count":    g.NovelCount,
				"active_readers": g.ActiveReaders,
				"total_views":    g.TotalViews,
				"trend":          g.Trend,
			}
		}
		result.Genres = genresData
	}()

	// Wait for all goroutines to complete
	wg.Wait()

	// If all fetches failed, return error
	if len(errs) == 3 {
		return nil, fmt.Errorf("all home data fetches failed: %v", errs)
	}

	// Log partial failures but return partial data
	if len(errs) > 0 {
		s.logger.Warn("some home data fetches failed", zap.Int("error_count", len(errs)))
	}

	return result, nil
}

// InvalidateHomeCache invalidates the home data cache.
// Call this when content is updated.
func (s *PublicService) InvalidateHomeCache(ctx context.Context) error {
	return s.cacheService.Delete(ctx, HomeDataCacheKey)
}
