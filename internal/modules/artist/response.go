package artist

import "github.com/gofrs/uuid/v5"

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

// PreviewMergeArtistResponse là response preview trước khi merge
type PreviewMergeArtistResponse struct {
	AffectedNovels []AffectedNovel `json:"affected_novels"`
	SourceArtists  []string        `json:"source_artists"`
}

// AffectedNovel là thông tin novel bị ảnh hưởng khi merge
type AffectedNovel struct {
	ID            uuid.UUID `json:"id"`
	Title         string    `json:"title"`
	Slug          string    `json:"slug"`
	CoverImageURL *string   `json:"cover_image_url"`
}
