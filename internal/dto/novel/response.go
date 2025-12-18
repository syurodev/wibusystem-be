package novel

import (
	"encoding/json"

	mediares "system/internal/dto/media"
	chapterres "system/internal/dto/novel_chapter"
	volumeres "system/internal/dto/novel_volume"
)

// Aliases for common types from media
type GenreInfo = mediares.GenreInfo
type OwnerInfo = mediares.OwnerInfo
type CreatorInfo = mediares.CreatorInfo
type LatestChapterInfo = mediares.LatestChapterInfo

// NovelResponse là DTO cho response novel (cho danh sách)
type NovelResponse = mediares.MediaSeriesResponse

// AffectedNovel là thông tin novel tối thiểu dùng cho merge preview
type AffectedNovel struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Slug          string  `json:"slug"`
	CoverImageURL *string `json:"cover_image_url,omitempty"`
}

// SimilarNovelResponse là response cho API similar novels
type SimilarNovelResponse struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Slug          string  `json:"slug"`
	CoverImageURL *string `json:"cover_image_url,omitempty"`
	Distance      float32 `json:"distance"` // Cosine distance (0 = identical, 2 = opposite)
}

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
	Authors          []CreatorInfo   `json:"authors"`
	Artists          []CreatorInfo   `json:"artists"`
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

// Aliases for volume/chapter types - import từ packages tương ứng
type ChapterSummaryResponse = chapterres.ChapterSummaryResponse
type VolumeInfoResponse = volumeres.VolumeInfoResponse
type VolumeInfoResponseWithChapters = volumeres.VolumeInfoResponseWithChapters

// NovelFullResponse - response đầy đủ cho novel detail page
type NovelFullResponse struct {
	NovelDetailResponse
	Owner    OwnerInfo                      `json:"owner"`
	Volumes  []VolumeInfoResponseWithChapters `json:"volumes"`
	Chapters []ChapterSummaryResponse       `json:"chapters"` // chapters không thuộc volume nào
}

// Aliases for volume/chapter response types (backward compat)
type VolumeResponse = volumeres.VolumeResponse
type VolumeDetailResponse = volumeres.VolumeDetailResponse
type ChapterResponse = chapterres.ChapterResponse
type ChapterDetailResponse = chapterres.ChapterDetailResponse
