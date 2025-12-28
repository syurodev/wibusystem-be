package media

import (
	"encoding/json"

	"system/internal/dto/shared"
)

// Re-export shared types for backward compatibility
type (
	OwnerInfo   = shared.OwnerInfo
	GenreInfo   = shared.GenreInfo
	CreatorInfo = shared.CreatorInfo
)

// MediaSeriesResponse matches the frontend MediaSeriesSchema
// This is a unified response for Novel, Manga, and Anime lists.
type MediaSeriesResponse struct {
	// Required fields
	ID               string  `json:"id"`
	Title            string  `json:"title"`
	OriginalTitle    *string `json:"original_title,omitempty"`
	Slug             string  `json:"slug"`
	OriginalLanguage *string `json:"original_language,omitempty"`

	// Content fields
	Synopsis json.RawMessage `json:"synopsis"` // TNode[] as raw JSON
	CoverURL *string         `json:"cover_url,omitempty"`

	// Type and status
	Type   string `json:"type"`   // "novel", "manga", "anime"
	Status string `json:"status"` // "ongoing", "completed", etc.

	// Relations
	Genres []GenreInfo `json:"genres"`
	Owner  OwnerInfo   `json:"owner"`

	// Stats
	Rating    float64 `json:"rating"`
	Views     int64   `json:"views"`
	Favorites int     `json:"favorites"`

	// Optional
	LatestChapter *LatestChapterInfo `json:"latest_chapter,omitempty"`

	// Timestamps
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`

	// Rank Comparison (Optional)
	CurrentRank  *int `json:"current_rank,omitempty"`
	PreviousRank *int `json:"previous_rank,omitempty"`
	RankChange   *int `json:"rank_change,omitempty"`
}

type LatestChapterInfo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	PublishedAt string `json:"published_at"`
}

// HomeData là DTO cho response trang chủ media
type HomeData struct {
	Hero                []map[string]any `json:"hero"`
	Trending            []map[string]any `json:"trending"`
	Creators            []any            `json:"creators"`
	Genres              []any            `json:"genres"`
	MostActiveCreators  []any            `json:"most_active_creators"`
	RisingStars         []any            `json:"rising_stars"`
	ActiveOrganizations []any            `json:"active_organizations"`
	FreshUpdates        []any            `json:"fresh_updates"`
}
