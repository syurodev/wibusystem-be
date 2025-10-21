package service

import (
	"context"
	"wibusystem/internal/modules/catalog/domain"
	"wibusystem/internal/modules/catalog/repository"

	"github.com/google/uuid"
)

// genreService implements the GenreService interface.
type genreService struct {
	genreRepo repository.GenreRepository
}

// NewGenreService creates a new instance of GenreService.
func NewGenreService(genreRepo repository.GenreRepository) GenreService {
	return &genreService{genreRepo: genreRepo}
}

// CreateGenre handles the business logic for creating a new genre.
func (s *genreService) CreateGenre(ctx context.Context, name, description string) (*domain.Genre, error) {
	genre, err := domain.NewGenre(name, description)
	if err != nil {
		return nil, err
	}

	if err := s.genreRepo.Create(ctx, genre); err != nil {
		return nil, err
	}

	return genre, nil
}

// GetGenre retrieves a single genre.
func (s *genreService) GetGenre(ctx context.Context, id uuid.UUID) (*domain.Genre, error) {
	return s.genreRepo.GetByID(ctx, id)
}

// ListGenres retrieves a list of genres.
func (s *genreService) ListGenres(ctx context.Context, page, pageSize int) ([]*domain.Genre, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	return s.genreRepo.List(ctx, pageSize, offset)
}
