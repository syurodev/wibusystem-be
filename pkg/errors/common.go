package errors

import "errors"

// Common errors - Lỗi chung cho tất cả domains
var (
	// Resource errors
	ErrResourceNotFound = errors.New("resource not found")
	ErrDuplicateEntry   = errors.New("duplicate entry")
	ErrResourceConflict = errors.New("resource conflict")

	// Validation errors
	ErrInvalidInput     = errors.New("invalid input")
	ErrValidationFailed = errors.New("validation failed")
	ErrInvalidFormat    = errors.New("invalid format")

	// System errors
	ErrInternalServer     = errors.New("internal server error")
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrDatabaseError      = errors.New("database error")
	ErrRedisError         = errors.New("redis error")
)
