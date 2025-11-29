package i18nkeys

// Common message keys
const (
	// General
	GeneralSuccess = "general.success"
	GeneralError   = "general.error"

	// Validation
	ValidationFailed        = "validation.failed"
	ValidationRequired      = "validation.required"
	ValidationInvalidFormat = "validation.invalid_format"

	// Auth
	AuthUnauthorized               = "auth.unauthorized"
	AuthForbidden                  = "auth.forbidden"
	AuthInvalidToken               = "auth.invalid_token"
	AuthMissingAuthorizationHeader = "auth.missing_authorization_header"
	AuthInvalidAuthorizationFormat = "auth.invalid_authorization_format"
	AuthInsufficientScope          = "auth.insufficient_scope"
	AuthMiddlewareError            = "auth.middleware_error"
	AuthInvalidUserID              = "auth.invalid_user_id"

	// Resource
	ResourceNotFound = "resource.not_found"
	ResourceConflict = "resource.conflict"

	// Rate limit
	RateLimitExceeded = "rate_limit.exceeded"
)
