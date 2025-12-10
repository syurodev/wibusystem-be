package author

import (
	novelres "system/internal/dto/novel"
)

// Alias for AffectedNovel from novel dto
type AffectedNovel = novelres.AffectedNovel

// AuthorResponse là DTO cho response author (cho danh sách)
type AuthorResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
	NovelCount  int     `json:"novel_count"`
	TotalViews  int64   `json:"total_views"`
	CreatedAt   string  `json:"created_at"`
}

// AuthorDetailResponse là DTO chi tiết cho author
type AuthorDetailResponse struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Slug          string  `json:"slug"`
	Description   *string `json:"description,omitempty"`
	AvatarURL     *string `json:"avatar_url,omitempty"`
	SocialLinks   *string `json:"social_links,omitempty"`
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
