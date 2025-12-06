package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid/v5"
)

// Author là domain model cho tác giả
type Author struct {
	ID     uuid.UUID
	UserID *uuid.UUID // Link đến user account nếu đã đăng ký

	Name string
	Slug string

	// Biography lưu trữ dạng JSONB
	Biography json.RawMessage

	AvatarURL *string

	// Social links lưu trữ dạng JSONB
	// Ví dụ: {"facebook": "...", "twitter": "...", "website": "..."}
	SocialLinks json.RawMessage

	// Thống kê
	NovelCount     int
	TotalChapters  int
	TotalViews     int64
	FollowerCount  int

	// Status
	IsVerified bool

	// Audit
	CreatedBy uuid.UUID
	UpdatedBy *uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	DeletedBy *uuid.UUID
}

// AuthorRepository định nghĩa interface cho việc truy cập dữ liệu author
type AuthorRepository interface {
	// GetByID lấy author theo ID
	GetByID(ctx context.Context, id uuid.UUID) (*Author, error)

	// GetBySlug lấy author theo slug
	GetBySlug(ctx context.Context, slug string) (*Author, error)

	// GetByUserID lấy author theo user ID
	GetByUserID(ctx context.Context, userID uuid.UUID) (*Author, error)

	// List lấy danh sách authors với filter
	List(ctx context.Context, filter AuthorFilter) ([]*Author, int64, error)

	// ListSelection lấy danh sách authors rút gọn (chỉ ID và Name)
	ListSelection(ctx context.Context, offset, limit int, search string) ([]*Author, int64, error)

	// Create tạo author mới
	Create(ctx context.Context, author *Author) error

	// Update cập nhật author
	Update(ctx context.Context, author *Author) error

	// Delete xóa mềm author
	Delete(ctx context.Context, id uuid.UUID) error

	// GetNovelAuthors lấy danh sách authors của một novel
	GetNovelAuthors(ctx context.Context, novelID uuid.UUID) ([]*NovelAuthor, error)

	// AddNovelAuthor thêm author cho novel
	AddNovelAuthor(ctx context.Context, novelID, authorID uuid.UUID, displayOrder int) error

	// RemoveNovelAuthor xóa author khỏi novel
	RemoveNovelAuthor(ctx context.Context, novelID, authorID uuid.UUID) error

	// UpdateStatistics cập nhật thống kê
	UpdateStatistics(ctx context.Context, id uuid.UUID, stats AuthorStatistics) error

	// Merge gộp nhiều authors thành một
	Merge(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID, mergedBy uuid.UUID) error

	// GetMergePreview lấy danh sách các novel sẽ bị ảnh hưởng khi merge
	GetMergePreview(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID) ([]*Novel, error)
}

// AuthorFilter định nghĩa các filter cho việc query authors
type AuthorFilter struct {
	SearchQuery *string
	IsVerified  *bool
	SortBy      string // "name", "novel_count", "follower_count"
	SortOrder   string // "asc", "desc"
	Limit       int
	Offset      int
}

// AuthorStatistics chứa thông tin thống kê để update
type AuthorStatistics struct {
	NovelCount    *int
	TotalChapters *int
	TotalViews    *int64
	FollowerCount *int
}

// NovelAuthor chứa thông tin về author và role trong novel
type NovelAuthor struct {
	Author       *Author
	Role         string // "original_author", "co_author"
	DisplayOrder int
}
