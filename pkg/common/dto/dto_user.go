package dto

import (
	"encoding/json"
)

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Email         string           `json:"email" validate:"required,email"`
	Username      string           `json:"username,omitempty" validate:"max=100"`
	DisplayName   string           `json:"display_name,omitempty" validate:"max=100"`
	Password      string           `json:"password,omitempty" validate:"min=8"`
	AvatarURL     *string          `json:"avatar_url,omitempty" validate:"omitempty,url"`
	CoverImageURL *string          `json:"cover_image_url,omitempty" validate:"omitempty,url"`
	Bio           *json.RawMessage `json:"bio,omitempty"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Email         *string          `json:"email,omitempty" validate:"omitempty,email"`
	Username      *string          `json:"username,omitempty" validate:"omitempty,max=100"`
	DisplayName   *string          `json:"display_name,omitempty" validate:"omitempty,max=100"`
	AvatarURL     *string          `json:"avatar_url,omitempty" validate:"omitempty,url"`
	CoverImageURL *string          `json:"cover_image_url,omitempty" validate:"omitempty,url"`
	Bio           *json.RawMessage `json:"bio,omitempty"`
}

// UpdateUserStatusRequest represents the request to update user status (admin only)
type UpdateUserStatusRequest struct {
	IsBlocked *bool `json:"is_blocked,omitempty"`
}
