package novel_volume

// VolumeResponse represents the basic volume information in a list
type VolumeResponse struct {
	ID            string  `json:"id"`
	NovelID       string  `json:"novel_id"`
	NovelTitle    string  `json:"novel_title"` // Added: Novel title for display
	VolumeNumber  int     `json:"volume_number"`
	Title         string  `json:"title"`
	Slug          string  `json:"slug"`
	CoverImageURL *string `json:"cover_image_url,omitempty"`
	ChapterCount  int     `json:"chapter_count"`
	WordCount     int64   `json:"word_count"`
	DisplayOrder  int     `json:"display_order"`
	IsPublished   bool    `json:"is_published"`
	PublishedAt   *string `json:"published_at,omitempty"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// VolumeDetailResponse represents the detailed volume information
type VolumeDetailResponse struct {
	ID            string  `json:"id"`
	NovelID       string  `json:"novel_id"`
	VolumeNumber  int     `json:"volume_number"`
	Title         string  `json:"title"`
	Slug          string  `json:"slug"`
	Description   *string `json:"description,omitempty"`
	CoverImageURL *string `json:"cover_image_url,omitempty"`
	ChapterCount  int     `json:"chapter_count"`
	WordCount     int64   `json:"word_count"`
	DisplayOrder  int     `json:"display_order"`
	IsPublished   bool    `json:"is_published"`
	PublishedAt   *string `json:"published_at,omitempty"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

// ListVolumesResponse wraps volumes list with novel info
type ListVolumesResponse struct {
	NovelID    string           `json:"novel_id"`
	NovelTitle string           `json:"novel_title"`
	Volumes    []VolumeResponse `json:"volumes"`
}
