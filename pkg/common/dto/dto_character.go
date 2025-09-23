package dto

// CreateCharacterRequest represents the request to create a new character
type CreateCharacterRequest struct {
	Name        string  `json:"name" validate:"required,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	ImageURL    *string `json:"image_url,omitempty" validate:"omitempty,url"`
}

// UpdateCharacterRequest represents the request to update a character
type UpdateCharacterRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	ImageURL    *string `json:"image_url,omitempty" validate:"omitempty,url"`
}

// ListCharactersRequest represents the request to list characters with pagination
type ListCharactersRequest struct {
	Page     int    `form:"page,default=1" validate:"min=1"`
	PageSize int    `form:"page_size,default=20" validate:"min=1,max=100"`
	Search   string `form:"search,omitempty" validate:"max=100"`
}