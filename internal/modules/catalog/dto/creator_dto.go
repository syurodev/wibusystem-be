package dto

// CreateCreatorRequest defines the structure for creating a new creator.
type CreateCreatorRequest struct {
	Name string `json:"name" validate:"required,min=1,max=150"`
	Bio  string `json:"bio"`
}
