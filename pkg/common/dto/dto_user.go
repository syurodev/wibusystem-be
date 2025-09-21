package dto

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Email       string `json:"email" validate:"required,email"`
	Username    string `json:"username,omitempty" validate:"max=100"`
	DisplayName string `json:"display_name,omitempty" validate:"max=100"`
	Password    string `json:"password,omitempty" validate:"min=8"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Username    *string `json:"username,omitempty" validate:"omitempty,max=100"`
	DisplayName *string `json:"display_name,omitempty" validate:"omitempty,max=100"`
}
