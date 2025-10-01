package dto

import (
	"encoding/json"
	"time"
)

// CreateChapterRequest represents the payload for creating a new chapter
// This follows the API design spec from /services/catalog/api-design/novel.md section 3.2
type CreateChapterRequest struct {
	ChapterNumber       int              `json:"chapter_number" validate:"required,min=1"` // Chapter number in volume (required)
	Title               *string          `json:"title,omitempty" validate:"omitempty,max=500"` // Chapter title (optional)
	Content             *json.RawMessage `json:"content,omitempty"` // Chapter content in JSONB rich text format (optional)
	IsPublic            bool             `json:"is_public"` // Public visibility flag
	IsDraft             bool             `json:"is_draft"` // Draft status flag
	PriceCoins          *int             `json:"price_coins,omitempty" validate:"omitempty,min=0"` // Purchase price in coins (optional)
	ContentWarnings     *json.RawMessage `json:"content_warnings,omitempty"` // Content warnings as JSONB array (optional)
	HasMatureContent    bool             `json:"has_mature_content"` // Mature content flag
	ScheduledPublishAt  *time.Time       `json:"scheduled_publish_at,omitempty"` // Scheduled publication time (optional)
}

// UpdateChapterRequest represents the payload for updating an existing chapter
// This follows the API design spec from /services/catalog/api-design/novel.md section 3.4
type UpdateChapterRequest struct {
	Title            *string          `json:"title,omitempty" validate:"omitempty,max=500"` // Chapter title (optional)
	Content          *json.RawMessage `json:"content,omitempty"` // Chapter content in JSONB rich text format (optional)
	IsPublic         *bool            `json:"is_public,omitempty"` // Public visibility flag (optional)
	IsDraft          *bool            `json:"is_draft,omitempty"` // Draft status flag (optional)
	PriceCoins       *int             `json:"price_coins,omitempty" validate:"omitempty,min=0"` // Purchase price in coins (optional)
	ContentWarnings  *json.RawMessage `json:"content_warnings,omitempty"` // Content warnings as JSONB array (optional)
	HasMatureContent *bool            `json:"has_mature_content,omitempty"` // Mature content flag (optional)
}

// ListChaptersRequest represents query parameters for listing chapters
// This follows the API design spec from /services/catalog/api-design/novel.md section 3.1
type ListChaptersRequest struct {
	Page           int  `form:"page" validate:"omitempty,min=1"` // Current page number (default: 1)
	Limit          int  `form:"limit" validate:"omitempty,min=1,max=100"` // Items per page (default: 50, max: 100)
	IncludeContent bool `form:"include_content"` // Include chapter content in response (default: false)
}

// PublishChapterRequest represents the payload for publishing a chapter
// This follows the API design spec from /services/catalog/api-design/novel.md section 3.6
type PublishChapterRequest struct {
	PublishAt *time.Time `json:"publish_at,omitempty"` // Publication time (optional, defaults to now)
}

// ChapterResponse represents a chapter in list/detail responses
// This follows the API design spec from /services/catalog/api-design/novel.md section 3.1
type ChapterResponse struct {
	ID                 string           `json:"id"` // Chapter UUID
	VolumeID           string           `json:"volume_id"` // Parent volume UUID
	ChapterNumber      int              `json:"chapter_number"` // Chapter number in volume
	Title              *string          `json:"title,omitempty"` // Chapter title (optional)
	Content            *json.RawMessage `json:"content,omitempty"` // Chapter content (only included if requested)
	PublishedAt        *time.Time       `json:"published_at,omitempty"` // Publication date (optional)
	IsPublic           bool             `json:"is_public"` // Public visibility flag
	IsDraft            bool             `json:"is_draft"` // Draft status flag
	PriceCoins         *int             `json:"price_coins,omitempty"` // Purchase price in coins (optional)
	WordCount          *int             `json:"word_count,omitempty"` // Word count (optional)
	CharacterCount     *int             `json:"character_count,omitempty"` // Character count (optional)
	ReadingTimeMinutes *int             `json:"reading_time_minutes,omitempty"` // Estimated reading time in minutes (optional)
	ViewCount          int64            `json:"view_count"` // View count
	LikeCount          int64            `json:"like_count"` // Like count
	CommentCount       int64            `json:"comment_count"` // Comment count
	ContentWarnings    []string         `json:"content_warnings,omitempty"` // Content warnings array
	HasMatureContent   bool             `json:"has_mature_content"` // Mature content flag
	Version            int              `json:"version"` // Version number
	CreatedAt          time.Time        `json:"created_at"` // Creation timestamp
	UpdatedAt          time.Time        `json:"updated_at"` // Last update timestamp
}

// CreateChapterResponse represents the response after creating a chapter
// This follows the API design spec from /services/catalog/api-design/novel.md section 3.2
type CreateChapterResponse struct {
	ID                 string    `json:"id"` // Chapter UUID
	VolumeID           string    `json:"volume_id"` // Parent volume UUID
	ChapterNumber      int       `json:"chapter_number"` // Chapter number in volume
	Title              *string   `json:"title,omitempty"` // Chapter title (optional)
	IsDraft            bool      `json:"is_draft"` // Draft status flag
	WordCount          *int      `json:"word_count,omitempty"` // Word count (optional)
	CharacterCount     *int      `json:"character_count,omitempty"` // Character count (optional)
	ReadingTimeMinutes *int      `json:"reading_time_minutes,omitempty"` // Estimated reading time (optional)
	CreatedAt          time.Time `json:"created_at"` // Creation timestamp
}

// UpdateChapterResponse represents the response after updating a chapter
// This follows the API design spec from /services/catalog/api-design/novel.md section 3.4
type UpdateChapterResponse struct {
	ID                 string    `json:"id"` // Chapter UUID
	Title              *string   `json:"title,omitempty"` // Chapter title (optional)
	WordCount          *int      `json:"word_count,omitempty"` // Word count (optional)
	CharacterCount     *int      `json:"character_count,omitempty"` // Character count (optional)
	ReadingTimeMinutes *int      `json:"reading_time_minutes,omitempty"` // Estimated reading time (optional)
	Version            int       `json:"version"` // Version number
	UpdatedAt          time.Time `json:"updated_at"` // Last update timestamp
}

// PublishChapterResponse represents the response after publishing/unpublishing a chapter
// This follows the API design spec from /services/catalog/api-design/novel.md sections 3.6 and 3.7
type PublishChapterResponse struct {
	ID          string     `json:"id"` // Chapter UUID
	IsPublic    bool       `json:"is_public"` // Public visibility flag
	IsDraft     bool       `json:"is_draft"` // Draft status flag
	PublishedAt *time.Time `json:"published_at,omitempty"` // Publication date (null if unpublished)
}

// PaginatedChaptersResponse represents a paginated list of chapters
// This follows the API design spec from /services/catalog/api-design/novel.md section 3.1
type PaginatedChaptersResponse struct {
	Chapters   []ChapterResponse `json:"chapters"` // List of chapters
	Pagination PaginationMeta    `json:"pagination"` // Pagination metadata
}