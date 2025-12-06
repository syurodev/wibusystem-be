package dto

import "encoding/json"

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
	Synopsis    json.RawMessage `json:"synopsis"` // TNode[] as raw JSON
	CoverURL    *string         `json:"cover_url,omitempty"`
	
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
}

type GenreInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type OwnerInfo struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"display_name"`
	Username    string  `json:"username"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
}

type LatestChapterInfo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	PublishedAt string `json:"published_at"`
}
