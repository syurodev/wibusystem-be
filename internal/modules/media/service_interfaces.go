package media

import (
	"context"

	mediadto "system/internal/dto/media"
)

// MediaService provides unified access to media content (Novel, Manga, Anime).
type MediaService interface {
	GetHomeData(ctx context.Context) (*mediadto.HomeData, error)
	GetTopMediaWithRankComparison(ctx context.Context, mediaType string, period string, limit int) ([]mediadto.MediaSeriesResponse, error)
	InvalidateHomeCache(ctx context.Context) error
}
