package author

import "github.com/gofrs/uuid/v5"

// CreateAuthorRequest là DTO cho việc tạo author mới
type CreateAuthorRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=200"`
	Biography   string  `json:"biography" binding:"max=5000"`
	AvatarURL   *string `json:"avatar_url,omitempty" binding:"omitempty,url"`
	SocialLinks *string `json:"social_links,omitempty"` // JSON string
}

// UpdateAuthorRequest là DTO cho việc cập nhật author
type UpdateAuthorRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=200"`
	Biography   string  `json:"biography" binding:"max=5000"`
	AvatarURL   *string `json:"avatar_url,omitempty" binding:"omitempty,url"`
	SocialLinks *string `json:"social_links,omitempty"` // JSON string
}

// AuthorResponse là DTO cho response author (cho danh sách)
type AuthorResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"` // Biography
	NovelCount  int     `json:"novel_count"`
	TotalViews  int64   `json:"total_views"`
	CreatedAt   string  `json:"created_at"` // Ngày tạo
}

// ListAuthorsRequest là DTO cho việc lấy danh sách authors
type ListAuthorsRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	Limit      int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Search     string `form:"search" binding:"omitempty,max=100"`                          // Tìm kiếm theo tên
	SortBy     string `form:"sort_by" binding:"omitempty,oneof=name views novels created"` // name, views, novels, created
	SortOrder  string `form:"sort_order" binding:"omitempty,oneof=asc desc"`               // asc hoặc desc
	IsVerified *bool  `form:"is_verified"`                                                 // Filter theo verified status
}

// AuthorDetailResponse là DTO chi tiết cho author
type AuthorDetailResponse struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Slug          string  `json:"slug"`
	Description   *string `json:"description,omitempty"` // Biography
	AvatarURL     *string `json:"avatar_url,omitempty"`
	SocialLinks   *string `json:"social_links,omitempty"` // JSON string
	NovelCount    int     `json:"novel_count"`
	TotalChapters int     `json:"total_chapters"`
	TotalViews    int64   `json:"total_views"`
	FollowerCount int     `json:"follower_count"`
	IsVerified    bool    `json:"is_verified"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// MergeAuthorRequest là DTO cho việc gộp authors
type MergeAuthorRequest struct {
	TargetID  uuid.UUID   `json:"target_id" binding:"required"`
	SourceIDs []uuid.UUID `json:"source_ids" binding:"required,min=1"`
}

type PreviewMergeAuthorResponse struct {
	AffectedNovels []AffectedNovel `json:"affected_novels"`
	SourceAuthors   []string       `json:"source_authors"`
}

type AffectedNovel struct {
	ID            uuid.UUID `json:"id"`
	Title         string    `json:"title"`
	Slug          string    `json:"slug"`
	CoverImageURL *string   `json:"cover_image_url"`
}
