package interfaces

import (
	"context"

	"github.com/google/uuid"
	d "wibusystem/pkg/common/dto"
	m "wibusystem/pkg/common/model"
)

// CharacterServiceInterface defines the contract for character operations
type CharacterServiceInterface interface {
	// CreateCharacter creates a new character with validation
	CreateCharacter(ctx context.Context, req d.CreateCharacterRequest) (*m.Character, error)

	// GetCharacterByID retrieves a character by ID
	GetCharacterByID(ctx context.Context, characterID uuid.UUID) (*m.Character, error)

	// ListCharacters retrieves paginated list of characters with optional search
	ListCharacters(ctx context.Context, req d.ListCharactersRequest) ([]*m.Character, int64, error)

	// UpdateCharacter updates character information
	UpdateCharacter(ctx context.Context, characterID uuid.UUID, req d.UpdateCharacterRequest) (*m.Character, error)

	// DeleteCharacter removes a character
	DeleteCharacter(ctx context.Context, characterID uuid.UUID) error

	// ValidateCharacterName validates character name before creation/update
	ValidateCharacterName(name string) error

	// CheckCharacterExists checks if a character name already exists
	CheckCharacterExists(ctx context.Context, name string) (bool, error)
}