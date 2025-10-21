package service

import (
	"context"
	"wibusystem/internal/modules/catalog/domain"

	"github.com/google/uuid"
)

// GenreService defines the interface for genre-related business logic.
type GenreService interface {
	CreateGenre(ctx context.Context, name, description string) (*domain.Genre, error)
	GetGenre(ctx context.Context, id uuid.UUID) (*domain.Genre, error)
	ListGenres(ctx context.Context, page, pageSize int) ([]*domain.Genre, int64, error)
}
