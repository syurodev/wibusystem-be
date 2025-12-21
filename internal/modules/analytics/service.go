package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"system/internal/domain"
	"system/internal/modules/novel"
	"system/internal/platform/database"

	"github.com/gofrs/uuid/v5"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// AnalyticsService handles analytics and trending logic.
// AnalyticsService handles analytics and trending logic.
type analyticsServiceImpl struct {
	viewAnalyticsRepo domain.ViewAnalyticsRepository
	novelService      novel.NovelService // Keep NovelService for complex hydration logic if safe
	creatorRepo       domain.CreatorRepository // Use Repo to avoid cycle
	orgRepo           domain.OrganizationRepository
	chapterRepo       domain.NovelChapterRepository
	userRepo          domain.UserRepository
	genreRepo         domain.GenreRepository
	redisClient       *database.RedisClient
	logger            *zap.Logger
}

// NewService creates a new analytics service instance.
func NewService(
	viewAnalyticsRepo domain.ViewAnalyticsRepository,
	novelService novel.NovelService,
	creatorRepo domain.CreatorRepository,
	orgRepo domain.OrganizationRepository,
	chapterRepo domain.NovelChapterRepository,
	userRepo domain.UserRepository,
	genreRepo domain.GenreRepository,
	redisClient *database.RedisClient,
	logger *zap.Logger,
) *analyticsServiceImpl {
	return &analyticsServiceImpl{
		viewAnalyticsRepo: viewAnalyticsRepo,
		novelService:      novelService,
		creatorRepo:       creatorRepo,
		orgRepo:           orgRepo,
		chapterRepo:       chapterRepo,
		userRepo:          userRepo,
		genreRepo:         genreRepo,
		redisClient:       redisClient,
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
		"synopsis":          n.Synopsis,
	}
}

// GetMostActiveCreators retrieves creators with most activity (uploads)
func (s *analyticsServiceImpl) GetMostActiveCreators(ctx context.Context, limit int) ([]domain.CreatorActivityStat, error) {
	// 1. Get stats from ClickHouse (last 30 days default)
	stats, err := s.viewAnalyticsRepo.GetTopActiveCreators(ctx, "month", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get active creators stats: %w", err)
	}

	if len(stats) == 0 {
		return []domain.CreatorActivityStat{}, nil
	}

	// 2. Hydrate user info
	for i, stat := range stats {
		user, err := s.userRepo.GetByID(ctx, stat.UserID)
		if err != nil {
			s.logger.Warn("Failed to hydrate user for active creator", zap.String("userID", stat.UserID.String()), zap.Error(err))
			continue
		}
		// Note: User struct in CreatorActivityStat is *domain.User
		stat.User = user
		stats[i] = stat
	}

	return stats, nil
}

// GetActiveOrganizations retrieves organizations with most activity
func (s *analyticsServiceImpl) GetActiveOrganizations(ctx context.Context, limit int) ([]domain.OrgActivityStat, error) {
	// 1. Get stats from ClickHouse
	stats, err := s.viewAnalyticsRepo.GetTopActiveOrgs(ctx, "month", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get active orgs stats: %w", err)
	}

	if len(stats) == 0 {
		return []domain.OrgActivityStat{}, nil
	}

	// 2. Hydrate org info
	for i, stat := range stats {
		org, err := s.orgRepo.GetByID(ctx, stat.OrgID)
		if err != nil {
			s.logger.Warn("Failed to hydrate org for active org", zap.String("orgID", stat.OrgID.String()), zap.Error(err))
			continue
		}
		stat.Organization = org
		stats[i] = stat
	}

	return stats, nil
}

