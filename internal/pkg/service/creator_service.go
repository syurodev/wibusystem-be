package service

import (
	"context"
	"fmt"

	"system/internal/domain"

	"github.com/gofrs/uuid/v5"
	"go.uber.org/zap"
)

// CreatorService handles creator-related business logic
type CreatorService struct {
	creatorRepo       domain.CreatorRepository
	viewAnalyticsRepo domain.ViewAnalyticsRepository
	novelRepo         domain.NovelRepository
	logger            *zap.Logger
}

// NewCreatorService creates a new CreatorService instance
func NewCreatorService(
	creatorRepo domain.CreatorRepository,
	viewAnalyticsRepo domain.ViewAnalyticsRepository,
	novelRepo domain.NovelRepository,
	logger *zap.Logger,
) *CreatorService {
	return &CreatorService{
		creatorRepo:       creatorRepo,
		viewAnalyticsRepo: viewAnalyticsRepo,
		novelRepo:         novelRepo,
		logger:            logger,
	}
}

// ListCreators returns paginated list of creators with view stats and popular work info
func (s *CreatorService) ListCreators(ctx context.Context, filter domain.CreatorListFilter) (*domain.CreatorListResult, error) {
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

	// 3. Get view stats from ClickHouse
	viewPeriod := "all"
	if filter.ViewPeriod != nil {
		viewPeriod = *filter.ViewPeriod
	}

	viewStats, err := s.viewAnalyticsRepo.GetCreatorViewStats(ctx, userIDs, viewPeriod)
	if err != nil {
		// Log error but don't fail - views are optional
		s.logger.Warn("failed to get creator view stats", zap.Error(err))
		viewStats = make(map[uuid.UUID]domain.CreatorViewStats)
	}

	// 4. Collect popular work IDs for fetching metadata
	popularWorkIDs := make([]uuid.UUID, 0)
	workTypeMap := make(map[uuid.UUID]string)
	for _, stats := range viewStats {
		if stats.PopularWorkID != nil {
			popularWorkIDs = append(popularWorkIDs, *stats.PopularWorkID)
			workTypeMap[*stats.PopularWorkID] = stats.PopularWorkType
		}
	}

	// 5. Fetch novel metadata for popular works (currently only novels supported)
	novelInfoMap := make(map[uuid.UUID]struct {
		Title    string
		CoverURL string
	})
	if len(popularWorkIDs) > 0 && s.novelRepo != nil {
		novels, _, err := s.novelRepo.List(ctx, domain.NovelFilter{
			IDs:   popularWorkIDs,
			Limit: len(popularWorkIDs),
		})
		if err != nil {
			s.logger.Warn("failed to get popular work novels", zap.Error(err))
		} else {
			for _, novel := range novels {
				novelInfoMap[novel.ID] = struct {
					Title    string
					CoverURL string
				}{
					Title:    novel.Title,
					CoverURL: getValue(novel.CoverImageURL),
				}
			}
		}
	}

	// 6. Merge view stats and popular work info into creators
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

// getValue is a helper to dereference string pointers safely
func getValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
