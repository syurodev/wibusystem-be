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
	Name        string  `json:"name" binding:"required,min=1,max=100"`
	Description string  `json:"description" binding:"max=1000"`
	ParentID    *string `json:"parent_id,omitempty"` // UUID as string
	IsActive    bool    `json:"is_active"`
}

// MergeGenreRequest là DTO cho việc gộp genres
type MergeGenreRequest struct {
	TargetID  uuid.UUID   `json:"target_id" binding:"required"`
	SourceIDs []uuid.UUID `json:"source_ids" binding:"required,min=1"`
}

// ListGenresRequest là DTO cho việc lấy danh sách genres
type ListGenresRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	Limit      int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Search     string `form:"search" binding:"omitempty,max=100"`                                          // Tìm kiếm theo tên
	SortBy     string `form:"sort_by" binding:"omitempty,oneof=name views series created updated readers"` // name, views, series, created, updated, readers
	SortOrder  string `form:"sort_order" binding:"omitempty,oneof=asc desc"`                               // asc hoặc desc
	ActiveOnly bool   `form:"active_only"`
}
