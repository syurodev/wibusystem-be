package repository

import (
	"context"
	"wibusystem/internal/modules/catalog/domain"

	"github.com/google/uuid"
)

// CharacterRepository defines the interface for interacting with character data.
type CharacterRepository interface {
	Create(ctx context.Context, character *domain.Character) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Character, error)
	Update(ctx context.Context, character *domain.Character) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByNovelID(ctx context.Context, novelID uuid.UUID, limit, offset int) ([]*domain.Character, int64, error)
}
