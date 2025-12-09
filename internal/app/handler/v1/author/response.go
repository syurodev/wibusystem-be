package author

import "github.com/gofrs/uuid/v5"

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

// PreviewMergeAuthorResponse là response preview trước khi merge
type PreviewMergeAuthorResponse struct {
	AffectedNovels []AffectedNovel `json:"affected_novels"`
	SourceAuthors  []string        `json:"source_authors"`
}

// AffectedNovel là thông tin novel bị ảnh hưởng khi merge
type AffectedNovel struct {
	ID            uuid.UUID `json:"id"`
	Title         string    `json:"title"`
	Slug          string    `json:"slug"`
	CoverImageURL *string   `json:"cover_image_url"`
}
