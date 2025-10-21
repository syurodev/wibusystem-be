package repository

import (
	"context"
	"wibusystem/internal/modules/catalog/domain"

	"github.com/google/uuid"
)

// GenreRepository defines the interface for interacting with genre data.
type GenreRepository interface {
	Create(ctx context.Context, genre *domain.Genre) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Genre, error)
	Update(ctx context.Context, genre *domain.Genre) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*domain.Genre, int64, error)
}