// GetRisingStars retrieves new creators with high view counts
func (s *analyticsServiceImpl) GetRisingStars(ctx context.Context, limit int) ([]domain.CreatorActivityStat, error) {
	// 1. Get creators who started in the last 90 days
	since := time.Now().AddDate(0, 0, -90)
	filter := domain.CreatorListFilter{
		FirstContentPostedFrom: &since,
		Limit:                  limit * 2, // Fetch more candidates
		SortBy:                 "total_views", // Note: Sorting by views here relies on PG data if available, or just fetch recent ones
	}

	creatorsResult, err := s.creatorRepo.ListCreators(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list rising creators: %w", err)
	}

	if len(creatorsResult.Creators) == 0 {
		return []domain.CreatorActivityStat{}, nil
	}

	// 2. Get accurate view stats from ClickHouse for these creators
	creatorIDs := make([]uuid.UUID, len(creatorsResult.Creators))
	userMap := make(map[uuid.UUID]*domain.User)
	for i, c := range creatorsResult.Creators {
		creatorIDs[i] = c.User.ID
		// Copy user to heap to take address? 
		// c.User is domain.User struct.
		u := c.User
		userMap[c.User.ID] = &u
	}

	viewStats, err := s.viewAnalyticsRepo.GetCreatorViewStats(ctx, creatorIDs, "month")
	if err != nil {
		return nil, fmt.Errorf("failed to get creator view stats: %w", err)
	}

	// 3. Construct result and sort
	var results []domain.CreatorActivityStat
	for _, id := range creatorIDs {
		stat, ok := viewStats[id]
		views := int64(0)
		if ok {
			views = stat.TotalViews
		}

		results = append(results, domain.CreatorActivityStat{
			UserID:        id,
			TotalActivity: views, // Reuse TotalActivity field to store TotalViews for this API
			TotalWeight:   views,
			User:          userMap[id],
		})
	}

	// Sort by Views (TotalWeight) descending
	for i := 0; i < len(results)-1; i++ {
		for j := 0; j < len(results)-i-1; j++ {
			if results[j].TotalWeight < results[j+1].TotalWeight {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}

	if len(results) > limit {
		results = results[:limit]
	}

	return results, nil
}

// GetFreshUpdates retrieves recently published chapters
func (s *analyticsServiceImpl) GetFreshUpdates(ctx context.Context, limit int) ([]domain.NovelChapterSummary, error) {
	// Call ChapterRepo
	chapters, err := s.chapterRepo.GetRecentChapters(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent chapters: %w", err)
	}
	
	// Convert to summary
	var summaries []domain.NovelChapterSummary
	for _, c := range chapters {
		summary := domain.NovelChapterSummary{
			ID:            c.ID,
			NovelID:       c.NovelID,
			VolumeID:      c.VolumeID,
			ChapterNumber: c.ChapterNumber,
			Title:         c.Title,
			Slug:          c.Slug,
			WordCount:     c.WordCount,
			IsFree:        c.IsFree,
			Price:         c.Price,
			Currency:      c.Currency,
			Status:        c.Status,
			ViewCount:     c.ViewCount,
			PublishedAt:   c.PublishedAt,
			CreatedAt:     c.CreatedAt,
			UpdatedAt:     c.UpdatedAt,
		}
		summaries = append(summaries, summary)
	}
	
	return summaries, nil
}

// GetTopGenresByViews retrieves genres with most views for a time range.
// Results are cached in Redis for 30 minutes.
// offset: 0 = current period, 1 = previous period
func (s *analyticsServiceImpl) GetTopGenresByViews(ctx context.Context, period string, offset int, limit int) ([]*domain.Genre, error) {
	// 1. Build cache key (including offset)
	cacheKey := fmt.Sprintf("analytics:top_genres_by_views:%s:%d:%d", period, offset, limit)

	// 2. Try to get from cache
	cached, err := s.redisClient.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		// Cache hit - deserialize
		var genres []*domain.Genre
		if err := json.Unmarshal([]byte(cached), &genres); err == nil {
			s.logger.Debug("Cache hit for top genres by views", zap.String("key", cacheKey))
			return genres, nil
		}
	} else if err != nil && err != redis.Nil {
		s.logger.Warn("Failed to get cache for top genres", zap.Error(err))
	}

	// 3. Cache miss - query ClickHouse
	stats, err := s.viewAnalyticsRepo.GetTopGenresByViews(ctx, period, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top genres by views: %w", err)
	}

	// 4. Hydrate genre details
	genres := make([]*domain.Genre, 0, len(stats))
	for _, stat := range stats {
		genre, err := s.genreRepo.GetByID(ctx, stat.GenreID)
		if err != nil {
			s.logger.Warn("Failed to hydrate genre",
				zap.String("genreID", stat.GenreID.String()),
				zap.Error(err))
			continue
		}
		// Override TotalViews with the time-range views from ClickHouse
		genre.TotalViews = int64(stat.TotalViews)
		genre.ActiveReaders = int64(stat.UniqueUsers)
		genres = append(genres, genre)
	}

	// 5. Store in cache (30 minutes)
	if len(genres) > 0 {
		data, err := json.Marshal(genres)
		if err == nil {
			if cacheErr := s.redisClient.Set(ctx, cacheKey, string(data), 30*time.Minute); cacheErr != nil {
				s.logger.Warn("Failed to cache top genres", zap.Error(cacheErr))
			}
		}
	}

	return genres, nil
}

// GetTopCreatorsByViews retrieves creators with most views for a time range.
// Results are cached in Redis for 30 minutes.
func (s *analyticsServiceImpl) GetTopCreatorsByViews(ctx context.Context, period string, offset int, limit int) ([]*domain.User, error) {
	cacheKey := fmt.Sprintf("analytics:top_creators_by_views:%s:%d:%d", period, offset, limit)

	// Try cache
	cached, err := s.redisClient.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		var users []*domain.User
		if err := json.Unmarshal([]byte(cached), &users); err == nil {
			s.logger.Debug("Cache hit for top creators by views", zap.String("key", cacheKey))
			return users, nil
		}
	} else if err != nil && err != redis.Nil {
		s.logger.Warn("Failed to get cache for top creators", zap.Error(err))
	}

	// Query ClickHouse
	stats, err := s.viewAnalyticsRepo.GetTopCreatorsByViews(ctx, period, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top creators by views: %w", err)
	}

	// Hydrate user details
	users := make([]*domain.User, 0, len(stats))
	for _, stat := range stats {
		user, err := s.userRepo.GetByID(ctx, stat.CreatorID)
		if err != nil {
			s.logger.Warn("Failed to hydrate creator", zap.String("creatorID", stat.CreatorID.String()), zap.Error(err))
			continue
		}
		users = append(users, user)
	}

	// Cache
	if len(users) > 0 {
		if data, err := json.Marshal(users); err == nil {
			if cacheErr := s.redisClient.Set(ctx, cacheKey, string(data), 30*time.Minute); cacheErr != nil {
				s.logger.Warn("Failed to cache top creators", zap.Error(cacheErr))
			}
		}
	}

	return users, nil
}

