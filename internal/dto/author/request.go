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

// MergeAuthorRequest là DTO cho việc gộp authors
type MergeAuthorRequest struct {
	TargetID  uuid.UUID   `json:"target_id" binding:"required"`
	SourceIDs []uuid.UUID `json:"source_ids" binding:"required,min=1"`
}

// ListAuthorsRequest là DTO cho việc lấy danh sách authors
type ListAuthorsRequest struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	Limit      int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Search     string `form:"search" binding:"omitempty,max=100"`
	SortBy     string `form:"sort_by" binding:"omitempty,oneof=name views novels created"`
	SortOrder  string `form:"sort_order" binding:"omitempty,oneof=asc desc"`
	IsVerified *bool  `form:"is_verified"`
}
