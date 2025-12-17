package artist

import "github.com/gofrs/uuid/v5"

// CreateArtistRequest là DTO cho việc tạo artist mới
type CreateArtistRequest struct {
	Name           string  `json:"name" binding:"required,min=1,max=200"`
	Biography      string  `json:"biography" binding:"max=5000"`
	AvatarURL      *string `json:"avatar_url,omitempty" binding:"omitempty,url"`
	SocialLinks    *string `json:"social_links,omitempty"` // JSON string
	Specialization *string `json:"specialization,omitempty" binding:"omitempty,max=100"`
	PortfolioURL   *string `json:"portfolio_url,omitempty" binding:"omitempty,url"`
}

// UpdateArtistRequest là DTO cho việc cập nhật artist
type UpdateArtistRequest struct {
	Name           string  `json:"name" binding:"required,min=1,max=200"`
	Biography      string  `json:"biography" binding:"max=5000"`
	AvatarURL      *string `json:"avatar_url,omitempty" binding:"omitempty,url"`
	SocialLinks    *string `json:"social_links,omitempty"` // JSON string
	Specialization *string `json:"specialization,omitempty" binding:"omitempty,max=100"`
	PortfolioURL   *string `json:"portfolio_url,omitempty" binding:"omitempty,url"`
}

// MergeArtistRequest là DTO cho việc gộp artists
type MergeArtistRequest struct {
	TargetID  uuid.UUID   `json:"target_id" binding:"required"`
	SourceIDs []uuid.UUID `json:"source_ids" binding:"required,min=1"`
}

// ListArtistsRequest là DTO cho việc lấy danh sách artists
type ListArtistsRequest struct {
	Page           int     `form:"page" binding:"omitempty,min=1"`
	Limit          int     `form:"limit" binding:"omitempty,min=1,max=100"`
	Search         string  `form:"search" binding:"omitempty,max=100"`
	SortBy         string  `form:"sort_by" binding:"omitempty,oneof=name novels created"`
	SortOrder      string  `form:"sort_order" binding:"omitempty,oneof=asc desc"`
	Specialization *string `form:"specialization"`
	IsVerified     *bool   `form:"is_verified"`
}
