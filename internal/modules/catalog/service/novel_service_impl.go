package service

import (
	"context"
	"wibusystem/internal/modules/catalog/domain"
	"wibusystem/internal/modules/catalog/repository"

	"github.com/google/uuid"
)

// novelService implements the NovelService interface.
type novelService struct {
	novelRepo repository.NovelRepository
}

// NewNovelService creates a new instance of NovelService.
func NewNovelService(novelRepo repository.NovelRepository) NovelService {
	return &novelService{novelRepo: novelRepo}
}

// CreateNovel handles the business logic for creating a new novel.
func (s *novelService) CreateNovel(ctx context.Context, title, summary string, authorID uuid.UUID) (*domain.Novel, error) {
	slug, err := domain.GenerateSlug(title)
	if err != nil {
		return nil, err
	}

	novel, err := domain.NewNovel(title, slug, domain.OwnershipPersonal, authorID, authorID, "en") // Simplified for now
	if err != nil {
		return nil, err
	}

	// The summary is a json.RawMessage, so we need to marshal it.
	if summary != "" {
		novel.Summary = []byte(`"` + summary + `"`)
	}

	if err := s.novelRepo.Create(ctx, novel); err != nil {
		return nil, err
	}

	return novel, nil
}

// GetNovel retrieves a single novel.
func (s *novelService) GetNovel(ctx context.Context, id uuid.UUID) (*domain.Novel, error) {
	return s.novelRepo.GetByID(ctx, id)
}

// UpdateNovel handles the logic for updating a novel.
func (s *novelService) UpdateNovel(ctx context.Context, id uuid.UUID, title, summary string) (*domain.Novel, error) {
	novel, err := s.novelRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := novel.UpdateTitle(title); err != nil {
		return nil, err
	}
	// In a real app, you'd update more fields here.

	if err := s.novelRepo.Update(ctx, novel); err != nil {
		return nil, err
	}

	return novel, nil
}

// DeleteNovel handles the logic for deleting a novel.
func (s *novelService) DeleteNovel(ctx context.Context, id uuid.UUID) error {
	// Here you would add business logic, e.g., checking if the novel can be deleted.
	return s.novelRepo.Delete(ctx, id)
}

// ListNovels retrieves a list of novels.
func (s *novelService) ListNovels(ctx context.Context, page, pageSize int) ([]*domain.Novel, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	return s.novelRepo.List(ctx, pageSize, offset)
}
