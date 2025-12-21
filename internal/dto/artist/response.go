package artist

import (
	novelres "system/internal/dto/novel"
)

// Alias for AffectedNovel from novel dto
type AffectedNovel = novelres.AffectedNovel

// ArtistResponse là DTO cho response artist (cho danh sách)
type ArtistResponse struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Slug           string  `json:"slug"`
	Description    *string `json:"description,omitempty"`
	NovelCount     int     `json:"novel_count"`
	Specialization *string `json:"specialization,omitempty"`
	PortfolioURL   *string `json:"portfolio_url,omitempty"`
	CreatedAt      string  `json:"created_at"`
}

// ArtistDetailResponse là DTO chi tiết cho artist
type ArtistDetailResponse struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Slug           string  `json:"slug"`
	Description    *string `json:"description,omitempty"`
	AvatarURL      *string `json:"avatar_url,omitempty"`
	SocialLinks    *string `json:"social_links,omitempty"`
	Specialization *string `json:"specialization,omitempty"`
	PortfolioURL   *string `json:"portfolio_url,omitempty"`
	NovelCount     int     `json:"novel_count"`
	ArtworkCount   int     `json:"artwork_count"`
	FollowerCount  int     `json:"follower_count"`
	IsVerified     bool    `json:"is_verified"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// PreviewMergeArtistResponse là response preview trước khi merge
type PreviewMergeArtistResponse struct {
	AffectedNovels []AffectedNovel `json:"affected_novels"`
	SourceArtists  []string        `json:"source_artists"`
}

// SelectionResponse là DTO cho dropdown selection (ID + Name)
// Sử dụng cho các select fields khi cần chọn artist
type SelectionResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
