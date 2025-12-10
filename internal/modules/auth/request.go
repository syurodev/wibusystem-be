package auth

import (
	dtoauth "system/internal/dto/auth"
)

// Re-export request types from dto package
type RegisterRequest = dtoauth.RegisterRequest
type VerifyEmailRequest = dtoauth.VerifyEmailRequest
type ForgotPasswordRequest = dtoauth.ForgotPasswordRequest
type ResetPasswordRequest = dtoauth.ResetPasswordRequest
