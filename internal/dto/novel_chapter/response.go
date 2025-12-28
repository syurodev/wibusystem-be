package novel_chapter

import (
	"encoding/json"
)

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
	SourceLanguage *string         `json:"source_language,omitempty"`
	PublishedAt    *string         `json:"published_at,omitempty"`
	ScheduledAt    *string         `json:"scheduled_at,omitempty"`
	CreatedAt      string          `json:"created_at"`
	UpdatedAt      string          `json:"updated_at"`
}

// ListChaptersResponse represents the response for listing chapters of a volume
type ListChaptersResponse struct {
	VolumeID    string            `json:"volume_id"`
	VolumeTitle string            `json:"volume_title"`
	Chapters    []ChapterResponse `json:"chapters"`
}

// ChapterSummaryResponse - chapter info không có content (dùng trong novel full response)
type ChapterSummaryResponse struct {
	ID            string  `json:"id"`
	VolumeID      *string `json:"volume_id,omitempty"`
	ChapterNumber int     `json:"chapter_number"`
	Title         string  `json:"title"`
	Slug          string  `json:"slug"`
	DisplayOrder  int     `json:"display_order"`
	Status        string  `json:"status"`
	PublishedAt   *string `json:"published_at,omitempty"`
}

type MinimalChapterInfo struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	PublishedAt string `json:"published_at"`
}

// ChapterFullResponse - response cho endpoint /:identifier/full
// Bao gồm thông tin novel, volume, và owner
type ChapterFullResponse struct {
	ID             string          `json:"id"`
	NovelID        string          `json:"novel_id"`
	NovelName      string          `json:"novel_name"`
	VolumeID       *string         `json:"volume_id,omitempty"`
	VolumeName     *string         `json:"volume_name,omitempty"`
	ChapterNumber  int             `json:"chapter_number"`
	Title          string          `json:"title"`
	Slug           string          `json:"slug"`
	Content        json.RawMessage `json:"content"`
	WordCount      int             `json:"word_count"`
	CharacterCount int             `json:"character_count"`
	IsFree         bool            `json:"is_free"`
	Status         string          `json:"status"`
	ViewCount      int64           `json:"view_count"`
	LikeCount      int             `json:"like_count"`
	CommentCount   int             `json:"comment_count"`
	AuthorNotes    json.RawMessage `json:"author_notes,omitempty"`
	PublishedAt    *string         `json:"published_at,omitempty"`
	CreatedBy      *string         `json:"created_by,omitempty"`
	Owner          *OwnerInfo      `json:"owner,omitempty"`
}

// OwnerInfo - alias to shared.OwnerInfo for backward compatibility
type OwnerInfo = struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"display_name"`
	Username    string  `json:"username"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
	Slug        *string `json:"slug,omitempty"`
}
