package creator

import (
	"context"

	"system/internal/domain"
)

// CreatorService interface định nghĩa business logic cho Creator module
type CreatorService interface {
	// ListCreators returns paginated list of creators with view stats and popular work info
	ListCreators(ctx context.Context, filter domain.CreatorListFilter) (*domain.CreatorListResult, error)
}
