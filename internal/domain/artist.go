package domain

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid/v5"
)

// Artist là domain model cho hoạ sĩ/minh họa
type Artist struct {
	ID     uuid.UUID
	UserID *uuid.UUID

	Name string
	Slug string

	Biography   json.RawMessage
	AvatarURL   *string
	SocialLinks json.RawMessage

	// Specialization: cover_artist, illustrator, manga_artist, etc.
	Specialization *string
	PortfolioURL   *string

	// Thống kê
	NovelCount    int
	ArtworkCount  int
	FollowerCount int

	IsVerified bool

	// Metadata (JSONB)
	Metadata json.RawMessage

	// Audit
	CreatedBy uuid.UUID
	UpdatedBy *uuid.UUID
	DeletedBy *uuid.UUID
	Version   int
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// ArtistRepository định nghĩa interface cho việc truy cập dữ liệu artist
type ArtistRepository interface {
	// GetByID lấy artist theo ID
	GetByID(ctx context.Context, id uuid.UUID) (*Artist, error)

	// GetBySlug lấy artist theo slug
	GetBySlug(ctx context.Context, slug string) (*Artist, error)

	// GetByUserID lấy artist theo user ID
	GetByUserID(ctx context.Context, userID uuid.UUID) (*Artist, error)

	// List lấy danh sách artists với filter
	List(ctx context.Context, filter ArtistFilter) ([]*Artist, int64, error)

	// ListSelection lấy danh sách artists rút gọn (chỉ ID và Name)
	ListSelection(ctx context.Context, offset, limit int, search string) ([]*Artist, int64, error)

	// Create tạo artist mới
	Create(ctx context.Context, artist *Artist) error

	// Update cập nhật artist
	Update(ctx context.Context, artist *Artist) error

	// Delete xóa mềm artist
	Delete(ctx context.Context, id uuid.UUID) error

	// GetNovelArtists lấy danh sách artists của một novel
	GetNovelArtists(ctx context.Context, novelID uuid.UUID) ([]*NovelArtist, error)

	// AddNovelArtist thêm artist cho novel
	AddNovelArtist(ctx context.Context, novelID, artistID uuid.UUID, displayOrder int) error

	// RemoveNovelArtist xóa artist khỏi novel
	RemoveNovelArtist(ctx context.Context, novelID, artistID uuid.UUID, role string) error

	// UpdateStatistics cập nhật thống kê
	UpdateStatistics(ctx context.Context, id uuid.UUID, stats ArtistStatistics) error

	// Merge gộp nhiều artists thành một
	Merge(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID, mergedBy uuid.UUID) error

	// GetMergePreview lấy danh sách các novel sẽ bị ảnh hưởng khi merge
	GetMergePreview(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID) ([]*Novel, error)
}

// ArtistSortField định nghĩa các field có thể sort cho artist list
type ArtistSortField string

const (
	ArtistSortByName    ArtistSortField = "name"
	ArtistSortByNovels  ArtistSortField = "novels"  // API value
	ArtistSortByCreated ArtistSortField = "created" // API value
)

// ToDBColumn chuyển đổi API sort field sang DB column name
func (s ArtistSortField) ToDBColumn() string {
	switch s {
	case ArtistSortByName:
		return "name"
	case ArtistSortByNovels:
		return "novel_count"
	case ArtistSortByCreated:
		return "created_at"
	default:
		return "created_at"
	}
}

// ArtistFilter định nghĩa các filter cho việc query artists
type ArtistFilter struct {
	SearchQuery    *string
	Specialization *string
	IsVerified     *bool
	SortBy         string // DB column name (đã được convert từ ArtistSortField)
	SortOrder      string // "asc", "desc"
	Limit          int
	Offset         int
}

// ArtistStatistics chứa thông tin thống kê để update
type ArtistStatistics struct {
	NovelCount    *int
	ArtworkCount  *int
	FollowerCount *int
}

// NovelArtist chứa thông tin về artist và role trong novel
type NovelArtist struct {
	Artist       *Artist
	Role         string // "cover_artist", "illustrator", "character_designer"
	DisplayOrder int
}
