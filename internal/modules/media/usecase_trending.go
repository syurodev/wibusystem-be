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
