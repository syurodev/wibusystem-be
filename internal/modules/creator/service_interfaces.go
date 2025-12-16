package creator

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
)

// CreatorService interface định nghĩa business logic cho Creator module
type CreatorService interface {
	// ListCreators returns paginated list of creators with view stats and popular work info
	ListCreators(ctx context.Context, filter domain.CreatorListFilter) (*domain.CreatorListResult, error)
	// Novel count methods (used by CreateNovelUseCase)
	IncrementNovelCount(ctx context.Context, userID uuid.UUID) error
	DecrementNovelCount(ctx context.Context, userID uuid.UUID) error
}
