package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid/v5"
)

// NovelChapterStatus định nghĩa trạng thái của chapter
type NovelChapterStatus string

const (
	NovelChapterStatusDraft     NovelChapterStatus = "draft"
	NovelChapterStatusPublished NovelChapterStatus = "published"
	NovelChapterStatusScheduled NovelChapterStatus = "scheduled"
)

// NovelChapter là domain model cho chapter/chương trong hệ thống
// Đây là cấp thấp nhất trong cấu trúc phân cấp Novel > NovelVolume > NovelChapter
type NovelChapter struct {
	ID       uuid.UUID
	NovelID  uuid.UUID
	VolumeID *uuid.UUID // Nullable: chapter có thể tồn tại mà không thuộc volume

	ChapterNumber int
	Title         string
	Slug          string // SEO-friendly URL

	// Content lưu trữ dạng JSONB cho nội dung phong phú
	// Ví dụ: {\"blocks\": [{\"type\": \"paragraph\", \"content\": \"...\"}], \"version\": \"1.0\"}
	Content json.RawMessage

	// Metadata về nội dung
	WordCount      int
	CharacterCount int

	// Kiểm soát truy cập
	IsFree   bool
	Price    *float64
	Currency *string

	// Trạng thái và hiển thị
	Status NovelChapterStatus

	// Thống kê
	ViewCount    int64
	LikeCount    int
	CommentCount int

	// Sắp xếp
	DisplayOrder int

	// Author notes cũng là JSONB cho nội dung phong phú
	AuthorNotes json.RawMessage

	// Ngày xuất bản
	PublishedAt *time.Time
	ScheduledAt *time.Time // Cho xuất bản theo lịch

	// Audit fields
	CreatedBy uuid.UUID
	UpdatedBy *uuid.UUID
	DeletedBy *uuid.UUID
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	// Relationships (loaded via JOIN)
	Novel  *Novel       `db:"-"` // Optional: được load bởi JOIN query
	Volume *NovelVolume `db:"-"` // Optional: được load bởi JOIN query
}

// NovelChapterRepository định nghĩa interface cho việc truy cập dữ liệu chapter
type NovelChapterRepository interface {
	// GetByID lấy chapter theo ID
	GetByID(ctx context.Context, id uuid.UUID) (*NovelChapter, error)

	// GetByNovelIDAndNumber lấy chapter theo novel ID và chapter number
	GetByNovelIDAndNumber(ctx context.Context, novelID uuid.UUID, chapterNumber int) (*NovelChapter, error)

	// GetByNovelID lấy danh sách chapter theo novel ID
	GetByNovelID(ctx context.Context, novelID uuid.UUID, filter NovelChapterFilter) ([]*NovelChapter, error)

	// GetByVolumeID lấy danh sách chapter theo volume ID
	GetByVolumeID(ctx context.Context, volumeID uuid.UUID, publishedOnly bool) ([]*NovelChapter, error)

	// Create tạo chapter mới
	Create(ctx context.Context, chapter *NovelChapter) error

	// Update cập nhật thông tin chapter
	Update(ctx context.Context, chapter *NovelChapter) error

	// Delete xóa mềm chapter
	Delete(ctx context.Context, id uuid.UUID) error

	// Publish xuất bản chapter ngay lập tức
	Publish(ctx context.Context, id uuid.UUID) error

	// Schedule đặt lịch xuất bản chapter
	Schedule(ctx context.Context, id uuid.UUID, scheduledAt time.Time) error

	// GetScheduledChapters lấy danh sách chapter cần xuất bản
	GetScheduledChapters(ctx context.Context, before time.Time) ([]*NovelChapter, error)

	// IncrementViewCount tăng view count
	IncrementViewCount(ctx context.Context, id uuid.UUID) error

	// BatchIncrementViewCount tăng view count cho nhiều chapters cùng lúc
	// Sử dụng bulk UPDATE với VALUES để tối ưu performance
	BatchIncrementViewCount(ctx context.Context, increments map[uuid.UUID]int64) error

	// UpdateStatistics cập nhật thống kê của chapter
	// UpdateStatistics cập nhật thống kê của chapter
	UpdateStatistics(ctx context.Context, id uuid.UUID, stats NovelChapterStatistics) error

	// GetRecentChapters retrieves recently published chapters across all novels
	GetRecentChapters(ctx context.Context, limit int) ([]*NovelChapter, error)
}

// NovelChapterFilter định nghĩa các filter cho việc query chapter
type NovelChapterFilter struct {
	Status        *NovelChapterStatus
	VolumeID      *uuid.UUID
	PublishedOnly bool
	IsFree        *bool
	Limit         int
	Offset        int
	SortBy        string // "chapter_number", "published_at", "views"
	SortOrder     string // "asc", "desc"
}

// NovelChapterStatistics chứa thông tin thống kê để update
type NovelChapterStatistics struct {
	ViewCount    *int64
	LikeCount    *int
	CommentCount *int
}

// NovelChapterSummary là phiên bản rút gọn của chapter (không có content)
// Dùng cho danh sách chapter để giảm kích thước response
type NovelChapterSummary struct {
	ID            uuid.UUID
	NovelID       uuid.UUID
	VolumeID      *uuid.UUID
	ChapterNumber int
	Title         string
	Slug          string
	WordCount     int
	IsFree        bool
	Price         *float64
	Currency      *string
	Status        NovelChapterStatus
	ViewCount     int64
	PublishedAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
