package genre

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
)

// GenreService interface định nghĩa business logic cho Genre module
type GenreService interface {
	CreateGenre(ctx context.Context, name, description string, parentID *uuid.UUID, userID uuid.UUID) (*domain.Genre, error)
	UpdateGenre(ctx context.Context, id uuid.UUID, name, description string, parentID *uuid.UUID, isActive bool, userID uuid.UUID) (*domain.Genre, error)
	DeleteGenre(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	GetGenreByID(ctx context.Context, id uuid.UUID) (*domain.Genre, error)
	GetGenreBySlug(ctx context.Context, slug string) (*domain.Genre, error)
	ListGenres(ctx context.Context, page, limit int, search, sortBy, sortOrder string, activeOnly bool) ([]*GenreWithTrend, int, error)
	ListSelection(ctx context.Context, page, limit int, search string) ([]*domain.Genre, int, error)
	MergeGenres(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID, mergedBy uuid.UUID) error
	PreviewMergeGenres(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID) ([]*domain.AffectedNovel, error)
	// Novel relation methods (used by CreateNovelUseCase)
	AddNovelGenres(ctx context.Context, novelID uuid.UUID, genreIDs []uuid.UUID, createdBy uuid.UUID) error
	RemoveNovelGenres(ctx context.Context, novelID uuid.UUID) error
	GetNovelGenres(ctx context.Context, novelID uuid.UUID) ([]*domain.Genre, error)
	BatchIncrementNovelCount(ctx context.Context, increments map[uuid.UUID]int) error
}