// GetTopOrgsByViews retrieves organizations with most views for a time range.
// Results are cached in Redis for 30 minutes.
func (s *analyticsServiceImpl) GetTopOrgsByViews(ctx context.Context, period string, offset int, limit int) ([]*domain.Organization, error) {
	cacheKey := fmt.Sprintf("analytics:top_orgs_by_views:%s:%d:%d", period, offset, limit)

	// Try cache
	cached, err := s.redisClient.Get(ctx, cacheKey)
	if err == nil && cached != "" {
		var orgs []*domain.Organization
		if err := json.Unmarshal([]byte(cached), &orgs); err == nil {
			s.logger.Debug("Cache hit for top orgs by views", zap.String("key", cacheKey))
			return orgs, nil
		}
	} else if err != nil && err != redis.Nil {
		s.logger.Warn("Failed to get cache for top orgs", zap.Error(err))
	}

	// Query ClickHouse
	stats, err := s.viewAnalyticsRepo.GetTopOrgsByViews(ctx, period, offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top orgs by views: %w", err)
	}

	// Hydrate org details
	orgs := make([]*domain.Organization, 0, len(stats))
	for _, stat := range stats {
		org, err := s.orgRepo.GetByID(ctx, stat.OrgID)
		if err != nil {
			s.logger.Warn("Failed to hydrate org", zap.String("orgID", stat.OrgID.String()), zap.Error(err))
			continue
		}
		orgs = append(orgs, org)
	}

	// Cache
	if len(orgs) > 0 {
		if data, err := json.Marshal(orgs); err == nil {
			if cacheErr := s.redisClient.Set(ctx, cacheKey, string(data), 30*time.Minute); cacheErr != nil {
				s.logger.Warn("Failed to cache top orgs", zap.Error(cacheErr))
			}
		}
	}

	return orgs, nil
}

