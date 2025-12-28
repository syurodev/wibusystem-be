package shared

// OwnerInfo represents basic user info for ownership display
// Used across multiple domains: media, novel, chapter, etc.
type OwnerInfo struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"display_name"`
	Username    string  `json:"username"`
	AvatarURL   *string `json:"avatar_url,omitempty"`
	Slug        *string `json:"slug,omitempty"`
}

// GenreInfo represents basic genre information
type GenreInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// CreatorInfo represents author/artist information
type CreatorInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}
