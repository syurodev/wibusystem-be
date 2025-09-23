package dto

// CreateCreatorRequest represents the request to create a new creator
type CreateCreatorRequest struct {
	Name        string  `json:"name" validate:"required,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
}

// UpdateCreatorRequest represents the request to update a creator
type UpdateCreatorRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
}

// ListCreatorsRequest represents the request to list creators with pagination
type ListCreatorsRequest struct {
	Page     int    `form:"page,default=1" validate:"min=1"`
	PageSize int    `form:"page_size,default=20" validate:"min=1,max=100"`
	Search   string `form:"search,omitempty" validate:"max=100"`
}