package router

import (
	"context"

	analytics_module "system/internal/modules/analytics"
	novel_module "system/internal/modules/novel"
)

// TopNovelAdapter adapts analytics service to novel.TopNovelService interface
// This adapter helps avoid import cycle between novel and analytics modules
type TopNovelAdapter struct {
	analyticsSvc analytics_module.AnalyticsService
}

// NewTopNovelAdapter creates a new TopNovelAdapter
func NewTopNovelAdapter(analyticsSvc analytics_module.AnalyticsService) *TopNovelAdapter {
	return &TopNovelAdapter{
		analyticsSvc: analyticsSvc,
	}
}

// GetTopNovelsWithRank implements novel.TopNovelService interface
func (a *TopNovelAdapter) GetTopNovelsWithRank(ctx context.Context, period string, offset int, limit int) ([]novel_module.MediaRankResult, error) {
	ranks, err := a.analyticsSvc.GetTopMediaWithRankComparison(ctx, period, "novel", offset, limit)
	if err != nil {
		return nil, err
	}

	results := make([]novel_module.MediaRankResult, 0, len(ranks))
	for _, rank := range ranks {
		results = append(results, novel_module.MediaRankResult{
			Novel: rank.Novel,
			Stats: novel_module.MediaRankStat{
				EntityID:     rank.ID,
				TotalViews:   int(rank.Stats.TotalViews),
				CurrentRank:  rank.Stats.CurrentRank,
				PreviousRank: rank.Stats.PreviousRank,
				RankChange:   rank.Stats.RankChange,
			},
		})
	}

	return results, nil
}
