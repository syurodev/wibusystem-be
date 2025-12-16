package novel

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
)

// UCNovelService interface cho novel service operations dùng bởi UseCases
type UCNovelService interface {
	CreateNovelEntity(ctx context.Context, novel *domain.Novel) error
	DeleteNovelEntity(ctx context.Context, id uuid.UUID) error
	GetNovelByID(ctx context.Context, id uuid.UUID) (*domain.Novel, error)
}

// UCGenreService interface cho genre operations từ novel use case
type UCGenreService interface {
	AddNovelGenres(ctx context.Context, novelID uuid.UUID, genreIDs []uuid.UUID, createdBy uuid.UUID) error
	RemoveNovelGenres(ctx context.Context, novelID uuid.UUID) error
	GetNovelGenres(ctx context.Context, novelID uuid.UUID) ([]*domain.Genre, error)
	BatchIncrementNovelCount(ctx context.Context, increments map[uuid.UUID]int) error
}

// UCAuthorService interface cho author operations từ novel use case
type UCAuthorService interface {
	AddNovelAuthors(ctx context.Context, novelID uuid.UUID, authorIDs []uuid.UUID) error
}

// UCArtistService interface cho artist operations từ novel use case
type UCArtistService interface {
	AddNovelArtists(ctx context.Context, novelID uuid.UUID, artistIDs []uuid.UUID) error
}

// UCCreatorService interface cho creator/user statistics operations
type UCCreatorService interface {
	IncrementNovelCount(ctx context.Context, userID uuid.UUID) error
	DecrementNovelCount(ctx context.Context, userID uuid.UUID) error
}
