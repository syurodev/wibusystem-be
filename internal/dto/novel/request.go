package novel

import (
	"encoding/json"
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

// ListNovelsRequest là DTO cho việc lấy danh sách novels
type ListNovelsRequest struct {
	Page             int      `form:"page" binding:"omitempty,min=1"`
	Limit            int      `form:"limit" binding:"omitempty,min=1,max=100"`
	Owner            string   `form:"owner" binding:"omitempty,uuid"`
	KeySearch        string   `form:"key_search" binding:"omitempty,max=200"`
	GenreIDs         []string `form:"genre_ids" binding:"omitempty,dive,uuid"`
	Statuses         []string `form:"status" binding:"omitempty,dive,oneof=draft ongoing completed hiatus dropped"`
	OriginalLanguage string   `form:"original_language" binding:"omitempty,len=2"`
	SortBy           string   `form:"sort_by" binding:"omitempty,oneof=created_at rating views last_chapter"`
	SortOrder        string   `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}
