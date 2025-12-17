package artist

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
)

// ArtistService interface định nghĩa business logic cho Artist module
type ArtistService interface {
	CreateArtist(ctx context.Context, name, biography string, avatarURL *string, socialLinksJSON *string, specialization *string, portfolioURL *string, createdBy uuid.UUID) (*domain.Artist, error)
	UpdateArtist(ctx context.Context, id uuid.UUID, name, biography string, avatarURL *string, socialLinksJSON *string, specialization *string, portfolioURL *string) (*domain.Artist, error)
	DeleteArtist(ctx context.Context, id uuid.UUID) error
	GetArtistByID(ctx context.Context, id uuid.UUID) (*domain.Artist, error)
	GetArtistBySlug(ctx context.Context, slug string) (*domain.Artist, error)
	ListArtists(ctx context.Context, page, limit int, search, sortBy, sortOrder string, specialization *string, isVerified *bool) ([]*domain.Artist, int, error)
	ListSelection(ctx context.Context, page, limit int, search string) ([]*domain.Artist, int64, error)
	MergeArtists(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID, mergedBy uuid.UUID) error
	PreviewMergeArtists(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID) ([]*domain.Novel, error)
	// Novel relation methods (used by CreateNovelUseCase)
	AddNovelArtists(ctx context.Context, novelID uuid.UUID, artistIDs []uuid.UUID) error
}
