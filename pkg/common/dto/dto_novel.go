package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CreateNovelRequest represents the payload for creating a new novel
type CreateNovelRequest struct {
	// Core content fields
	Title      string           `json:"title" validate:"required,max=1000"`     // Tên novel (bắt buộc) - theo API design
	CoverImage string           `json:"cover_image" validate:"max=1000"`        // URL ảnh bìa
	Summary    *json.RawMessage `json:"summary,omitempty"`                      // Tóm tắt đa ngôn ngữ (JSONB)
	Genres     []string         `json:"genres" validate:"dive,uuid"`            // Mảng UUID genres

	// Creator and Character associations
	Creators   []CreatorRole    `json:"creators,omitempty" validate:"dive"`       // Mảng creators với vai trò
	Characters []string         `json:"characters,omitempty" validate:"dive,uuid"` // Mảng UUID characters

	// Publishing information
	PublishedAt      *time.Time `json:"published_at,omitempty"`                     // Ngày xuất bản (optional)
	OriginalLanguage string     `json:"original_language" validate:"max=5"`         // Ngôn ngữ gốc (ISO 639-1)
	SourceURL        string     `json:"source_url,omitempty" validate:"max=2000"`   // URL nguồn gốc
	ISBN             string     `json:"isbn,omitempty" validate:"max=17"`           // ISBN code

	// Content rating và warnings
	AgeRating       string           `json:"age_rating,omitempty" validate:"oneof=G PG PG-13 R NC-17"` // Phân loại độ tuổi
	ContentWarnings *json.RawMessage `json:"content_warnings,omitempty"`                               // Cảnh báo nội dung (JSONB)
	MatureContent   bool             `json:"mature_content"`                                            // Nội dung người lớn

	// Visibility settings
	IsPublic    bool `json:"is_public"`    // Công khai hay riêng tư
	IsFeatured  bool `json:"is_featured"`  // Được đề xuất trên trang chủ
	IsCompleted bool `json:"is_completed"` // Đã hoàn thành

	// SEO fields
	Slug            string           `json:"slug,omitempty" validate:"max=255"`      // URL-friendly identifier
	Tags            *json.RawMessage `json:"tags,omitempty"`                         // Tags tìm kiếm (JSONB array)
	Keywords        string           `json:"keywords,omitempty" validate:"max=500"`  // SEO keywords
	MetaDescription string           `json:"meta_description,omitempty" validate:"max=500"` // SEO meta description

	// Pricing (optional)
	PriceCoins         *int `json:"price_coins,omitempty" validate:"omitempty,min=0"`         // Giá mua series (coins)
	RentalPriceCoins   *int `json:"rental_price_coins,omitempty" validate:"omitempty,min=0"`  // Giá thuê series (coins)
	RentalDurationDays *int `json:"rental_duration_days,omitempty" validate:"omitempty,min=1"` // Thời hạn thuê (ngày)
	IsPremium          bool `json:"is_premium"`                                               // Nội dung premium

	// Ownership fields (set from middleware/auth context)
	OwnershipType     string     `json:"ownership_type" validate:"required,oneof=PERSONAL TENANT COLLABORATIVE"` // PERSONAL, TENANT, COLLABORATIVE
	PrimaryOwnerID    uuid.UUID  `json:"-"`                                                                      // Set from context - user_id or tenant_id
	OriginalCreatorID uuid.UUID  `json:"-"`                                                                      // Set from context - user who creates
	AccessLevel       string     `json:"access_level" validate:"required,oneof=PRIVATE TENANT_ONLY PUBLIC"`     // PRIVATE, TENANT_ONLY, PUBLIC
}

// CreatorRole represents a creator with their role in the novel
type CreatorRole struct {
	CreatorID string `json:"creator_id" validate:"required,uuid"` // UUID của creator
	Role      string `json:"role" validate:"required,oneof=AUTHOR ILLUSTRATOR ARTIST STUDIO VOICE_ACTOR"` // Vai trò creator
}