// GetTopGenresWithRankComparison retrieves top genres with rank comparison
func (s *analyticsServiceImpl) GetTopGenresWithRankComparison(ctx context.Context, period string, limit int) ([]GenreRankResponse, error) {
	stats, err := s.viewAnalyticsRepo.GetRankWithComparison(ctx, period, "genre", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get genre rank stats: %w", err)
	}

	results := make([]GenreRankResponse, 0, len(stats))
	for _, stat := range stats {
		genre, err := s.genreRepo.GetByID(ctx, stat.EntityID)
		if err != nil {
			s.logger.Warn("Failed to hydrate genre for rank", zap.String("id", stat.EntityID.String()), zap.Error(err))
			continue
		}
		// Hydrate stats into genre object
		genre.TotalViews = stat.TotalViews
		genre.ActiveReaders = stat.UniqueUsers
		
		results = append(results, GenreRankResponse{
			Genre: genre,
			Stats: stat,
		})
	}
	return results, nil
}

// GetTopCreatorsWithRankComparison retrieves top creators with rank comparison
func (s *analyticsServiceImpl) GetTopCreatorsWithRankComparison(ctx context.Context, period string, limit int) ([]CreatorRankResponse, error) {
	stats, err := s.viewAnalyticsRepo.GetRankWithComparison(ctx, period, "creator", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get creator rank stats: %w", err)
	}

	results := make([]CreatorRankResponse, 0, len(stats))
	for _, stat := range stats {
		user, err := s.userRepo.GetByID(ctx, stat.EntityID)
		if err != nil {
			s.logger.Warn("Failed to hydrate creator for rank", zap.String("id", stat.EntityID.String()), zap.Error(err))
			continue
		}
		
		results = append(results, CreatorRankResponse{
			User:  user,
			Stats: stat,
		})
	}
	return results, nil
}

// GetTopOrgsWithRankComparison retrieves top orgs with rank comparison
func (s *analyticsServiceImpl) GetTopOrgsWithRankComparison(ctx context.Context, period string, limit int) ([]OrgRankResponse, error) {
	stats, err := s.viewAnalyticsRepo.GetRankWithComparison(ctx, period, "org", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get org rank stats: %w", err)
	}

	results := make([]OrgRankResponse, 0, len(stats))
	for _, stat := range stats {
		org, err := s.orgRepo.GetByID(ctx, stat.EntityID)
		if err != nil {
			s.logger.Warn("Failed to hydrate org for rank", zap.String("id", stat.EntityID.String()), zap.Error(err))
			continue
		}

		results = append(results, OrgRankResponse{
			Organization: org,
			Stats:        stat,
		})
	}
	return results, nil
}

// GetTopMediaWithRankComparison retrieves top media (novel/manga/anime) with rank comparison
func (s *analyticsServiceImpl) GetTopMediaWithRankComparison(ctx context.Context, period string, mediaType string, limit int) ([]MediaRankResponse, error) {
	// Validate mediaType
	if mediaType != "novel" && mediaType != "manga" && mediaType != "anime" {
		return nil, fmt.Errorf("invalid media type: %s", mediaType)
	}

	stats, err := s.viewAnalyticsRepo.GetRankWithComparison(ctx, period, mediaType, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get media rank stats: %w", err)
	}

	results := make([]MediaRankResponse, 0, len(stats))
	
	// Hydrate based on type
	// Currently only Novel hydration is supported
	if mediaType == "novel" {
		for _, stat := range stats {
			novel, err := s.novelService.GetNovelByID(ctx, stat.EntityID)
			if err != nil {
				s.logger.Warn("Failed to hydrate novel for rank", zap.String("id", stat.EntityID.String()), zap.Error(err))
				continue
			}

			// Map details
			authors := make([]*domain.Author, 0, len(novel.Authors))
			for _, na := range novel.Authors {
				if na.Author != nil {
					authors = append(authors, na.Author)
				}
			}

			resp := MediaRankResponse{
				ID:    novel.ID,
				Title: novel.Title,
				Slug:  novel.Slug,
				Type:  "novel",
				Stats: stat,
				Authors: authors,
			}
			if novel.CoverImageURL != nil {
				resp.Cover = *novel.CoverImageURL
			}
			
			results = append(results, resp)
		}
	} else {
		// For others, return basic info if possible or empty details with stats
		// Since we can't hydrate, we might just return stats with empty details?
		// Or skip?
		// Better to return what we have (ID and rank)
		for _, stat := range stats {
			results = append(results, MediaRankResponse{
				ID:    stat.EntityID,
				Type:  mediaType,
				Stats: stat,
				Title: "Unknown (Hydration not implemented)",
			})
		}
	}

	return results, nil
}


