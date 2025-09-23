package dto

// CreateGenreRequest represents the request to create a new genre
type CreateGenreRequest struct {
	Name string `json:"name" validate:"required,max=100"`
}

// UpdateGenreRequest represents the request to update a genre
type UpdateGenreRequest struct {
	Name *string `json:"name,omitempty" validate:"omitempty,max=100"`
}

// ListGenresRequest represents the request to list genres with pagination
type ListGenresRequest struct {
	Page     int    `form:"page,default=1" validate:"min=1"`
	PageSize int    `form:"page_size,default=20" validate:"min=1,max=100"`
	Search   string `form:"search,omitempty" validate:"max=100"`
}