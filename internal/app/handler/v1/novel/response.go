package novel

import (
	"encoding/json"
	"system/internal/app/dto"
)

// OwnerInfo là DTO cho owner (user hoặc tenant) - match BaseUserSchema
type OwnerInfo = dto.OwnerInfo

// GenreInfo là DTO cho genre - match GenreSchema
type GenreInfo = dto.GenreInfo

// LatestChapterInfo là DTO cho chapter mới nhất - match MediaUnitSchema
type LatestChapterInfo = dto.LatestChapterInfo

// NovelResponse là DTO cho response novel (cho danh sách) - match MediaSeriesSchema
type NovelResponse = dto.MediaSeriesResponse

// NovelDetailResponse là DTO chi tiết cho novel
type NovelDetailResponse struct {
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	Slug             string          `json:"slug"`
	Synopsis         json.RawMessage `json:"synopsis"` // JSONB
	CoverImageURL    *string         `json:"cover_image_url,omitempty"`
	ThumbnailURL     *string         `json:"thumbnail_url,omitempty"`
	Status           string          `json:"status"`
	IsOneshot        bool            `json:"is_oneshot"`
	GenreIDs         []string        `json:"genre_ids"`
	AuthorIDs        []string        `json:"author_ids"`
	ArtistIDs        []string        `json:"artist_ids"`
	Genres           []GenreInfo     `json:"genres"`
	Authors          []OwnerInfo     `json:"authors"` // Reusing OwnerInfo for simple ID/Name structure or define new
	Artists          []OwnerInfo     `json:"artists"`
	OriginalLanguage *string         `json:"original_language,omitempty"`
	OriginalTitle    *string         `json:"original_title,omitempty"`
	TotalVolumes     int             `json:"total_volumes"`
	TotalChapters    int             `json:"total_chapters"`
	TotalWords       int64           `json:"total_words"`
	ViewCount        int64           `json:"view_count"`
	FavoriteCount    int             `json:"favorite_count"`
	RatingAverage    float64         `json:"rating_average"`
	RatingCount      int             `json:"rating_count"`
	Metadata         *string         `json:"metadata,omitempty"` // JSON string
	FirstPublishedAt *string         `json:"first_published_at,omitempty"`
	LastChapterAt    *string         `json:"last_chapter_at,omitempty"`
	CompletedAt      *string         `json:"completed_at,omitempty"`
	CreatedAt        string          `json:"created_at"`
	UpdatedAt        string          `json:"updated_at"`
}

// ChapterSummaryResponse - chapter info không có content (dùng cho API full)
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

// VolumeInfoResponse - volume info cơ bản với chapters (dùng cho API full)
type VolumeInfoResponse struct {
	ID            string                   `json:"id"`
	VolumeNumber  int                      `json:"volume_number"`
	Title         string                   `json:"title"`
	Slug          string                   `json:"slug"`
	CoverImageURL *string                  `json:"cover_image_url,omitempty"`
	DisplayOrder  int                      `json:"display_order"`
	IsPublished   bool                     `json:"is_published"`
	PublishedAt   *string                  `json:"published_at,omitempty"`
	Chapters      []ChapterSummaryResponse `json:"chapters"`
}

// NovelFullResponse - response đầy đủ cho novel detail page (public API)
type NovelFullResponse struct {
	NovelDetailResponse
	Owner    OwnerInfo                `json:"owner"`
	Volumes  []VolumeInfoResponse     `json:"volumes"`
	Chapters []ChapterSummaryResponse `json:"chapters"` // chapters không thuộc volume nào
}