// UpdateNovelRequest represents the payload for updating an existing novel
type UpdateNovelRequest struct {
	// Core content fields
	Title      *string          `json:"title,omitempty" validate:"omitempty,max=1000"` // Tên novel
	CoverImage *string          `json:"cover_image,omitempty" validate:"omitempty,max=1000"` // URL ảnh bìa
	Summary    *json.RawMessage `json:"summary,omitempty"`                            // Tóm tắt đa ngôn ngữ
	Genres     []string         `json:"genres,omitempty" validate:"dive,uuid"`        // Mảng UUID genres

	// Publishing information
	PublishedAt      *time.Time `json:"published_at,omitempty"`                       // Ngày xuất bản
	OriginalLanguage *string    `json:"original_language,omitempty" validate:"omitempty,max=5"` // Ngôn ngữ gốc
	SourceURL        *string    `json:"source_url,omitempty" validate:"omitempty,max=2000"` // URL nguồn gốc
	ISBN             *string    `json:"isbn,omitempty" validate:"omitempty,max=17"`   // ISBN code

	// Content rating
	AgeRating       *string          `json:"age_rating,omitempty" validate:"omitempty,oneof=G PG PG-13 R NC-17"` // Phân loại độ tuổi
	ContentWarnings *json.RawMessage `json:"content_warnings,omitempty"`                                         // Cảnh báo nội dung
	MatureContent   *bool            `json:"mature_content,omitempty"`                                            // Nội dung người lớn

	// Visibility settings
	IsPublic    *bool `json:"is_public,omitempty"`    // Công khai
	IsFeatured  *bool `json:"is_featured,omitempty"`  // Được đề xuất
	IsCompleted *bool `json:"is_completed,omitempty"` // Đã hoàn thành

	// SEO fields
	Slug            *string          `json:"slug,omitempty" validate:"omitempty,max=255"` // URL identifier
	Tags            *json.RawMessage `json:"tags,omitempty"`                              // Tags tìm kiếm
	Keywords        *string          `json:"keywords,omitempty" validate:"omitempty,max=500"` // SEO keywords
	MetaDescription *string          `json:"meta_description,omitempty" validate:"omitempty,max=500"` // SEO description

	// Pricing
	PriceCoins         *int  `json:"price_coins,omitempty" validate:"omitempty,min=0"`         // Giá mua series
	RentalPriceCoins   *int  `json:"rental_price_coins,omitempty" validate:"omitempty,min=0"`  // Giá thuê series
	RentalDurationDays *int  `json:"rental_duration_days,omitempty" validate:"omitempty,min=1"` // Thời hạn thuê
	IsPremium          *bool `json:"is_premium,omitempty"`                                     // Nội dung premium
}

// ListNovelsRequest represents query parameters for listing novels
type ListNovelsRequest struct {
	// Pagination
	Page     int `form:"page" validate:"omitempty,min=1"`      // Trang hiện tại (default: 1)
	PageSize int `form:"page_size" validate:"omitempty,min=1,max=100"` // Kích thước trang (default: 20, max: 100)

	// Filtering
	Status        string `form:"status" validate:"omitempty,oneof=ONGOING COMPLETED HIATUS"` // Lọc theo trạng thái
	AgeRating     string `form:"age_rating" validate:"omitempty,oneof=G PG PG-13 R NC-17"`   // Lọc theo độ tuổi
	MatureContent *bool  `form:"mature_content"`                                             // Lọc nội dung người lớn
	IsPublic      *bool  `form:"is_public"`                                                  // Lọc công khai/riêng tư
	IsFeatured    *bool  `form:"is_featured"`                                                // Lọc nội dung đề xuất
	IsCompleted   *bool  `form:"is_completed"`                                               // Lọc đã hoàn thành
	IsPremium     *bool  `form:"is_premium"`                                                 // Lọc nội dung premium

	// Search
	Search   string   `form:"search" validate:"omitempty,max=100"`   // Tìm kiếm trong tên, description
	Tags     []string `form:"tags" validate:"dive,max=50"`           // Lọc theo tags
	GenreIDs []string `form:"genre_ids" validate:"dive,uuid"`        // Lọc theo genres

	// Sorting
	SortBy    string `form:"sort_by" validate:"omitempty,oneof=name created_at updated_at published_at view_count rating_average"` // Sắp xếp theo trường
	SortOrder string `form:"sort_order" validate:"omitempty,oneof=asc desc"`                                                      // Thứ tự sắp xếp (default: desc)

	// Date filtering
	CreatedAfter  *time.Time `form:"created_after"`  // Lọc tạo sau ngày
	CreatedBefore *time.Time `form:"created_before"` // Lọc tạo trước ngày
	PublishedAfter  *time.Time `form:"published_after"`  // Lọc xuất bản sau ngày
	PublishedBefore *time.Time `form:"published_before"` // Lọc xuất bản trước ngày

	// Ownership filters
	OwnershipType  string     `form:"ownership_type" validate:"omitempty,oneof=PERSONAL TENANT COLLABORATIVE"` // Lọc theo loại sở hữu
	PrimaryOwnerID *uuid.UUID `form:"primary_owner_id" validate:"omitempty,uuid"`                             // Lọc theo owner (user or tenant)
	CreatorID      *uuid.UUID `form:"creator_id" validate:"omitempty,uuid"`                                   // Lọc theo original creator

	// Date ranges cho latest chapter update
	LatestChapterUpdatedAfter  *time.Time `form:"latest_chapter_updated_after"`
	LatestChapterUpdatedBefore *time.Time `form:"latest_chapter_updated_before"`

	// View count range
	MinViewCount *int64 `form:"min_view_count" validate:"omitempty,min=0"`
	MaxViewCount *int64 `form:"max_view_count" validate:"omitempty,min=0"`
}

