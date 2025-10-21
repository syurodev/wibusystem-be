package service

import (
	"context"
	"fmt"
	"wibusystem/internal/modules/catalog/domain"
	"wibusystem/internal/modules/catalog/repository"

	"github.com/google/uuid"
)

// characterService implements the CharacterService interface.
type characterService struct {
	characterRepo repository.CharacterRepository
	novelRepo     repository.NovelRepository // To check if novel exists
}

// NewCharacterService creates a new instance of CharacterService.
func NewCharacterService(characterRepo repository.CharacterRepository, novelRepo repository.NovelRepository) CharacterService {
	return &characterService{characterRepo: characterRepo, novelRepo: novelRepo}
}

// CreateCharacter handles the business logic for creating a new character.
func (s *characterService) CreateCharacter(ctx context.Context, novelID uuid.UUID, name, description string) (*domain.Character, error) {
	// Check if the novel exists
	_, err := s.novelRepo.GetByID(ctx, novelID)
	if err != nil {
		return nil, fmt.Errorf("cannot create character for a non-existent novel: %w", err)
	}

	character, err := domain.NewCharacter(name, description, "", novelID)
	if err != nil {
		return nil, err
	}

	if err := s.characterRepo.Create(ctx, character); err != nil {
		return nil, err
	}

	return character, nil
}

// GetCharacter retrieves a single character.
func (s *characterService) GetCharacter(ctx context.Context, id uuid.UUID) (*domain.Character, error) {
	return s.characterRepo.GetByID(ctx, id)
}

// ListCharactersByNovel retrieves a list of characters for a specific novel.
func (s *characterService) ListCharactersByNovel(ctx context.Context, novelID uuid.UUID, page, pageSize int) ([]*domain.Character, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	return s.characterRepo.ListByNovelID(ctx, novelID, pageSize, offset)
}
