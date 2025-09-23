package auth

import "errors"

// Authentication errors
var (
	// ErrMissingToken is returned when no authentication token is provided
	ErrMissingToken = errors.New("missing authentication token")

	// ErrInvalidTokenFormat is returned when the token format is invalid
	ErrInvalidTokenFormat = errors.New("invalid token format")

	// ErrTokenExpired is returned when the token has expired
	ErrTokenExpired = errors.New("token has expired")

	// ErrInvalidToken is returned when the token is invalid
	ErrInvalidToken = errors.New("invalid token")

	// ErrInsufficientPermissions is returned when user lacks required permissions
	ErrInsufficientPermissions = errors.New("insufficient permissions")

	// ErrServiceUnavailable is returned when the identify service is unavailable
	ErrServiceUnavailable = errors.New("identity service unavailable")
)