package novel_volume

// CreateVolumeRequest represents the request to create a new volume
type CreateVolumeRequest struct {
	NovelID       string  `json:"novel_id" binding:"omitempty,uuid"` // Optional, handler uses URL param
	Title         string  `json:"title" binding:"required,min=1,max=500"`
	Description   *string `json:"description,omitempty" binding:"omitempty,max=5000"`
	CoverImageURL *string `json:"cover_image_url,omitempty"` // No url validation - empty string is valid
	DisplayOrder  int     `json:"display_order" binding:"min=0"`
	IsPublished   bool    `json:"is_published"`
}

// UpdateVolumeRequest represents the request to update a volume
type UpdateVolumeRequest struct {
	VolumeNumber  int     `json:"volume_number" binding:"required,min=1"`
	Title         string  `json:"title" binding:"required,min=1,max=500"`
	Description   *string `json:"description,omitempty" binding:"omitempty,max=5000"`
	CoverImageURL *string `json:"cover_image_url,omitempty" binding:"omitempty"`
	DisplayOrder  int     `json:"display_order" binding:"min=0"`
	IsPublished   bool    `json:"is_published"`
}

// UpdateDisplayOrderRequest represents the request to update volume display order
type UpdateDisplayOrderRequest struct {
	DisplayOrder int `json:"display_order" binding:"required,min=0"`
}

// ListVolumesRequest represents query parameters for listing volumes
type ListVolumesRequest struct {
	PublishedOnly bool `form:"published_only"`
}
