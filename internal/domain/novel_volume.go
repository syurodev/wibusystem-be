package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

// NovelVolume là domain model cho volume/tập trong hệ thống
// Đây là cấp giữa trong cấu trúc phân cấp Novel > NovelVolume > NovelChapter
type NovelVolume struct {
	ID      uuid.UUID
	NovelID uuid.UUID

	VolumeNumber int
	Title        string
	Slug         string // SEO-friendly URL
	Description  *string

	CoverImageURL *string

	// Thống kê
	ChapterCount int
	WordCount    int64

	// Sắp xếp và trạng thái
	DisplayOrder int
	IsPublished  bool

	// Ngày xuất bản
	PublishedAt *time.Time

	// Audit fields
	CreatedBy uuid.UUID
	UpdatedBy *uuid.UUID
	DeletedBy *uuid.UUID
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time

	// Relationships (loaded via JOIN)
	Novel      *Novel `db:"-"` // Optional: được load bởi JOIN query
	NovelTitle string // Novel title for display
}

// NovelVolumeRepository định nghĩa interface cho việc truy cập dữ liệu volume
type NovelVolumeRepository interface {
	// GetByID lấy volume theo ID
	GetByID(ctx context.Context, id uuid.UUID) (*NovelVolume, error)

	// GetByNovelIDAndNumber lấy volume theo novel ID và volume number
	GetByNovelIDAndNumber(ctx context.Context, novelID uuid.UUID, volumeNumber int) (*NovelVolume, error)

	// GetByNovelID lấy danh sách volume theo novel ID
	GetByNovelID(ctx context.Context, novelID uuid.UUID, publishedOnly bool) ([]*NovelVolume, error)

	// Create tạo volume mới
	Create(ctx context.Context, volume *NovelVolume) error

	// Update cập nhật thông tin volume
	Update(ctx context.Context, volume *NovelVolume) error

	// Delete xóa mềm volume
	Delete(ctx context.Context, id uuid.UUID) error

	// UpdateDisplayOrder cập nhật thứ tự hiển thị của volume
	UpdateDisplayOrder(ctx context.Context, id uuid.UUID, order int) error

	// Publish xuất bản volume
	Publish(ctx context.Context, id uuid.UUID) error

	// Unpublish ẩn volume
	Unpublish(ctx context.Context, id uuid.UUID) error

	// UpdateStatistics cập nhật chapter_count và word_count dựa trên chapters
	UpdateStatistics(ctx context.Context, volumeID uuid.UUID) error
}

// NovelVolumeWithChapters chứa volume và danh sách chapter của nó
type NovelVolumeWithChapters struct {
	Volume   *NovelVolume
	Chapters []*NovelChapter
}
