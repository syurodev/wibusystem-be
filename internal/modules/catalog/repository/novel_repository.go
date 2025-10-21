package repository

import (
	"context"
	"wibusystem/internal/modules/catalog/domain"

	"github.com/google/uuid"
)

// NovelRepository defines the interface for interacting with novel data.
type NovelRepository interface {
	Create(ctx context.Context, novel *domain.Novel) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Novel, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Novel, error)
	Update(ctx context.Context, novel *domain.Novel) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*domain.Novel, int64, error)
	// Add other specific methods like search, etc.
}
