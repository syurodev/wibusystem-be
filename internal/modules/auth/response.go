package auth

import (
	dtoauth "system/internal/dto/auth"
)

// Re-export types from dto package for backward compatibility
type RegisterResponse = dtoauth.RegisterResponse
type UserProfileResponse = dtoauth.UserProfileResponse
