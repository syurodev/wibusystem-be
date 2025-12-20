package media

import (
	"context"

	mediadto "system/internal/dto/media"
)

type GetTrendingInput struct {
	MediaType string
	Range     string
	Limit     int
}

type GetTrendingUseCase interface {
	Execute(ctx context.Context, input GetTrendingInput) ([]mediadto.MediaSeriesResponse, error)
}

type GetHomeDataInput struct{}

type GetHomeDataUseCase interface {
	Execute(ctx context.Context, input GetHomeDataInput) (*mediadto.HomeData, error)
}

