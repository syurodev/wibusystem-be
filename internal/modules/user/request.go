package user

import (
	dtouser "system/internal/dto/user"
)

// Re-export request types from dto package
type UpdateProfileRequest = dtouser.UpdateProfileRequest
type ChangePasswordRequest = dtouser.ChangePasswordRequest
type UpdateSettingsRequest = dtouser.UpdateSettingsRequest
