package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Novel represents a novel series master table
type Novel struct {
	// Core identification fields
	ID     uuid.UUID `json:"id" db:"id"`                       // Primary key UUID
	Status string    `json:"status" db:"status"`               // content_status enum: ONGOING, COMPLETED, HIATUS

	// Content fields
	Name       *string          `json:"name,omitempty" db:"name"`               // Primary/default title
	CoverImage *string          `json:"cover_image,omitempty" db:"cover_image"` // URL ảnh bìa novel
	Summary    *json.RawMessage `json:"summary,omitempty" db:"summary"`         // Tóm tắt đa ngôn ngữ (JSONB)

	// Ownership Model - Clean ownership tracking
	OwnershipType          string     `json:"ownership_type" db:"ownership_type"`                               // PERSONAL, TENANT, COLLABORATIVE
	PrimaryOwnerID         uuid.UUID  `json:"primary_owner_id" db:"primary_owner_id"`                           // User ID (PERSONAL) or Tenant ID (TENANT/COLLABORATIVE)
	OriginalCreatorID      uuid.UUID  `json:"original_creator_id" db:"original_creator_id"`                     // User who originally created the content - immutable
	AccessLevel            string     `json:"access_level" db:"access_level"`                                   // PRIVATE, TENANT_ONLY, PUBLIC
	LastModifiedByUserID   *uuid.UUID `json:"last_modified_by_user_id,omitempty" db:"last_modified_by_user_id"` // User who last modified
	OwnershipTransferredAt *time.Time `json:"ownership_transferred_at,omitempty" db:"ownership_transferred_at"` // When ownership was last transferred

	// Publishing information
	PublishedAt      *time.Time `json:"published_at,omitempty" db:"published_at"`           // Ngày xuất bản chính thức
	OriginalLanguage string     `json:"original_language" db:"original_language"`           // Ngôn ngữ gốc (ISO 639-1)
	SourceURL        *string    `json:"source_url,omitempty" db:"source_url"`               // URL nguồn gốc (nếu chuyển thể)
	ISBN             *string    `json:"isbn,omitempty" db:"isbn"`                           // ISBN code cho xuất bản

	// Content rating và warnings
	AgeRating       *string          `json:"age_rating,omitempty" db:"age_rating"`             // G, PG, PG-13, R, NC-17
	ContentWarnings *json.RawMessage `json:"content_warnings,omitempty" db:"content_warnings"` // Cảnh báo nội dung (JSONB)
	MatureContent   bool             `json:"mature_content" db:"mature_content"`               // Nội dung người lớn

	// Visibility & status flags
	IsPublic    bool       `json:"is_public" db:"is_public"`       // Công khai hay riêng tư
	IsFeatured  bool       `json:"is_featured" db:"is_featured"`   // Được đề xuất trên trang chủ
	IsCompleted bool       `json:"is_completed" db:"is_completed"` // Đã hoàn thành
	IsDeleted   bool       `json:"is_deleted" db:"is_deleted"`     // Soft delete flag
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"` // Thời gian xóa
	DeletedByUserID *uuid.UUID `json:"deleted_by_user_id,omitempty" db:"deleted_by_user_id"` // User thực hiện xóa

	// SEO và discovery fields
	Slug            *string          `json:"slug,omitempty" db:"slug"`                         // URL-friendly identifier
	Tags            *json.RawMessage `json:"tags,omitempty" db:"tags"`                         // Tags tìm kiếm (JSONB array)
	Keywords        *string          `json:"keywords,omitempty" db:"keywords"`                 // SEO keywords
	MetaDescription *string          `json:"meta_description,omitempty" db:"meta_description"` // SEO meta description

	// Analytics và engagement metrics
	ViewCount     int64    `json:"view_count" db:"view_count"`           // Tổng lượt xem novel
	LikeCount     int64    `json:"like_count" db:"like_count"`           // Tổng lượt thích
	BookmarkCount int64    `json:"bookmark_count" db:"bookmark_count"`   // Tổng lượt bookmark/favorite
	CommentCount  int64    `json:"comment_count" db:"comment_count"`     // Tổng số bình luận
	RatingAverage *float64 `json:"rating_average,omitempty" db:"rating_average"` // Điểm đánh giá TB (0.00-5.00)
	RatingCount   int      `json:"rating_count" db:"rating_count"`       // Số lượt đánh giá

	// Pricing và monetization
	PriceCoins         *int `json:"price_coins,omitempty" db:"price_coins"`                 // Giá mua toàn series (coins)
	RentalPriceCoins   *int `json:"rental_price_coins,omitempty" db:"rental_price_coins"`   // Giá thuê series (coins)
	RentalDurationDays *int `json:"rental_duration_days,omitempty" db:"rental_duration_days"` // Thời hạn thuê (ngày)
	IsPremium          bool `json:"is_premium" db:"is_premium"`                             // Nội dung premium

	// Content metadata
	TotalChapters        int  `json:"total_chapters" db:"total_chapters"`                           // Tổng số chương
	TotalVolumes         int  `json:"total_volumes" db:"total_volumes"`                             // Tổng số tập
	EstimatedReadingTime *int `json:"estimated_reading_time,omitempty" db:"estimated_reading_time"` // Thời gian đọc ước tính (phút)
	WordCount            *int `json:"word_count,omitempty" db:"word_count"`                         // Tổng số từ trong novel

	// Audit timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"` // Thời gian tạo record
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"` // Thời gian cập nhật cuối
}

