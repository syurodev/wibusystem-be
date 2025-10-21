package dto

// CreateGenreRequest defines the structure for creating a new genre.
type CreateGenreRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description"`
}
