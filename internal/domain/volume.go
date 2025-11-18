package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

// Volume là domain model cho volume/tập trong hệ thống
// Đây là cấp giữa trong cấu trúc phân cấp Novel > Volume > Chapter
type Volume struct {
	ID       uuid.UUID
	NovelID  uuid.UUID
	Novel    *Novel // Optional: được load bởi JOIN query

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
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// VolumeRepository định nghĩa interface cho việc truy cập dữ liệu volume
type VolumeRepository interface {
	// GetByID lấy volume theo ID
	GetByID(ctx context.Context, id uuid.UUID) (*Volume, error)

	// GetByNovelIDAndNumber lấy volume theo novel ID và volume number
	GetByNovelIDAndNumber(ctx context.Context, novelID uuid.UUID, volumeNumber int) (*Volume, error)

	// GetByNovelID lấy danh sách volume theo novel ID
	GetByNovelID(ctx context.Context, novelID uuid.UUID, publishedOnly bool) ([]*Volume, error)

	// Create tạo volume mới
	Create(ctx context.Context, volume *Volume) error

	// Update cập nhật thông tin volume
	Update(ctx context.Context, volume *Volume) error

	// Delete xóa mềm volume
	Delete(ctx context.Context, id uuid.UUID) error

	// UpdateDisplayOrder cập nhật thứ tự hiển thị của volume
	UpdateDisplayOrder(ctx context.Context, id uuid.UUID, order int) error

	// Publish xuất bản volume
	Publish(ctx context.Context, id uuid.UUID) error

	// Unpublish ẩn volume
	Unpublish(ctx context.Context, id uuid.UUID) error
}

// VolumeWithChapters chứa volume và danh sách chapter của nó
type VolumeWithChapters struct {
	Volume   *Volume
	Chapters []*Chapter
}
