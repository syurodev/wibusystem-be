package media

import (
	"context"
	"fmt"
	"sync"
	"time"

	"system/internal/domain"
	mediadto "system/internal/dto/media"
	analytics_module "system/internal/modules/analytics"
	creator_module "system/internal/modules/creator"
	"system/internal/platform/cache"

	"go.uber.org/zap"
)

const (
	// HomeDataCacheKey is the Redis key for home page data
	HomeDataCacheKey = "public:home:data"
	// HomeDataCacheTTL is the TTL for home page data cache (10 minutes)
	HomeDataCacheTTL = 10 * time.Minute
)

// mediaServiceImpl handles public-facing media data aggregation with caching.
// Cung cấp API aggregated cho tất cả media types (anime, manga, novel).
type mediaServiceImpl struct {
	analyticsService analytics_module.AnalyticsService
	creatorService   creator_module.CreatorService
	cacheService     *cache.CacheService
	logger           *zap.Logger
}

// NewMediaService creates a new MediaService instance.
func NewMediaService(
	analyticsService analytics_module.AnalyticsService,
	creatorService creator_module.CreatorService,
	cacheService *cache.CacheService,
	logger *zap.Logger,
) MediaService {
	return &mediaServiceImpl{
		analyticsService: analyticsService,
		creatorService:   creatorService,
		cacheService:     cacheService,
		logger:           logger,
	}
}

// GetHomeData retrieves all data needed for the home page.
// Uses caching to reduce load on ClickHouse.
func (s *mediaServiceImpl) GetHomeData(ctx context.Context) (*mediadto.HomeData, error) {
	// Try to get from cache first
	return cache.GetOrSet(ctx, s.cacheService, HomeDataCacheKey, HomeDataCacheTTL, func() (*mediadto.HomeData, error) {
		return s.fetchHomeData(ctx)
	})
}

// fetchHomeData fetches all home data using goroutines for parallel execution.
func (s *mediaServiceImpl) fetchHomeData(ctx context.Context) (*mediadto.HomeData, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	result := &mediadto.HomeData{
		Hero:                []map[string]any{},
		Trending:            []map[string]any{},
		Creators:            []any{},
		Genres:              []any{},
		MostActiveCreators:  []any{},
		RisingStars:         []any{},
		ActiveOrganizations: []any{},
		FreshUpdates:        []any{},
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
				"follower_count":         c.FollowerCount,
				"works_count":            c.WorksCount,
				"total_views":            c.TotalViews,
				"is_verified":            c.User.IsVerified,
				"popular_work_id":        c.PopularWorkID,
				"popular_work_title":     c.PopularWorkTitle,
				"popular_work_cover_url": c.PopularWorkCoverURL,
			}
		}
		result.Creators = creators
	}()

	// Fetch Most Active Creators
	wg.Add(1)
	go func() {
		defer wg.Done()
		data, err := s.analyticsService.GetMostActiveCreators(ctx, 10)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			s.logger.Warn("failed to fetch most active creators", zap.Error(err))
			errs = append(errs, fmt.Errorf("active_creators: %w", err))
			return
		}
		
		output := make([]any, len(data))
		for i, d := range data {
			user := d.User
			output[i] = map[string]any{
				"id":            user.ID,
				"display_name":  getValue(user.DisplayName),
				"username":      getValue(user.Username),
				"avatar_url":    getValue(user.AvatarURL),
				"is_verified":   user.IsVerified,
				"created_at":    user.CreatedAt,
				"total_weight":  d.TotalWeight,
			}
		}
		result.MostActiveCreators = output
	}()

	// Fetch Rising Stars
	wg.Add(1)
	go func() {
		defer wg.Done()
		data, err := s.analyticsService.GetRisingStars(ctx, 10)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			s.logger.Warn("failed to fetch rising stars", zap.Error(err))
			errs = append(errs, fmt.Errorf("rising_stars: %w", err))
			return
		}
		
		output := make([]any, len(data))
		for i, d := range data {
			user := d.User
			output[i] = map[string]any{
				"id":            user.ID,
				"display_name":  getValue(user.DisplayName),
				"username":      getValue(user.Username),
				"avatar_url":    getValue(user.AvatarURL),
				"is_verified":   user.IsVerified,
				"created_at":    user.CreatedAt,
				"total_views":   d.TotalWeight, // Assuming TotalWeight stores views for Rising Stars
			}
		}
		result.RisingStars = output
	}()

	// Fetch Active Organizations
	wg.Add(1)
	go func() {
		defer wg.Done()
		data, err := s.analyticsService.GetActiveOrganizations(ctx, 10)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			s.logger.Warn("failed to fetch active orgs", zap.Error(err))
			errs = append(errs, fmt.Errorf("active_orgs: %w", err))
			return
		}
		
		output := make([]any, len(data))
		for i, d := range data {
			org := d.Organization
			output[i] = map[string]any{
				"id":            org.ID,
				"name":          org.Name,
				"slug":          org.Slug,
				"avatar_url":    getValue(org.AvatarURL),
				"member_count":  org.MemberCount,
				"total_weight":  d.TotalActivity,
			}
		}
		result.ActiveOrganizations = output
	}()

	// Fetch Fresh Updates
	wg.Add(1)
	go func() {
		defer wg.Done()
		data, err := s.analyticsService.GetFreshUpdates(ctx, 10)
		mu.Lock()
		defer mu.Unlock()
		if err != nil {
			s.logger.Warn("failed to fetch fresh updates", zap.Error(err))
			errs = append(errs, fmt.Errorf("fresh_updates: %w", err))
			return
		}
		
		// Map domain.NovelChapterSummary to generic map (or specialized DTO if we want)
		// For consistency with other responses, let's map it.
		output := make([]any, len(data))
		for i, c := range data {
			output[i] = map[string]any{
				"id":             c.ID,
				"novel_id":       c.NovelID,
				"volume_id":      c.VolumeID,
				"chapter_number": c.ChapterNumber,
				"title":          c.Title,
				"slug":           c.Slug,
				"word_count":     c.WordCount,
				"published_at":   getTimeValue(c.PublishedAt),
				"created_at":     c.CreatedAt,
			}
		}
		result.FreshUpdates = output
	}()

	// TODO: Genres temporarily disabled - will be moved to genre module API later
	// result.Genres remains empty []any{}

	// Wait for all goroutines to complete
	wg.Wait()

	// If all fetches failed, return error
	if len(errs) == 6 {
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
func (s *mediaServiceImpl) InvalidateHomeCache(ctx context.Context) error {
	return s.cacheService.Delete(ctx, HomeDataCacheKey)
}

// getValue is a helper to dereference string pointers safely
func getValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// getTimeValue safely gets string formatted time from pointer
func getTimeValue(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
