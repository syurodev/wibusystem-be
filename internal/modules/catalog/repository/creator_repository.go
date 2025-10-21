package repository

import (
	"context"
	"wibusystem/internal/modules/catalog/domain"

	"github.com/google/uuid"
)

// CreatorRepository defines the interface for interacting with creator data.
type CreatorRepository interface {
	Create(ctx context.Context, creator *domain.Creator) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Creator, error)
	Update(ctx context.Context, creator *domain.Creator) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*domain.Creator, int64, error)
}
