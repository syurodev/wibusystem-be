package oauth

import "errors"

// Common OAuth validation errors
var (
	ErrInvalidToken     = errors.New("invalid or expired token")
	ErrInsufficientScope = errors.New("insufficient scope")
	ErrTokenRequired    = errors.New("access token required")
	ErrServerError      = errors.New("token validation server error")
	ErrUnauthorized     = errors.New("unauthorized")
)

// ErrorCode represents standard OAuth error codes
type ErrorCode string

const (
	ErrorCodeInvalidToken     ErrorCode = "invalid_token"
	ErrorCodeInsufficientScope ErrorCode = "insufficient_scope"
	ErrorCodeUnauthorized     ErrorCode = "unauthorized"
	ErrorCodeServerError      ErrorCode = "server_error"
)

// ValidationError represents a token validation error with code and description
type ValidationError struct {
	Code        ErrorCode `json:"error"`
	Description string    `json:"error_description,omitempty"`
}

func (e *ValidationError) Error() string {
	if e.Description != "" {
		return string(e.Code) + ": " + e.Description
	}
	return string(e.Code)
}

// NewValidationError creates a new validation error
func NewValidationError(code ErrorCode, description string) *ValidationError {
	return &ValidationError{
		Code:        code,
		Description: description,
	}
}