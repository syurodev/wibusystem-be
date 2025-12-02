package volume_chapter

import (
	"encoding/json"
)

// CreateChapterRequest represents the request to create a new chapter
type CreateChapterRequest struct {
	ChapterNumber int             `json:"chapter_number" binding:"required,min=1"`
	Title         string          `json:"title" binding:"required,min=1,max=500"`
	Content       json.RawMessage `json:"content" binding:"required"`
	AuthorNotes   json.RawMessage `json:"author_notes,omitempty"`
	IsFree        bool            `json:"is_free"`
	Price         *float64        `json:"price,omitempty" binding:"omitempty,min=0"`
	Currency      *string         `json:"currency,omitempty" binding:"omitempty,len=3"` // ISO 4217
	Status        string          `json:"status" binding:"required,oneof=draft published scheduled"`
	DisplayOrder  int             `json:"display_order" binding:"min=0"`
	ScheduledAt   *string         `json:"scheduled_at,omitempty"` // ISO 8601 format
}

// UpdateChapterRequest represents the request to update a chapter
type UpdateChapterRequest struct {
	VolumeID      *string         `json:"volume_id,omitempty" binding:"omitempty,uuid"`
	ChapterNumber int             `json:"chapter_number" binding:"required,min=1"`
	Title         string          `json:"title" binding:"required,min=1,max=500"`
	Content       json.RawMessage `json:"content" binding:"required"`
	AuthorNotes   json.RawMessage `json:"author_notes,omitempty"`
	IsFree        bool            `json:"is_free"`
	Price         *float64        `json:"price,omitempty" binding:"omitempty,min=0"`
	Currency      *string         `json:"currency,omitempty" binding:"omitempty,len=3"`
	Status        string          `json:"status" binding:"required,oneof=draft published scheduled"`
	DisplayOrder  int             `json:"display_order" binding:"min=0"`
	ScheduledAt   *string         `json:"scheduled_at,omitempty"`
}

// ScheduleChapterRequest represents the request to schedule a chapter
type ScheduleChapterRequest struct {
	ScheduledAt string `json:"scheduled_at" binding:"required"` // ISO 8601 format
}

// ChapterResponse represents the basic chapter information in a list
type ChapterResponse struct {
	ID            string   `json:"id"`
	NovelID       string   `json:"novel_id"`
	VolumeID      *string  `json:"volume_id,omitempty"`
	ChapterNumber int      `json:"chapter_number"`
	Title         string   `json:"title"`
	Slug          string   `json:"slug"`
	WordCount     int      `json:"word_count"`
	IsFree        bool     `json:"is_free"`
	Price         *float64 `json:"price,omitempty"`
	Currency      *string  `json:"currency,omitempty"`
	Status        string   `json:"status"`
	ViewCount     int64    `json:"view_count"`
	LikeCount     int      `json:"like_count"`
	CommentCount  int      `json:"comment_count"`
	DisplayOrder  int      `json:"display_order"`
	PublishedAt   *string  `json:"published_at,omitempty"`
	ScheduledAt   *string  `json:"scheduled_at,omitempty"`
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

// ChapterDetailResponse represents the detailed chapter information
type ChapterDetailResponse struct {
	ID             string          `json:"id"`
	NovelID        string          `json:"novel_id"`
	VolumeID       *string         `json:"volume_id,omitempty"`
	ChapterNumber  int             `json:"chapter_number"`
	Title          string          `json:"title"`
	Slug           string          `json:"slug"`
	Content        json.RawMessage `json:"content"`
	WordCount      int             `json:"word_count"`
	CharacterCount int             `json:"character_count"`
	IsFree         bool            `json:"is_free"`
	Price          *float64        `json:"price,omitempty"`
	Currency       *string         `json:"currency,omitempty"`
	Status         string          `json:"status"`
	ViewCount      int64           `json:"view_count"`
	LikeCount      int             `json:"like_count"`
	CommentCount   int             `json:"comment_count"`
	DisplayOrder   int             `json:"display_order"`
	AuthorNotes    json.RawMessage `json:"author_notes,omitempty"`
	PublishedAt    *string         `json:"published_at,omitempty"`
	ScheduledAt    *string         `json:"scheduled_at,omitempty"`
	CreatedAt      string          `json:"created_at"`
	UpdatedAt      string          `json:"updated_at"`
}

// ListChaptersRequest represents query parameters for listing chapters
type ListChaptersRequest struct {
	Status        string  `form:"status" binding:"omitempty,oneof=draft published scheduled"`
	VolumeID      *string `form:"volume_id" binding:"omitempty,uuid"`
	PublishedOnly bool    `form:"published_only"`
	IsFree        *bool   `form:"is_free"`
	SortBy        string  `form:"sort_by" binding:"omitempty,oneof=chapter_number published_at views"`
	SortOrder     string  `form:"sort_order" binding:"omitempty,oneof=asc desc"`
	Page          int     `form:"page" binding:"omitempty,min=1"`
	Limit         int     `form:"limit" binding:"omitempty,min=1,max=100"`
}

// UpdateStatisticsRequest represents the request to update chapter statistics
type UpdateStatisticsRequest struct {
	ViewCount    *int64 `json:"view_count,omitempty"`
	LikeCount    *int   `json:"like_count,omitempty"`
	CommentCount *int   `json:"comment_count,omitempty"`
}

// ListChaptersResponse represents the response for listing chapters of a volume
type ListChaptersResponse struct {
	VolumeID    string            `json:"volume_id"`
	VolumeTitle string            `json:"volume_title"`
	Chapters    []ChapterResponse `json:"chapters"`
}
