package novel_volume

// CreateVolumeRequest represents the request to create a new volume
type CreateVolumeRequest struct {
	NovelID      string  `json:"novel_id" binding:"required,uuid"`
	VolumeNumber int     `json:"volume_number" binding:"required,min=1"`
	Title        string  `json:"title" binding:"required,min=1,max=500"`
	Description  *string `json:"description,omitempty" binding:"omitempty,max=5000"`
	CoverImageURL *string `json:"cover_image_url,omitempty" binding:"omitempty,url"`
	DisplayOrder int     `json:"display_order" binding:"min=0"`
	IsPublished  bool    `json:"is_published"`
}

// UpdateVolumeRequest represents the request to update a volume
type UpdateVolumeRequest struct {
	VolumeNumber int     `json:"volume_number" binding:"required,min=1"`
	Title        string  `json:"title" binding:"required,min=1,max=500"`
	Description  *string `json:"description,omitempty" binding:"omitempty,max=5000"`
	CoverImageURL *string `json:"cover_image_url,omitempty" binding:"omitempty,url"`
	DisplayOrder int     `json:"display_order" binding:"min=0"`
	IsPublished  bool    `json:"is_published"`
}

// UpdateDisplayOrderRequest represents the request to update volume display order
type UpdateDisplayOrderRequest struct {
	DisplayOrder int `json:"display_order" binding:"required,min=0"`
}

// VolumeResponse represents the basic volume information in a list
type VolumeResponse struct {
	ID           string  `json:"id"`
	NovelID      string  `json:"novel_id"`
	VolumeNumber int     `json:"volume_number"`
	Title        string  `json:"title"`
	Slug         string  `json:"slug"`
	CoverImageURL *string `json:"cover_image_url,omitempty"`
	ChapterCount int     `json:"chapter_count"`
	WordCount    int64   `json:"word_count"`
	DisplayOrder int     `json:"display_order"`
	IsPublished  bool    `json:"is_published"`
	PublishedAt  *string `json:"published_at,omitempty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// VolumeDetailResponse represents the detailed volume information
type VolumeDetailResponse struct {
	ID           string  `json:"id"`
	NovelID      string  `json:"novel_id"`
	VolumeNumber int     `json:"volume_number"`
	Title        string  `json:"title"`
	Slug         string  `json:"slug"`
	Description  *string `json:"description,omitempty"`
	CoverImageURL *string `json:"cover_image_url,omitempty"`
	ChapterCount int     `json:"chapter_count"`
	WordCount    int64   `json:"word_count"`
	DisplayOrder int     `json:"display_order"`
	IsPublished  bool    `json:"is_published"`
	PublishedAt  *string `json:"published_at,omitempty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// ListVolumesRequest represents query parameters for listing volumes
type ListVolumesRequest struct {
	PublishedOnly bool `form:"published_only"`
}
