package media

import (
	"context"
	"encoding/json"

	mediadto "system/internal/dto/media"
	analytics_module "system/internal/modules/analytics"
)

type getTrendingUseCase struct {
	analyticsService analytics_module.AnalyticsService
}

func NewGetTrendingUseCase(analyticsService analytics_module.AnalyticsService) GetTrendingUseCase {
	return &getTrendingUseCase{
		analyticsService: analyticsService,
	}
}

func (uc *getTrendingUseCase) Execute(ctx context.Context, input GetTrendingInput) ([]mediadto.MediaSeriesResponse, error) {
	if input.IncludeRankChange {
		// New path for rank comparison
		// Note: GetTrendingUseCase uses AnalyticsService directly in current code (line 12).
		// But to use `GetTopMediaWithRankComparison` which I implemented in `MediaService`, 
		// this UseCase needs access to `MediaService`.
		// Currently it has `analyticsService`.
		// AND `MediaService` has `analyticsService`.
		// Circular dependency if I inject MediaService into UseCase which is used by Handler?
		// Start: Handler -> UseCase -> AnalyticsService
		// New Plan: Handler -> UseCase -> MediaService -> AnalyticsService
		// OR: Handler -> UseCase -> AnalyticsService (but I put the logic in MediaService?)
		
		// I put `GetTopMediaWithRankComparison` implementation in `MediaService`.
		// So UseCase needs `MediaService`.
		// BUT `MediaService` is usually higher level. UseCase uses Services.
		// `GetTrendingUseCase` struct definition (line 11) only has `AnalyticsService`.
		// I should rely on `AnalyticsService` directly here too?
		// NO, because `AnalyticsService` returns `MediaRankResponse` (DTO from analytics module),
		// but UseCase must return `MediaSeriesResponse` (DTO from media module).
		// Usage of `MediaService` is appropriate here to do the mapping.
		// Or I implement the mapping inside UseCase.
		
		// Let's implement mapping inside UseCase to avoid changing dependencies too much.
		// I need to update `GetTrendingUseCase` struct if I want to use `MediaService`.
		// But `GetTrendingUseCase` is in `media` package.
		// `MediaService` is in `media` package.
		
		// If I use `AnalyticsService` directly:
		results, err := uc.analyticsService.GetTopMediaWithRankComparison(ctx, input.Range, input.MediaType, input.Limit)
		if err != nil {
			return nil, err
		}
		
		// Map `results` (MediaRankResponse) to `mediadto.MediaSeriesResponse`
		responses := make([]mediadto.MediaSeriesResponse, len(results))
		for i, r := range results {
			currentRank := r.Stats.CurrentRank
			var prevRank *int
			if r.Stats.PreviousRank != nil {
				pr := *r.Stats.PreviousRank
				prevRank = &pr
			}
			var rankChange *int
			if r.Stats.RankChange != nil {
				rc := *r.Stats.RankChange
				rankChange = &rc
			}
			
			responses[i] = mediadto.MediaSeriesResponse{
				ID:           r.ID.String(),
				Title:        r.Title,
				Slug:         r.Slug,
				Type:         r.Type,
				Views:        r.Stats.TotalViews,
				CurrentRank:  &currentRank,
				PreviousRank: prevRank,
				RankChange:   rankChange,
			}
			if r.Cover != "" {
				responses[i].CoverURL = &r.Cover
			}
		}
		return responses, nil
	}

	results, err := uc.analyticsService.GetTopTrending(ctx, input.MediaType, input.Range, input.Limit)
	if err != nil {
		return nil, err
	}

	// Convert map to typed response
	var resp []mediadto.MediaSeriesResponse
	bytes, err := json.Marshal(results)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}
