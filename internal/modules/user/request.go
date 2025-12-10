package user

// UpdateProfileRequest defines the request body for updating user profile
type UpdateProfileRequest struct {
	FullName    *string `json:"full_name"`
	DisplayName *string `json:"display_name"`
	Username    *string `json:"username"`
	Bio         any     `json:"bio"` // PlateJS content (JSON)
	AvatarURL   *string `json:"avatar_url"`
}

// ChangePasswordRequest defines the request body for changing password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// UpdateSettingsRequest defines the request body for updating settings
type UpdateSettingsRequest map[string]any
