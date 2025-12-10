package media

import (
	"system/internal/domain"
)

// MediaSeriesResponse matches the frontend MediaSeriesSchema
type MediaSeriesResponse = domain.MediaSeriesResponse

type GenreInfo = domain.GenreInfo

type OwnerInfo = domain.OwnerInfo

type LatestChapterInfo = domain.LatestChapterInfo

// HomeData aggregates data for the home page.
type HomeData struct {
	Hero     []map[string]any `json:"hero"`
	Trending []map[string]any `json:"trending"`
	Creators []any            `json:"creators"`
	Genres   []any            `json:"genres"`
}
