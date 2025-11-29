package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid/v5"
)

// NovelStatus định nghĩa trạng thái của novel
type NovelStatus string

const (
	NovelStatusDraft     NovelStatus = "draft"
	NovelStatusOngoing   NovelStatus = "ongoing"
	NovelStatusCompleted NovelStatus = "completed"
	NovelStatusHiatus    NovelStatus = "hiatus"
	NovelStatusDropped   NovelStatus = "dropped"
)

// Novel là domain model cho tiểu thuyết trong hệ thống
// Đây là cấp cao nhất trong cấu trúc phân cấp Novel > Volume > Chapter
type Novel struct {
	ID    uuid.UUID
	Title string
	Slug  string // SEO-friendly URL
	
	// Owner information
	OwnerID   uuid.UUID
	OwnerType string // "user" or "group"

	// Synopsis lưu trữ dạng JSONB cho nội dung phong phú
	// Ví dụ: {"blocks": [{"type": "paragraph", "content": "..."}], "language": "vi"}
	Synopsis json.RawMessage

	CoverImageURL *string
	ThumbnailURL  *string

	Status NovelStatus

	// Original information
	OriginalLanguage *string // ISO 639-1 language code
	OriginalTitle    *string

	// Thống kê
	TotalVolumes  int
	TotalChapters int
	TotalWords    int64
	ViewCount     int64
	FavoriteCount int
	RatingAverage float64 // 0.00 đến 5.00
	RatingCount   int

	// Metadata lưu trữ dạng JSONB cho thông tin bổ sung
	// Ví dụ: {"external_ids": {"myanimelist": "123"}, "custom_fields": {...}}
	Metadata json.RawMessage

	// Ngày xuất bản
	FirstPublishedAt *time.Time
	LastChapterAt    *time.Time
	CompletedAt      *time.Time

	// Audit fields
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	// Relationships (loaded separately via JOINs)
	Authors     []*NovelAuthor     // Authors of this novel
	Artists     []*NovelArtist     // Artists of this novel
	Translators []*NovelTranslator // Translators of this novel
	Genres      []*Genre           // Genres of this novel
}

// NovelRepository định nghĩa interface cho việc truy cập dữ liệu novel
type NovelRepository interface {
	// GetByID lấy novel theo ID
	GetByID(ctx context.Context, id uuid.UUID) (*Novel, error)

	// GetBySlug lấy novel theo slug
	GetBySlug(ctx context.Context, slug string) (*Novel, error)

	// Create tạo novel mới
	Create(ctx context.Context, novel *Novel) error

	// Update cập nhật thông tin novel
	Update(ctx context.Context, novel *Novel) error

	// Delete xóa mềm novel
	Delete(ctx context.Context, id uuid.UUID) error

	// List lấy danh sách novel với filter và pagination
	List(ctx context.Context, filter NovelFilter) ([]*Novel, int64, error)

	// UpdateStatistics cập nhật thống kê của novel
	UpdateStatistics(ctx context.Context, id uuid.UUID, stats NovelStatistics) error

	// IncrementViewCount tăng view count
	IncrementViewCount(ctx context.Context, id uuid.UUID) error
}

// NovelFilter định nghĩa các filter cho việc query novel
type NovelFilter struct {
	Status           *NovelStatus
	GenreIDs         []uuid.UUID // Filter by genre IDs
	AuthorID         *uuid.UUID  // Filter by author ID (via novel_authors)
	TranslatorID     *uuid.UUID  // Filter by translator ID (via novel_translators)
	OriginalLanguage *string     // Filter by original language
	SearchQuery      *string     // Full-text search trong title và synopsis
	SortBy           string      // "created_at", "rating", "views", "last_chapter"
	SortOrder        string      // "asc", "desc"
	Limit            int
	Offset           int
	IncludeStats     bool // Include statistics trong kết quả
}

// NovelStatistics chứa thông tin thống kê để update
type NovelStatistics struct {
	ViewCount     *int64
	FavoriteCount *int
	RatingAverage *float64
	RatingCount   *int
}