// NovelSummaryResponse - response tối ưu cho list novels
type NovelSummaryResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	CoverImage  *string   `json:"cover_image"`
	ViewCount   int64     `json:"view_count"`
	CreatedAt   time.Time `json:"created_at"`

	// Ownership info
	OwnershipType     string `json:"ownership_type"`      // PERSONAL, TENANT, COLLABORATIVE
	PrimaryOwnerID    string `json:"primary_owner_id"`    // UUID string
	OriginalCreatorID string `json:"original_creator_id"` // UUID string

	// Owner/Creator objects (populated by service layer via gRPC)
	PrimaryOwner   interface{} `json:"primary_owner,omitempty"`   // UserSummary or TenantSummary based on ownership_type
	OriginalCreator *UserSummary `json:"original_creator,omitempty"` // Always a user

	// Latest chapter info
	LatestChapterUpdatedAt *time.Time `json:"latest_chapter_updated_at"`
}

// UserSummary represents user info for novel list response
type UserSummary struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

// TenantSummary represents tenant info for novel list response
type TenantSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// PaginatedNovelsResponse - response với pagination
type PaginatedNovelsResponse struct {
	Novels     []NovelSummaryResponse `json:"novels"`
	Pagination PaginationMeta         `json:"pagination"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Page        int   `json:"page"`
	PageSize    int   `json:"page_size"`
	Total       int64 `json:"total"`
	TotalPages  int   `json:"total_pages"`
	HasNext     bool  `json:"has_next"`
	HasPrevious bool  `json:"has_previous"`
}

// CreateNovelResponse represents the response after creating a novel (theo API design)
type CreateNovelResponse struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateNovelResponse represents the response after updating a novel (theo API design)
type UpdateNovelResponse struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// NovelDetailResponse represents detailed novel information (API 1.3)
type NovelDetailResponse struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`              // Tên theo ngôn ngữ client
	CoverImage      *string                `json:"cover_image"`       // URL ảnh bìa
	Summary         map[string]interface{} `json:"summary"`           // JSON content từ Plate editor
	Status          string                 `json:"status"`            // Trạng thái
	PublishedAt     *time.Time             `json:"published_at"`      // Ngày xuất bản
	OriginalLanguage string                 `json:"original_language"` // Ngôn ngữ gốc
	CurrentLanguage string                 `json:"current_language"`  // Ngôn ngữ hiện tại
	SourceURL       *string                `json:"source_url"`        // URL nguồn
	ISBN            *string                `json:"isbn"`              // Mã ISBN
	AgeRating       *string                `json:"age_rating"`        // Phân loại độ tuổi
	ContentWarnings []string               `json:"content_warnings"`  // Cảnh báo nội dung
	MatureContent   bool                   `json:"mature_content"`    // Nội dung người lớn
	IsPublic        bool                   `json:"is_public"`         // Công khai
	IsFeatured      bool                   `json:"is_featured"`       // Được đề xuất
	IsCompleted     bool                   `json:"is_completed"`      // Đã hoàn thành
	Slug            string                 `json:"slug"`              // URL identifier
	Keywords        *string                `json:"keywords"`          // SEO keywords
	PriceCoins      *int                   `json:"price_coins"`       // Giá mua series
	RentalPriceCoins *int                  `json:"rental_price_coins"` // Giá thuê series
	RentalDurationDays *int                `json:"rental_duration_days"` // Thời hạn thuê
	IsPremium       bool                   `json:"is_premium"`        // Nội dung premium
	ViewCount       int64                  `json:"view_count"`        // Lượt xem
	RatingAverage   *float64               `json:"rating_average"`    // Điểm đánh giá TB
	RatingCount     int                    `json:"rating_count"`      // Số lượt đánh giá
	ChapterCount    int                    `json:"chapter_count"`     // Số chương
	VolumeCount     int                    `json:"volume_count"`      // Số tập
	Genres          []GenreInfo            `json:"genres"`            // Danh sách thể loại
	Creators        []CreatorInfo          `json:"creators"`          // Danh sách creator
	Characters      []CharacterInfo        `json:"characters"`        // Danh sách nhân vật
	Translations    []interface{}          `json:"translations"`      // Bản dịch (nếu có)
	Stats           map[string]interface{} `json:"stats"`             // Thống kê (nếu có)
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// Supporting types for NovelDetailResponse
type GenreInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CreatorInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type CharacterInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
