package novel

import (
	"encoding/json"
	"system/internal/app/dto"
)

// CreateNovelRequest là DTO cho việc tạo novel mới
type CreateNovelRequest struct {
	Title            string          `json:"title" binding:"required,min=1,max=500"`
	Synopsis         json.RawMessage `json:"synopsis"`
	CoverImageURL    *string         `json:"cover_image_url,omitempty" binding:"omitempty"`
	ThumbnailURL     *string         `json:"thumbnail_url,omitempty" binding:"omitempty"`
	Status           string          `json:"status" binding:"required,oneof=draft ongoing completed hiatus dropped"`
	IsOneshot        bool            `json:"is_oneshot"`
	OriginalLanguage *string         `json:"original_language,omitempty" binding:"omitempty,len=2"` // ISO 639-1
	OriginalTitle    *string         `json:"original_title,omitempty" binding:"omitempty,max=500"`
	OwnerID          string          `json:"owner_id" binding:"required,uuid"`
	OwnerType        string          `json:"owner_type" binding:"required,oneof=user group"`
	GenreIDs         []string        `json:"genre_ids" binding:"omitempty,dive,uuid"`
	AuthorIDs        []string        `json:"author_ids" binding:"omitempty,dive,uuid"`
	ArtistIDs        []string        `json:"artist_ids" binding:"omitempty,dive,uuid"`
	Metadata         *string         `json:"metadata,omitempty"` // JSON string
}

// UpdateNovelRequest là DTO cho việc cập nhật novel
type UpdateNovelRequest struct {
	Title            string          `json:"title" binding:"required,min=1,max=500"`
	Synopsis         json.RawMessage `json:"synopsis"`
	CoverImageURL    *string         `json:"cover_image_url,omitempty" binding:"omitempty"`
	ThumbnailURL     *string         `json:"thumbnail_url,omitempty" binding:"omitempty"`
	Status           string          `json:"status" binding:"required,oneof=draft ongoing completed hiatus dropped"`
	IsOneshot        bool            `json:"is_oneshot"`
	OriginalLanguage *string         `json:"original_language,omitempty" binding:"omitempty,len=2"` // ISO 639-1
	OriginalTitle    *string         `json:"original_title,omitempty" binding:"omitempty,max=500"`
	Metadata         *string         `json:"metadata,omitempty"` // JSON string
}

// OwnerInfo là DTO cho owner (user hoặc tenant) - match BaseUserSchema
type OwnerInfo = dto.OwnerInfo

// GenreInfo là DTO cho genre - match GenreSchema
type GenreInfo = dto.GenreInfo

// LatestChapterInfo là DTO cho chapter mới nhất - match MediaUnitSchema
type LatestChapterInfo = dto.LatestChapterInfo

// NovelResponse là DTO cho response novel (cho danh sách) - match MediaSeriesSchema
type NovelResponse = dto.MediaSeriesResponse

// ListNovelsRequest là DTO cho việc lấy danh sách novels
type ListNovelsRequest struct {
	Page             int      `form:"page" binding:"omitempty,min=1"`
	Limit            int      `form:"limit" binding:"omitempty,min=1,max=100"`
	Owner            string   `form:"owner" binding:"omitempty,uuid"`                                         // Filter theo owner ID (user hoặc tenant)
	KeySearch        string   `form:"key_search" binding:"omitempty,max=200"`                                 // Tìm kiếm trong title và original_title
	GenreIDs         []string `form:"genre_ids" binding:"omitempty,dive,uuid"`                                // Filter theo genre IDs
	Statuses         []string `form:"status" binding:"omitempty,dive,oneof=draft ongoing completed hiatus dropped"` // Filter theo nhiều status
	OriginalLanguage string   `form:"original_language" binding:"omitempty,len=2"`                            // Filter theo ngôn ngữ gốc
	SortBy           string   `form:"sort_by" binding:"omitempty,oneof=created_at rating views last_chapter"` // created_at, rating, views, last_chapter
	SortOrder        string   `form:"sort_order" binding:"omitempty,oneof=asc desc"`                          // asc hoặc desc
}

// NovelDetailResponse là DTO chi tiết cho novel
type NovelDetailResponse struct {
	ID               string  `json:"id"`
	Title            string  `json:"title"`
	Slug             string  `json:"slug"`
	Synopsis         json.RawMessage  `json:"synopsis"` // JSONB
	CoverImageURL    *string `json:"cover_image_url,omitempty"`
	ThumbnailURL     *string  `json:"thumbnail_url,omitempty"`
	Status           string   `json:"status"`
	IsOneshot        bool     `json:"is_oneshot"`
	GenreIDs         []string `json:"genre_ids"`
	AuthorIDs        []string `json:"author_ids"`
	ArtistIDs        []string `json:"artist_ids"`
	Genres           []GenreInfo `json:"genres"`
	Authors          []OwnerInfo `json:"authors"` // Reusing OwnerInfo for simple ID/Name structure or define new
	Artists          []OwnerInfo `json:"artists"`
	OriginalLanguage *string `json:"original_language,omitempty"`
	OriginalTitle    *string `json:"original_title,omitempty"`
	TotalVolumes     int     `json:"total_volumes"`
	TotalChapters    int     `json:"total_chapters"`
	TotalWords       int64   `json:"total_words"`
	ViewCount        int64   `json:"view_count"`
	FavoriteCount    int     `json:"favorite_count"`
	RatingAverage    float64 `json:"rating_average"`
	RatingCount      int     `json:"rating_count"`
	Metadata         *string `json:"metadata,omitempty"` // JSON string
	FirstPublishedAt *string `json:"first_published_at,omitempty"`
	LastChapterAt    *string `json:"last_chapter_at,omitempty"`
	CompletedAt      *string `json:"completed_at,omitempty"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
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
