package novel

import (
	"encoding/json"
	"system/internal/domain"
)

// OwnerInfo là DTO cho owner (user hoặc tenant)
type OwnerInfo = domain.OwnerInfo

// GenreInfo là DTO cho genre
type GenreInfo = domain.GenreInfo

// LatestChapterInfo là DTO cho chapter mới nhất
type LatestChapterInfo = domain.LatestChapterInfo

// NovelResponse là DTO cho response novel (cho danh sách)
type NovelResponse = domain.MediaSeriesResponse

// NovelDetailResponse là DTO chi tiết cho novel
type NovelDetailResponse struct {
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	Slug             string          `json:"slug"`
	Synopsis         json.RawMessage `json:"synopsis"`
	CoverImageURL    *string         `json:"cover_image_url,omitempty"`
	ThumbnailURL     *string         `json:"thumbnail_url,omitempty"`
	Status           string          `json:"status"`
	IsOneshot        bool            `json:"is_oneshot"`
	GenreIDs         []string        `json:"genre_ids"`
	AuthorIDs        []string        `json:"author_ids"`
	ArtistIDs        []string        `json:"artist_ids"`
	Genres           []GenreInfo     `json:"genres"`
	Authors          []OwnerInfo     `json:"authors"`
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
	Metadata         *string         `json:"metadata,omitempty"`
	FirstPublishedAt *string         `json:"first_published_at,omitempty"`
	LastChapterAt    *string         `json:"last_chapter_at,omitempty"`
	CompletedAt      *string         `json:"completed_at,omitempty"`
	CreatedAt        string          `json:"created_at"`
	UpdatedAt        string          `json:"updated_at"`
}

// ChapterSummaryResponse - chapter info không có content
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

// VolumeInfoResponse - volume info cơ bản với chapters
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

// NovelFullResponse - response đầy đủ cho novel detail page
type NovelFullResponse struct {
	NovelDetailResponse
	Owner    OwnerInfo                `json:"owner"`
	Volumes  []VolumeInfoResponse     `json:"volumes"`
	Chapters []ChapterSummaryResponse `json:"chapters"` // chapters không thuộc volume nào
}
