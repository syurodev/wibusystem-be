package creator

import (
	"context"
	"fmt"

	"system/internal/domain"
	"system/internal/modules/analytics"
	"system/internal/modules/novel"

	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"
)

// CreatorService handles creator-related business logic
type creatorServiceImpl struct {
	creatorRepo      domain.CreatorRepository
	analyticsService analytics.AnalyticsService
	novelService     novel.NovelService
	logger           *zap.Logger
}

// NewCreatorService creates a new CreatorService instance
func NewService(
	creatorRepo domain.CreatorRepository,
	analyticsService analytics.AnalyticsService,
	novelService novel.NovelService,
	logger *zap.Logger,
) *creatorServiceImpl {
	return &creatorServiceImpl{
		creatorRepo:      creatorRepo,
		analyticsService: analyticsService,
		novelService:     novelService,
		logger:           logger,
	}
}

// ListCreators returns paginated list of creators with view stats and popular work info
func (s *creatorServiceImpl) ListCreators(ctx context.Context, filter domain.CreatorListFilter) (*domain.CreatorListResult, error) {
	// 1. Get creators from PostgreSQL
	result, err := s.creatorRepo.ListCreators(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list creators: %w", err)
	}

	if len(result.Creators) == 0 {
		return result, nil
	}

	// 2. Get user IDs for ClickHouse query
	userIDs := make([]uuid.UUID, len(result.Creators))
	for i, c := range result.Creators {
		userIDs[i] = c.User.ID
	}

	// 3. Get view stats and popular work IDs
	viewStats := s.getCreatorViewStats(ctx, userIDs, filter.ViewPeriod)
	
	popularWorkIDs := make([]uuid.UUID, 0)
	for _, stats := range viewStats {
		if stats.PopularWorkID != nil {
			popularWorkIDs = append(popularWorkIDs, *stats.PopularWorkID)
		}
	}

	// 4. Fetch novel metadata for popular works
	novelInfoMap := s.getPopularWorkMetadata(ctx, popularWorkIDs)

	// 5. Merge view stats and popular work info into creators
	for i := range result.Creators {
		userID := result.Creators[i].User.ID
		if stats, ok := viewStats[userID]; ok {
			result.Creators[i].TotalViews = stats.TotalViews

			if stats.PopularWorkID != nil {
				idStr := stats.PopularWorkID.String()
				result.Creators[i].PopularWorkID = &idStr
				result.Creators[i].PopularWorkType = &stats.PopularWorkType

				// Get title and cover from novel info
				if info, found := novelInfoMap[*stats.PopularWorkID]; found {
					result.Creators[i].PopularWorkTitle = &info.Title
					result.Creators[i].PopularWorkCoverURL = &info.CoverURL
				}
			}
		}
	}

	return result, nil
}

func (s *creatorServiceImpl) getCreatorViewStats(ctx context.Context, userIDs []uuid.UUID, filterPeriod *string) map[uuid.UUID]domain.CreatorViewStats {
	viewPeriod := "all"
	if filterPeriod != nil {
		viewPeriod = *filterPeriod
	}

	viewStats, err := s.analyticsService.GetCreatorViewStats(ctx, userIDs, viewPeriod)
	if err != nil {
		s.logger.Warn("failed to get creator view stats", zap.Error(err))
		return make(map[uuid.UUID]domain.CreatorViewStats)
	}
	return viewStats
}

func (s *creatorServiceImpl) getPopularWorkMetadata(ctx context.Context, popularWorkIDs []uuid.UUID) map[uuid.UUID]struct {
	Title    string
	CoverURL string
} {
	novelInfoMap := make(map[uuid.UUID]struct {
		Title    string
		CoverURL string
	})

	if len(popularWorkIDs) > 0 {
		// Used GetNovelsByIDs instead of loop
		novels, err := s.novelService.GetNovelsByIDs(ctx, popularWorkIDs)
		if err == nil {
			for _, novel := range novels {
				novelInfoMap[novel.ID] = struct {
					Title    string
					CoverURL string
				}{
					Title:    novel.Title,
					CoverURL: getValue(novel.CoverImageURL),
				}
			}
		} else {
			s.logger.Warn("Failed to fetch popular work metadata", zap.Error(err))
		}
	}
	return novelInfoMap
}

