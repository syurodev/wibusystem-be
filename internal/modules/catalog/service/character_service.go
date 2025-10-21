package service

import (
	"context"
	"wibusystem/internal/modules/catalog/domain"

	"github.com/google/uuid"
)

// CharacterService defines the interface for character-related business logic.
type CharacterService interface {
	CreateCharacter(ctx context.Context, novelID uuid.UUID, name, description string) (*domain.Character, error)
	GetCharacter(ctx context.Context, id uuid.UUID) (*domain.Character, error)
	ListCharactersByNovel(ctx context.Context, novelID uuid.UUID, page, pageSize int) ([]*domain.Character, int64, error)
}
