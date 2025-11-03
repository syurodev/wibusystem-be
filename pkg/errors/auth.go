package errors

import "errors"

// Authentication errors - Lỗi xác thực
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenAlreadyUsed   = errors.New("token has already been used")
	ErrMissingAuthHeader  = errors.New("missing authorization header")
	ErrInvalidAuthFormat  = errors.New("invalid authorization format")
)

// Registration errors - Lỗi đăng ký
var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrWeakPassword       = errors.New("password is too weak")
	ErrInvalidEmail       = errors.New("invalid email format")
)

// Authorization errors - Lỗi phân quyền
var (
	ErrForbidden              = errors.New("forbidden")
	ErrInsufficientScope      = errors.New("insufficient scope")
	ErrInsufficientPermission = errors.New("insufficient permission")
)

// Session errors - Lỗi session
var (
	ErrSessionExpired  = errors.New("session expired")
	ErrSessionNotFound = errors.New("session not found")
	ErrSessionInvalid  = errors.New("session invalid")
)

// Consent errors - Lỗi consent
var (
	ErrConsentNotFound = errors.New("consent not found")
	ErrConsentRevoked  = errors.New("consent has been revoked")
	ErrConsentDenied   = errors.New("user denied consent")
)