// NovelVolume represents volumes of a novel
type NovelVolume struct {
	// Core identification
	ID           uuid.UUID `json:"id" db:"id"`                       // Primary key UUID
	NovelID      uuid.UUID `json:"novel_id" db:"novel_id"`           // Foreign key tới novel
	VolumeNumber int       `json:"volume_number" db:"volume_number"` // Số thứ tự volume trong series

	// Content fields
	VolumeTitle *string `json:"volume_title,omitempty" db:"volume_title"` // Tiêu đề volume (nếu có)
	CoverImage  *string `json:"cover_image,omitempty" db:"cover_image"`   // URL ảnh bìa volume
	Description *string `json:"description,omitempty" db:"description"`   // Mô tả ngắn volume

	// User tracking (volumes inherit ownership from novel)
	LastModifiedByUserID *uuid.UUID `json:"last_modified_by_user_id,omitempty" db:"last_modified_by_user_id"` // User who last modified volume

	// Publishing và status
	PublishedAt *time.Time `json:"published_at,omitempty" db:"published_at"` // Ngày xuất bản volume
	IsDeleted   bool       `json:"is_deleted" db:"is_deleted"`               // Soft delete flag
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`     // Thời gian xóa volume
	IsAvailable bool       `json:"is_available" db:"is_available"`           // Volume có sẵn để đọc

	// Pricing
	PriceCoins         *int `json:"price_coins,omitempty" db:"price_coins"`                 // Giá mua volume (coins)
	RentalPriceCoins   *int `json:"rental_price_coins,omitempty" db:"rental_price_coins"`   // Giá thuê volume (coins)
	RentalDurationDays *int `json:"rental_duration_days,omitempty" db:"rental_duration_days"` // Thời hạn thuê volume (ngày)

	// Content metadata
	PageCount            *int `json:"page_count,omitempty" db:"page_count"`                         // Số trang ước tính
	WordCount            *int `json:"word_count,omitempty" db:"word_count"`                         // Số từ trong volume
	ChapterCount         int  `json:"chapter_count" db:"chapter_count"`                             // Số chương trong volume
	EstimatedReadingTime *int `json:"estimated_reading_time,omitempty" db:"estimated_reading_time"` // Thời gian đọc ước tính (phút)

	// Audit timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"` // Thời gian tạo record
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"` // Thời gian cập nhật cuối
}

// NovelChapter represents chapters of a novel
type NovelChapter struct {
	// Core identification
	ID            uuid.UUID `json:"id" db:"id"`                         // Primary key UUID
	VolumeID      uuid.UUID `json:"volume_id" db:"volume_id"`           // Foreign key tới volume
	ChapterNumber int       `json:"chapter_number" db:"chapter_number"` // Số thứ tự chapter trong volume

	// Content fields
	Title   *string          `json:"title,omitempty" db:"title"`     // Tiêu đề chapter (nếu có)
	Content *json.RawMessage `json:"content,omitempty" db:"content"` // Nội dung chapter (JSONB rich text)

	// User tracking (chapters inherit ownership from novel)
	LastModifiedByUserID *uuid.UUID `json:"last_modified_by_user_id,omitempty" db:"last_modified_by_user_id"` // User who last modified chapter

	// Publishing workflow
	PublishedAt        *time.Time `json:"published_at,omitempty" db:"published_at"`               // Thời gian xuất bản chapter
	ScheduledPublishAt *time.Time `json:"scheduled_publish_at,omitempty" db:"scheduled_publish_at"` // Lên lịch xuất bản
	IsDraft            bool       `json:"is_draft" db:"is_draft"`                                 // Chapter đang ở trạng thái draft
	IsPublic           bool       `json:"is_public" db:"is_public"`                               // Chapter công khai
	IsDeleted          bool       `json:"is_deleted" db:"is_deleted"`                             // Soft delete flag
	DeletedAt          *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`                   // Thời gian xóa chapter
	Version            int        `json:"version" db:"version"`                                   // Version number của chapter

	// Content warnings
	ContentWarnings   *json.RawMessage `json:"content_warnings,omitempty" db:"content_warnings"` // Cảnh báo nội dung riêng (JSONB)
	HasMatureContent  bool             `json:"has_mature_content" db:"has_mature_content"`       // Chapter có nội dung nhạy cảm

	// Pricing
	PriceCoins *int `json:"price_coins,omitempty" db:"price_coins"` // Giá mua lẻ chapter (coins)

	// Content metadata
	WordCount           *int `json:"word_count,omitempty" db:"word_count"`                     // Số từ trong chapter
	CharacterCount      *int `json:"character_count,omitempty" db:"character_count"`           // Số ký tự trong chapter
	ReadingTimeMinutes  *int `json:"reading_time_minutes,omitempty" db:"reading_time_minutes"` // Thời gian đọc ước tính (phút)

	// Analytics
	ViewCount    int64 `json:"view_count" db:"view_count"`       // Số lượt xem chapter
	LikeCount    int64 `json:"like_count" db:"like_count"`       // Số lượt thích chapter
	CommentCount int64 `json:"comment_count" db:"comment_count"` // Số bình luận chapter

	// Audit timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"` // Thời gian tạo record
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"` // Thời gian cập nhật cuối
}
