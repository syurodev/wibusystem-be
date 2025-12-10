package domain

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
)

// Genre là domain model cho thể loại novel
type Genre struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	Description *string

	// Parent genre cho phân cấp (ví dụ: Fantasy > Xuanhuan)
	ParentID *uuid.UUID
	Parent   *Genre `db:"-"` // Optional: được load bởi JOIN query

	// Display (removed display_order)
	IsActive     bool

	// Statistics (auto-updated by application)
	NovelCount    int
	AnimeCount    int
	MangaCount    int
	ActiveReaders int64
	TotalViews    int64
	Trend         Trend `db:"-"` // Calculated field (not in DB)

	// Audit
	CreatedBy *uuid.UUID
	UpdatedBy *uuid.UUID
	DeletedBy *uuid.UUID
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// Trend định nghĩa xu hướng của genre
type Trend string

const (
	TrendRising  Trend = "rising"
	TrendStable  Trend = "stable"
	TrendFalling Trend = "falling"
)

// GenreRepository định nghĩa interface cho việc truy cập dữ liệu genre
type GenreRepository interface {
	// GetByID lấy genre theo ID
	GetByID(ctx context.Context, id uuid.UUID) (*Genre, error)

	// GetBySlug lấy genre theo slug
	GetBySlug(ctx context.Context, slug string) (*Genre, error)

	// GetAll lấy tất cả genres (có thể filter)
	GetAll(ctx context.Context, activeOnly bool) ([]*Genre, error)

	// List lấy danh sách genres với pagination, search và sort
	List(ctx context.Context, offset, limit int, search, sortBy, sortOrder string, activeOnly bool) ([]*Genre, int, error)

	// ListSelection lấy danh sách genres rút gọn (chỉ ID và Name)
	ListSelection(ctx context.Context, offset, limit int, search string) ([]*Genre, int, error)

	// GetByParentID lấy các genre con theo parent ID
	GetByParentID(ctx context.Context, parentID uuid.UUID) ([]*Genre, error)

	// GetRootGenres lấy các genre gốc (không có parent)
	GetRootGenres(ctx context.Context) ([]*Genre, error)

	// Create tạo genre mới
	Create(ctx context.Context, genre *Genre) error

	// Update cập nhật genre
	Update(ctx context.Context, genre *Genre) error

	// Delete xóa genre (soft delete)
	Delete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error

	// GetNovelGenres lấy danh sách genres của một novel
	GetNovelGenres(ctx context.Context, novelID uuid.UUID) ([]*Genre, error)

	// AddNovelGenre thêm genre cho novel
	AddNovelGenre(ctx context.Context, novelID, genreID, createdBy uuid.UUID) error
	
	// AddNovelGenres thêm nhiều genres cho novel (Batch Insert)
	AddNovelGenres(ctx context.Context, novelID uuid.UUID, genreIDs []uuid.UUID, createdBy uuid.UUID) error

	// RemoveNovelGenre xóa genre khỏi novel
	RemoveNovelGenre(ctx context.Context, novelID, genreID uuid.UUID) error

	// UpdateNovelGenres cập nhật toàn bộ genres của novel
	UpdateNovelGenres(ctx context.Context, novelID uuid.UUID, genreIDs []uuid.UUID, createdBy uuid.UUID) error

	// BatchIncrementNovelCount tăng số lượng novel cho nhiều genres
	BatchIncrementNovelCount(ctx context.Context, increments map[uuid.UUID]int) error

	// BatchIncrementTotalViews tăng total views cho nhiều genres
	BatchIncrementTotalViews(ctx context.Context, increments map[uuid.UUID]int64) error

	// GetGenresByNovelIDs lấy genre IDs cho danh sách novel IDs
	GetGenresByNovelIDs(ctx context.Context, novelIDs []uuid.UUID) (map[uuid.UUID][]uuid.UUID, error)

	// BatchUpdateActiveReaders cập nhật số lượng active readers cho nhiều genres
	BatchUpdateActiveReaders(ctx context.Context, updates map[uuid.UUID]int64) error

	// Merge gộp nhiều genres (sources) thành một genre (target)
	Merge(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID, mergedBy uuid.UUID) error

	// GetMergePreview lấy thông tin preview trước khi merge
	GetMergePreview(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID) ([]*AffectedNovel, error)
}

// AffectedNovel là thông tin novel tối thiểu dùng cho merge preview
type AffectedNovel struct {
	ID            uuid.UUID
	Title         string
	Slug          string
	CoverImageURL *string
}

// GenreTree là cấu trúc phân cấp của genres
type GenreTree struct {
	Genre    *Genre
	Children []*GenreTree
}
