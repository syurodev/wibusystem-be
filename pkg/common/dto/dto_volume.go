package dto

import (
	"time"
)

// CreateVolumeRequest represents the payload for creating a new volume
// This follows the API design spec from /services/catalog/api-design/novel.md section 2.2
type CreateVolumeRequest struct {
	VolumeNumber int     `json:"volume_number" validate:"required,min=1"` // Volume number in series (required)
	Title        *string `json:"title,omitempty" validate:"omitempty,max=500"` // Volume title (optional)
	Description  *string `json:"description,omitempty" validate:"omitempty,max=5000"` // Volume description (optional)
	CoverImage   *string `json:"cover_image,omitempty" validate:"omitempty,url,max=1000"` // Cover image URL (optional)
	IsPublic     bool    `json:"is_public"` // Public visibility flag
	PriceCoins   *int    `json:"price_coins,omitempty" validate:"omitempty,min=0"` // Purchase price in coins (optional)
}

// UpdateVolumeRequest represents the payload for updating an existing volume
// This follows the API design spec from /services/catalog/api-design/novel.md section 2.4
type UpdateVolumeRequest struct {
	Title       *string `json:"title,omitempty" validate:"omitempty,max=500"` // Volume title (optional)
	Description *string `json:"description,omitempty" validate:"omitempty,max=5000"` // Volume description (optional)
	CoverImage  *string `json:"cover_image,omitempty" validate:"omitempty,url,max=1000"` // Cover image URL (optional)
	IsPublic    *bool   `json:"is_public,omitempty"` // Public visibility flag (optional)
	PriceCoins  *int    `json:"price_coins,omitempty" validate:"omitempty,min=0"` // Purchase price in coins (optional)
}

// ListVolumesRequest represents query parameters for listing volumes
// This follows the API design spec from /services/catalog/api-design/novel.md section 2.1
type ListVolumesRequest struct {
	Page            int  `form:"page" validate:"omitempty,min=1"` // Current page number (default: 1)
	Limit           int  `form:"limit" validate:"omitempty,min=1,max=100"` // Items per page (default: 20, max: 100)
	IncludeChapters bool `form:"include_chapters"` // Include chapters list in response (default: false)
}

// VolumeResponse represents a volume in list/detail responses
// This follows the API design spec from /services/catalog/api-design/novel.md section 2.1
type VolumeResponse struct {
	ID           string     `json:"id"` // Volume UUID
	NovelID      string     `json:"novel_id"` // Parent novel UUID
	VolumeNumber int        `json:"volume_number"` // Volume number in series
	Title        *string    `json:"title,omitempty"` // Volume title (optional)
	Description  *string    `json:"description,omitempty"` // Volume description (optional)
	CoverImage   *string    `json:"cover_image,omitempty"` // Cover image URL (optional)
	PublishedAt  *time.Time `json:"published_at,omitempty"` // Publication date (optional)
	IsPublic     bool       `json:"is_public"` // Public visibility flag
	PriceCoins   *int       `json:"price_coins,omitempty"` // Purchase price in coins (optional)
	ChapterCount int        `json:"chapter_count"` // Number of chapters in this volume
	CreatedAt    time.Time  `json:"created_at"` // Creation timestamp
	UpdatedAt    time.Time  `json:"updated_at"` // Last update timestamp

	// Chapters field is only populated when IncludeChapters=true
	Chapters []interface{} `json:"chapters,omitempty"` // Chapter list (optional, populated based on request)
}

// CreateVolumeResponse represents the response after creating a volume
// This follows the API design spec from /services/catalog/api-design/novel.md section 2.2
type CreateVolumeResponse struct {
	ID           string    `json:"id"` // Volume UUID
	NovelID      string    `json:"novel_id"` // Parent novel UUID
	VolumeNumber int       `json:"volume_number"` // Volume number in series
	Title        *string   `json:"title,omitempty"` // Volume title (optional)
	CreatedAt    time.Time `json:"created_at"` // Creation timestamp
}

// UpdateVolumeResponse represents the response after updating a volume
// This follows the API design spec from /services/catalog/api-design/novel.md section 2.4
type UpdateVolumeResponse struct {
	ID        string    `json:"id"` // Volume UUID
	Title     *string   `json:"title,omitempty"` // Volume title (optional)
	UpdatedAt time.Time `json:"updated_at"` // Last update timestamp
}

// PaginatedVolumesResponse represents a paginated list of volumes
// This follows the API design spec from /services/catalog/api-design/novel.md section 2.1
type PaginatedVolumesResponse struct {
	Volumes    []VolumeResponse `json:"volumes"` // List of volumes
	Pagination PaginationMeta   `json:"pagination"` // Pagination metadata
}