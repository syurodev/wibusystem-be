package artist

// CreateArtistRequest là DTO cho việc tạo artist mới
type CreateArtistRequest struct {
	Name           string  `json:"name" binding:"required,min=1,max=200"`
	Biography      string  `json:"biography" binding:"max=5000"`
	AvatarURL      *string `json:"avatar_url,omitempty" binding:"omitempty,url"`
	SocialLinks    *string `json:"social_links,omitempty"` // JSON string
	Specialization *string `json:"specialization,omitempty" binding:"omitempty,max=100"`
}

// UpdateArtistRequest là DTO cho việc cập nhật artist
type UpdateArtistRequest struct {
	Name           string  `json:"name" binding:"required,min=1,max=200"`
	Biography      string  `json:"biography" binding:"max=5000"`
	AvatarURL      *string `json:"avatar_url,omitempty" binding:"omitempty,url"`
	SocialLinks    *string `json:"social_links,omitempty"` // JSON string
	Specialization *string `json:"specialization,omitempty" binding:"omitempty,max=100"`
}

// ArtistResponse là DTO cho response artist (cho danh sách)
type ArtistResponse struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Slug           string  `json:"slug"`
	Description    *string `json:"description,omitempty"` // Biography
	NovelCount     int     `json:"novel_count"`
	Specialization *string `json:"specialization,omitempty"`
	CreatedAt      string  `json:"created_at"` // Ngày tạo
}

// ListArtistsRequest là DTO cho việc lấy danh sách artists
type ListArtistsRequest struct {
	Page           int     `form:"page" binding:"omitempty,min=1"`
	Limit          int     `form:"limit" binding:"omitempty,min=1,max=100"`
	Search         string  `form:"search" binding:"omitempty,max=100"`                    // Tìm kiếm theo tên
	SortBy         string  `form:"sort_by" binding:"omitempty,oneof=name novels created"` // name, novels, created
	SortOrder      string  `form:"sort_order" binding:"omitempty,oneof=asc desc"`         // asc hoặc desc
	Specialization *string `form:"specialization"`                                        // Filter theo specialization
	IsVerified     *bool   `form:"is_verified"`                                           // Filter theo verified status
}

// ArtistDetailResponse là DTO chi tiết cho artist
type ArtistDetailResponse struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Slug           string  `json:"slug"`
	Description    *string `json:"description,omitempty"` // Biography
	AvatarURL      *string `json:"avatar_url,omitempty"`
	SocialLinks    *string `json:"social_links,omitempty"`   // JSON string
	Specialization *string `json:"specialization,omitempty"` // cover_artist, illustrator, etc.
	NovelCount     int     `json:"novel_count"`
	ArtworkCount   int     `json:"artwork_count"`
	FollowerCount  int     `json:"follower_count"`
	IsVerified     bool    `json:"is_verified"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}
