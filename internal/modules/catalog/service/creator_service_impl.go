package service

import (
	"context"
	"wibusystem/internal/modules/catalog/domain"
	"wibusystem/internal/modules/catalog/repository"

	"github.com/google/uuid"
)

// creatorService implements the CreatorService interface.
type creatorService struct {
	creatorRepo repository.CreatorRepository
}

// NewCreatorService creates a new instance of CreatorService.
func NewCreatorService(creatorRepo repository.CreatorRepository) CreatorService {
	return &creatorService{creatorRepo: creatorRepo}
}

// CreateCreator handles the business logic for creating a new creator.
func (s *creatorService) CreateCreator(ctx context.Context, name, bio string) (*domain.Creator, error) {
	creator, err := domain.NewCreator(name, bio, "") // Profile image is handled separately
	if err != nil {
		return nil, err
	}

	if err := s.creatorRepo.Create(ctx, creator); err != nil {
		return nil, err
	}

	return creator, nil
}

// GetCreator retrieves a single creator.
func (s *creatorService) GetCreator(ctx context.Context, id uuid.UUID) (*domain.Creator, error) {
	return s.creatorRepo.GetByID(ctx, id)
}

// ListCreators retrieves a list of creators.
func (s *creatorService) ListCreators(ctx context.Context, page, pageSize int) ([]*domain.Creator, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	return s.creatorRepo.List(ctx, pageSize, offset)
}
