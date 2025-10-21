// Package middleware contains HTTP middleware for the Identity module.
package middleware

import (
	"errors"
	"log"
	"net/http"

	"wibusystem/internal/modules/identity/dto"

	"github.com/gofiber/fiber/v2"
)

// Custom error types
var (
	// ErrUnauthorized is returned when authentication fails
	ErrUnauthorized = errors.New("unauthorized")
	// ErrForbidden is returned when user lacks permissions
	ErrForbidden = errors.New("forbidden")
	// ErrNotFound is returned when a resource is not found
	ErrNotFound = errors.New("not found")
	// ErrBadRequest is returned for invalid requests
	ErrBadRequest = errors.New("bad request")
	// ErrConflict is returned when there's a conflict (e.g., duplicate email)
	ErrConflict = errors.New("conflict")
	// ErrInternalServer is returned for internal server errors
	ErrInternalServer = errors.New("internal server error")
	// ErrValidation is returned for validation errors
	ErrValidation = errors.New("validation error")
)

// AppError represents an application-level error with context
type AppError struct {
	Err        error
	Message    string
	Code       string
	StatusCode int
	Details    map[string]any
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return e.Err.Error()
}

// NewAppError creates a new AppError
func NewAppError(err error, message string, statusCode int) *AppError {
	return &AppError{
		Err:        err,
		Message:    message,
		StatusCode: statusCode,
		Details:    make(map[string]any),
	}
}

// WithCode adds an error code to the AppError
func (e *AppError) WithCode(code string) *AppError {
	e.Code = code
	return e
}

// WithDetails adds details to the AppError
func (e *AppError) WithDetails(details map[string]any) *AppError {
	e.Details = details
	return e
}

// ErrorHandler is a middleware that handles errors in a standardized way
func ErrorHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		if err == nil {
			return nil
		}

		// Log the error
		log.Printf("[ERROR] %v", err)

		// Handle fiber errors
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			return c.Status(fiberErr.Code).JSON(dto.ErrorResponse{
				Error:   http.StatusText(fiberErr.Code),
				Message: fiberErr.Message,
			})
		}

		// Handle application errors
		var appErr *AppError
		if errors.As(err, &appErr) {
			return c.Status(appErr.StatusCode).JSON(dto.ErrorResponse{
				Error:   appErr.Err.Error(),
				Message: appErr.Message,
				Code:    appErr.Code,
				Details: appErr.Details,
			})
		}

		// Handle known error types
		statusCode := http.StatusInternalServerError
		message := "An unexpected error occurred"
		errorType := "internal_server_error"

		switch {
		case errors.Is(err, ErrUnauthorized):
			statusCode = http.StatusUnauthorized
			message = "Authentication required"
			errorType = "unauthorized"
		case errors.Is(err, ErrForbidden):
			statusCode = http.StatusForbidden
			message = "You don't have permission to access this resource"
			errorType = "forbidden"
		case errors.Is(err, ErrNotFound):
			statusCode = http.StatusNotFound
			message = "Resource not found"
			errorType = "not_found"
		case errors.Is(err, ErrBadRequest):
			statusCode = http.StatusBadRequest
			message = "Invalid request"
			errorType = "bad_request"
		case errors.Is(err, ErrConflict):
			statusCode = http.StatusConflict
			message = "Resource conflict"
			errorType = "conflict"
		case errors.Is(err, ErrValidation):
			statusCode = http.StatusUnprocessableEntity
			message = "Validation failed"
			errorType = "validation_error"
		}

		return c.Status(statusCode).JSON(dto.ErrorResponse{
			Error:   errorType,
			Message: message,
		})
	}
}

// RecoverMiddleware recovers from panics and returns a 500 error
func RecoverMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC] %v", r)
				c.Status(http.StatusInternalServerError).JSON(dto.ErrorResponse{
					Error:   "internal_server_error",
					Message: "An unexpected error occurred",
				})
			}
		}()
		return c.Next()
	}
}

// NotFoundHandler handles 404 errors
func NotFoundHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(http.StatusNotFound).JSON(dto.ErrorResponse{
			Error:   "not_found",
			Message: "The requested resource was not found",
		})
	}
}

// MethodNotAllowedHandler handles 405 errors
func MethodNotAllowedHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Status(http.StatusMethodNotAllowed).JSON(dto.ErrorResponse{
			Error:   "method_not_allowed",
			Message: "The requested method is not allowed for this resource",
		})
	}
}
