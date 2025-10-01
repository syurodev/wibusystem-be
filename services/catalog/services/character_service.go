package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"

	d "wibusystem/pkg/common/dto"
	m "wibusystem/pkg/common/model"
	"wibusystem/services/catalog/repositories"
	"wibusystem/services/catalog/services/interfaces"
)

// CharacterService implements character-related business logic
type CharacterService struct {
	repos *repositories.Repositories
}

// NewCharacterService creates a new character service
func NewCharacterService(repos *repositories.Repositories) interfaces.CharacterServiceInterface {
	return &CharacterService{
		repos: repos,
	}
}

// CreateCharacter creates a new character with validation
func (s *CharacterService) CreateCharacter(ctx context.Context, req d.CreateCharacterRequest) (*m.Character, error) {
	if err := s.ValidateCharacterName(req.Name); err != nil {
		return nil, err
	}

	// Check if character name already exists
	exists, err := s.CheckCharacterExists(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check character existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("character with name '%s' already exists", req.Name)
	}

	// Create character
	character := &m.Character{
		Name:        strings.TrimSpace(req.Name),
		Description: req.Description,
		ImageURL:    req.ImageURL,
	}

	if err := s.repos.Character.Create(ctx, character); err != nil {
		return nil, fmt.Errorf("failed to create character: %w", err)
	}

	return character, nil
}

// GetCharacterByID retrieves a character by ID
func (s *CharacterService) GetCharacterByID(ctx context.Context, characterID uuid.UUID) (*m.Character, error) {
	if characterID == uuid.Nil {
		return nil, fmt.Errorf("character ID cannot be nil")
	}

	character, err := s.repos.Character.GetByID(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character by ID: %w", err)
	}

	return character, nil
}

// ListCharacters retrieves paginated list of characters with optional search
func (s *CharacterService) ListCharacters(ctx context.Context, req d.ListCharactersRequest) ([]*m.Character, int64, error) {
	// Validate pagination parameters
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	// Calculate offset
	offset := (req.Page - 1) * req.PageSize

	// Clean search term
	search := strings.TrimSpace(req.Search)

	characters, total, err := s.repos.Character.List(ctx, req.PageSize, offset, search)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list characters: %w", err)
	}

	return characters, total, nil
}

// UpdateCharacter updates character information
func (s *CharacterService) UpdateCharacter(ctx context.Context, characterID uuid.UUID, req d.UpdateCharacterRequest) (*m.Character, error) {
	if characterID == uuid.Nil {
		return nil, fmt.Errorf("character ID cannot be nil")
	}

	// Get existing character
	character, err := s.repos.Character.GetByID(ctx, characterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get character: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if err := s.ValidateCharacterName(name); err != nil {
			return nil, err
		}

		// Check if new name already exists (if different from current)
		if strings.ToLower(name) != strings.ToLower(character.Name) {
			exists, err := s.CheckCharacterExists(ctx, name)
			if err != nil {
				return nil, fmt.Errorf("failed to check character existence: %w", err)
			}
			if exists {
				return nil, fmt.Errorf("character with name '%s' already exists", name)
			}
		}

		character.Name = name
	}

	if req.Description != nil {
		character.Description = req.Description
	}

	if req.ImageURL != nil {
		character.ImageURL = req.ImageURL
	}

	// Update character
	if err := s.repos.Character.Update(ctx, character); err != nil {
		return nil, fmt.Errorf("failed to update character: %w", err)
	}

	return character, nil
}

// DeleteCharacter removes a character
func (s *CharacterService) DeleteCharacter(ctx context.Context, characterID uuid.UUID) error {
	if characterID == uuid.Nil {
		return fmt.Errorf("character ID cannot be nil")
	}

	// Check if character exists
	_, err := s.repos.Character.GetByID(ctx, characterID)
	if err != nil {
		return fmt.Errorf("character not found: %w", err)
	}

	// Delete character
	if err := s.repos.Character.Delete(ctx, characterID); err != nil {
		return fmt.Errorf("failed to delete character: %w", err)
	}

	return nil
}

// ValidateCharacterName validates character name
func (s *CharacterService) ValidateCharacterName(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return fmt.Errorf("character name is required")
	}

	if len(name) > 255 {
		return fmt.Errorf("character name must not exceed 255 characters")
	}

	if len(name) < 1 {
		return fmt.Errorf("character name must be at least 1 character long")
	}

	return nil
}

// CheckCharacterExists checks if a character name already exists
func (s *CharacterService) CheckCharacterExists(ctx context.Context, name string) (bool, error) {
	_, err := s.repos.Character.GetByName(ctx, name)
	if err != nil {
		// If error is "not found", character doesn't exist
		if strings.Contains(err.Error(), "no rows") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
