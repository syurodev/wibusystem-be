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

// GenreService implements genre-related business logic
type GenreService struct {
	repos *repositories.Repositories
}

// NewGenreService creates a new genre service
func NewGenreService(repos *repositories.Repositories) interfaces.GenreServiceInterface {
	return &GenreService{
		repos: repos,
	}
}

// CreateGenre creates a new genre with validation
func (s *GenreService) CreateGenre(ctx context.Context, req d.CreateGenreRequest) (*m.Genre, error) {
	if err := s.ValidateGenreName(req.Name); err != nil {
		return nil, err
	}

	// Check if genre name already exists
	exists, err := s.CheckGenreExists(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check genre existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("genre with name '%s' already exists", req.Name)
	}

	// Create genre
	genre := &m.Genre{
		Name: strings.TrimSpace(req.Name),
	}

	if err := s.repos.Genre.Create(ctx, genre); err != nil {
		return nil, fmt.Errorf("failed to create genre: %w", err)
	}

	return genre, nil
}

// GetGenreByID retrieves a genre by ID
func (s *GenreService) GetGenreByID(ctx context.Context, genreID uuid.UUID) (*m.Genre, error) {
	if genreID == uuid.Nil {
		return nil, fmt.Errorf("genre ID cannot be nil")
	}

	genre, err := s.repos.Genre.GetByID(ctx, genreID)
	if err != nil {
		return nil, fmt.Errorf("failed to get genre by ID: %w", err)
	}

	return genre, nil
}

// ListGenres retrieves paginated list of genres with optional search
func (s *GenreService) ListGenres(ctx context.Context, req d.ListGenresRequest) ([]*m.Genre, int64, error) {
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

	genres, total, err := s.repos.Genre.List(ctx, req.PageSize, offset, search)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list genres: %w", err)
	}

	return genres, total, nil
}

// UpdateGenre updates genre information
func (s *GenreService) UpdateGenre(ctx context.Context, genreID uuid.UUID, req d.UpdateGenreRequest) (*m.Genre, error) {
	if genreID == uuid.Nil {
		return nil, fmt.Errorf("genre ID cannot be nil")
	}

	// Get existing genre
	genre, err := s.repos.Genre.GetByID(ctx, genreID)
	if err != nil {
		return nil, fmt.Errorf("failed to get genre: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if err := s.ValidateGenreName(name); err != nil {
			return nil, err
		}

		// Check if new name already exists (if different from current)
		if strings.ToLower(name) != strings.ToLower(genre.Name) {
			exists, err := s.CheckGenreExists(ctx, name)
			if err != nil {
				return nil, fmt.Errorf("failed to check genre existence: %w", err)
			}
			if exists {
				return nil, fmt.Errorf("genre with name '%s' already exists", name)
			}
		}

		genre.Name = name
	}

	// Update genre
	if err := s.repos.Genre.Update(ctx, genre); err != nil {
		return nil, fmt.Errorf("failed to update genre: %w", err)
	}

	return genre, nil
}

// DeleteGenre removes a genre
func (s *GenreService) DeleteGenre(ctx context.Context, genreID uuid.UUID) error {
	if genreID == uuid.Nil {
		return fmt.Errorf("genre ID cannot be nil")
	}

	// Check if genre exists
	_, err := s.repos.Genre.GetByID(ctx, genreID)
	if err != nil {
		return fmt.Errorf("genre not found: %w", err)
	}

	// Delete genre
	if err := s.repos.Genre.Delete(ctx, genreID); err != nil {
		return fmt.Errorf("failed to delete genre: %w", err)
	}

	return nil
}

// ValidateGenreName validates genre name
func (s *GenreService) ValidateGenreName(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return fmt.Errorf("genre name is required")
	}

	if len(name) > 100 {
		return fmt.Errorf("genre name must not exceed 100 characters")
	}

	if len(name) < 2 {
		return fmt.Errorf("genre name must be at least 2 characters long")
	}

	// Check for valid characters (letters, numbers, spaces, hyphens, underscores)
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == ' ' || r == '-' || r == '_' || r == '&' || r == '/') {
			return fmt.Errorf("genre name contains invalid characters")
		}
	}

	return nil
}

// CheckGenreExists checks if a genre name already exists
func (s *GenreService) CheckGenreExists(ctx context.Context, name string) (bool, error) {
	_, err := s.repos.Genre.GetByName(ctx, name)
	if err != nil {
		// If error is "not found", genre doesn't exist
		if strings.Contains(err.Error(), "no rows") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}