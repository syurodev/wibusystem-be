package interfaces

import (
	"context"

	"github.com/google/uuid"
	d "wibusystem/pkg/common/dto"
	m "wibusystem/pkg/common/model"
)

// GenreServiceInterface defines the contract for genre operations
type GenreServiceInterface interface {
	// CreateGenre creates a new genre with validation
	CreateGenre(ctx context.Context, req d.CreateGenreRequest) (*m.Genre, error)

	// GetGenreByID retrieves a genre by ID
	GetGenreByID(ctx context.Context, genreID uuid.UUID) (*m.Genre, error)

	// ListGenres retrieves paginated list of genres with optional search
	ListGenres(ctx context.Context, req d.ListGenresRequest) ([]*m.Genre, int64, error)

	// UpdateGenre updates genre information
	UpdateGenre(ctx context.Context, genreID uuid.UUID, req d.UpdateGenreRequest) (*m.Genre, error)

	// DeleteGenre removes a genre
	DeleteGenre(ctx context.Context, genreID uuid.UUID) error

	// ValidateGenreName validates genre name before creation/update
	ValidateGenreName(name string) error

	// CheckGenreExists checks if a genre name already exists
	CheckGenreExists(ctx context.Context, name string) (bool, error)
}