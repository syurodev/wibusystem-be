package media

import (
	"context"
)

// MediaService provides unified access to media content (Novel, Manga, Anime).
type MediaService interface {
	GetHomeData(ctx context.Context) (*HomeData, error)
	InvalidateHomeCache(ctx context.Context) error
}
