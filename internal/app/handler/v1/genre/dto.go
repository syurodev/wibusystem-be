package genre

import "github.com/gofrs/uuid/v5"

// CreateGenreRequest là DTO cho việc tạo genre mới
type CreateGenreRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=100"`
	Description string  `json:"description" binding:"max=1000"`
	ParentID    *string `json:"parent_id,omitempty"` // UUID as string
}

// UpdateGenreRequest là DTO cho việc cập nhật genre
type UpdateGenreRequest struct {
	Name         string  `json:"name" binding:"required,min=1,max=100"`
	Description  string  `json:"description" binding:"max=1000"`
	ParentID     *string `json:"parent_id,omitempty"` // UUID as string
	DisplayOrder int     `json:"display_order" binding:"min=0"`
	IsActive     bool    `json:"is_active"`
}

// MergeGenreRequest là DTO cho việc gộp genres
type MergeGenreRequest struct {
	TargetID  uuid.UUID   `json:"target_id" binding:"required"`
	SourceIDs []uuid.UUID `json:"source_ids" binding:"required,min=1"`
}

type PreviewMergeGenreResponse struct {
	AffectedNovels []AffectedNovel `json:"affected_novels"`
	SourceGenres   []string        `json:"source_genres"`
}

type AffectedNovel struct {
	ID            uuid.UUID `json:"id"`
	Title         string    `json:"title"`
	Slug          string    `json:"slug"`
	CoverImageURL *string   `json:"cover_image_url"`
}

// GenreResponse là DTO cho response genre
type GenreResponse struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Slug          string  `json:"slug"`
	IsActive      bool    `json:"is_active"`
	SeriesCount   int     `json:"series_count"`   // tổng số novel (sau này sẽ có cả manga và anime)
	ActiveReaders int64   `json:"active_readers"` // số độc giả đang active
	TotalViews    int64   `json:"total_views"`    // tổng lượt xem
	Trend         string  `json:"trend"`          // "rising" | "stable" | "falling"
	Description   *string `json:"description,omitempty"`
	CreatedAt     string  `json:"created_at"` // Ngày tạo
	UpdatedAt     string  `json:"updated_at"` // Ngày cập nhật
}

// ListGenresRequest là DTO cho việc lấy danh sách genres
type ListGenresRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	Limit      int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Search     string `form:"search" binding:"omitempty,max=100"`                         // Tìm kiếm theo tên
	SortBy     string `form:"sort_by" binding:"omitempty,oneof=name views series created updated"` // name, views, series, created, updated
	SortOrder  string `form:"sort_order" binding:"omitempty,oneof=asc desc"`             // asc hoặc desc
	ActiveOnly bool   `form:"active_only"`
}

// GenreDetailResponse là DTO chi tiết cho genre (bao gồm thêm thông tin)
type GenreDetailResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Slug         string  `json:"slug"`
	Description  *string `json:"description,omitempty"`
	ParentID     *string `json:"parent_id,omitempty"`
	DisplayOrder int     `json:"display_order"`
	IsActive     bool    `json:"is_active"`

	// Statistics
	SeriesCount   int   `json:"series_count"`
	ActiveReaders int64 `json:"active_readers"`
	TotalViews    int64 `json:"total_views"`
	Trend         string `json:"trend"`

	// Audit
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
