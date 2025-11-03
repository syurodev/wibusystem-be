package errcode

// ErrorCode là kiểu định nghĩa cho error codes
type ErrorCode string

// String implements Stringer interface
func (e ErrorCode) String() string {
	return string(e)
}

// ============================================
// 4xxx: Client Errors
// ============================================

// 400x: Bad Request / Validation Errors
const (
	ValidationRequired      ErrorCode = "E4001" // Validation: field required
	ValidationInvalidFormat ErrorCode = "E4002" // Validation: invalid format
	ValidationInvalidUserID ErrorCode = "E4003" // Validation: invalid user ID
)

// 401x: Unauthorized / Authentication Errors
const (
	AuthInvalidToken               ErrorCode = "E4011" // Invalid or expired token
	AuthInvalidAuthorizationFormat ErrorCode = "E4012" // Invalid Authorization header format
	AuthInvalidCredentials         ErrorCode = "E4013" // Invalid username/password
	AuthMissingAuthHeader          ErrorCode = "E4014" // Missing Authorization header
)

// 403x: Forbidden / Authorization Errors
const (
	AuthInsufficientScope    ErrorCode = "E4031" // Insufficient scope
	AuthInsufficientAnyScope ErrorCode = "E4032" // No required scopes
	AuthForbidden            ErrorCode = "E4033" // General forbidden
)

// 404x: Not Found Errors
const (
	ResourceNotFound ErrorCode = "E4041" // Resource not found
	UserNotFound     ErrorCode = "E4042" // User not found
	ClientNotFound   ErrorCode = "E4043" // OAuth2 client not found
)

// 409x: Conflict Errors
const (
	ResourceConflict ErrorCode = "E4091" // Resource already exists
	DuplicateEntry   ErrorCode = "E4092" // Duplicate entry
)

// 429x: Rate Limit Errors
const (
	RateLimitExceeded ErrorCode = "E4291" // Too many requests
)

// ============================================
// 5xxx: Server Errors
// ============================================

// 500x: Internal Server Errors
const (
	InternalError        ErrorCode = "E5001" // General internal error
	MiddlewareError      ErrorCode = "E5002" // Middleware error
	ScopeValidationError ErrorCode = "E5003" // Scope validation error
	ContextDataError     ErrorCode = "E5004" // Context data error
	DatabaseError        ErrorCode = "E5010" // Database error
	RedisError           ErrorCode = "E5011" // Redis error
)

// 503x: Service Unavailable Errors
const (
	ServiceUnavailable  ErrorCode = "E5031" // Service temporarily unavailable
	DatabaseUnavailable ErrorCode = "E5032" // Database unavailable
)
