package novel

import (
	"context"
	"encoding/json"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
)

// NovelService interface định nghĩa business logic cho novels
type NovelService interface {
	// CreateNovel tạo novel mới
	CreateNovel(
		ctx context.Context,
		title string,
		synopsis json.RawMessage,
		coverImageURL, thumbnailURL *string,
		status, originalLanguage, originalTitle *string,
		metadataJSON *string,
		isOneshot bool,
		ownerID uuid.UUID,
		ownerType string,
		createdBy uuid.UUID,
		genreIDs []uuid.UUID,
		authorIDs []uuid.UUID,
		artistIDs []uuid.UUID,
	) (*domain.Novel, error)

	// UpdateNovel cập nhật thông tin novel
	UpdateNovel(ctx context.Context, id uuid.UUID, title string, synopsis json.RawMessage, coverImageURL, thumbnailURL *string, status, originalLanguage, originalTitle *string, metadataJSON *string, isOneshot bool) (*domain.Novel, error)

	// DeleteNovel xóa novel (soft delete)
	DeleteNovel(ctx context.Context, id uuid.UUID) error

	// GetNovelByID lấy thông tin novel theo ID
	GetNovelByID(ctx context.Context, id uuid.UUID) (*domain.Novel, error)

	// GetNovelBySlug lấy thông tin novel theo slug
	GetNovelBySlug(ctx context.Context, slug string) (*domain.Novel, error)

	// ListNovels lấy danh sách novels với pagination, search và sort
	ListNovels(ctx context.Context, page, limit int, ownerID *uuid.UUID, keySearch string, genreIDs []uuid.UUID, statusStrs []string, originalLanguage, sortBy, sortOrder string) ([]*domain.Novel, int, error)
	
	// GetNovelsByIDs lấy danh sách novels theo list IDs
	GetNovelsByIDs(ctx context.Context, ids []uuid.UUID) ([]*domain.Novel, error)

	// IncrementViewCount tăng view count của novel
	IncrementViewCount(ctx context.Context, id uuid.UUID) error

	// GetNovelGenres lấy danh sách genre IDs của novel
	GetNovelGenres(ctx context.Context, novelID uuid.UUID) ([]uuid.UUID, error)

	// GetNovelAuthors lấy danh sách author IDs của novel
	GetNovelAuthors(ctx context.Context, novelID uuid.UUID) ([]uuid.UUID, error)

	// GetNovelArtists lấy danh sách artist IDs của novel
	GetNovelArtists(ctx context.Context, novelID uuid.UUID) ([]uuid.UUID, error)

	// GetNovelGenresDetails lấy danh sách genre (chi tiết) của novel
	GetNovelGenresDetails(ctx context.Context, novelID uuid.UUID) ([]*domain.Genre, error)

	// GetNovelAuthorsDetails lấy danh sách author (chi tiết) của novel
	GetNovelAuthorsDetails(ctx context.Context, novelID uuid.UUID) ([]*domain.NovelAuthor, error)

	// GetNovelArtistsDetails lấy danh sách artist (chi tiết) của novel
	GetNovelArtistsDetails(ctx context.Context, novelID uuid.UUID) ([]*domain.NovelArtist, error)
}
