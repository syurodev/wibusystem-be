package author

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
)

// AuthorService interface định nghĩa business logic cho Author module
type AuthorService interface {
	CreateAuthor(ctx context.Context, name, biography string, avatarURL *string, socialLinksJSON *string, createdBy uuid.UUID) (*domain.Author, error)
	UpdateAuthor(ctx context.Context, id uuid.UUID, name, biography string, avatarURL *string, socialLinksJSON *string) (*domain.Author, error)
	DeleteAuthor(ctx context.Context, id uuid.UUID) error
	GetAuthorByID(ctx context.Context, id uuid.UUID) (*domain.Author, error)
	GetAuthorBySlug(ctx context.Context, slug string) (*domain.Author, error)
	ListAuthors(ctx context.Context, page, limit int, search, sortBy, sortOrder string, isVerified *bool) ([]*domain.Author, int, error)
	ListSelection(ctx context.Context, page, limit int, search string) ([]*domain.Author, int64, error)
	MergeAuthors(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID, mergedBy uuid.UUID) error
	PreviewMergeAuthors(ctx context.Context, targetID uuid.UUID, sourceIDs []uuid.UUID) ([]*domain.Novel, error)
	// Novel relation methods (used by CreateNovelUseCase)
	AddNovelAuthors(ctx context.Context, novelID uuid.UUID, authorIDs []uuid.UUID) error
}
