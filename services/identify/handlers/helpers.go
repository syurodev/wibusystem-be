package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"wibusystem/pkg/i18n"
	"wibusystem/services/identify/middleware"
)

// mapServiceError maps service errors to HTTP status codes and error messages
func mapServiceError(c *gin.Context, err error, context string) (status int, code string, message string, description string) {
	errStr := err.Error()

	// Check for common error patterns
	switch {
	case contains(errStr, "already exists") || contains(errStr, "already taken"):
		return http.StatusConflict, "resource_exists",
			i18n.Localize(c, "identify.errors.resource_exists.message", "Resource already exists"),
			i18n.Localize(c, "identify.errors.resource_exists.description", errStr)
	case contains(errStr, "invalid credentials") || contains(errStr, "incorrect"):
		return http.StatusUnauthorized, "invalid_credentials",
			i18n.Localize(c, "identify.errors.invalid_credentials.message", "Invalid credentials"),
			i18n.Localize(c, "identify.errors.invalid_credentials.description", "Invalid email or password")
	case contains(errStr, "not found"):
		return http.StatusNotFound, "not_found",
			i18n.Localize(c, "identify.errors.not_found.message", "Resource not found"),
			i18n.Localize(c, "identify.errors.not_found.description", errStr)
	case contains(errStr, "password") && contains(errStr, "must"):
		return http.StatusBadRequest, "password_validation_failed",
			i18n.Localize(c, "identify.errors.password_validation.message", "Password validation failed"),
			i18n.Localize(c, "identify.errors.password_validation.description", errStr)
	case contains(errStr, "required") || contains(errStr, "invalid") || contains(errStr, "must"):
		return http.StatusBadRequest, "validation_failed",
			i18n.Localize(c, "identify.errors.validation_failed.message", "Validation failed"),
			i18n.Localize(c, "identify.errors.validation_failed.description", errStr)
	default:
		return http.StatusInternalServerError, "internal_error",
			i18n.Localize(c, "identify.errors.internal_error.message", "Internal server error"),
			i18n.Localize(c, "identify.errors.internal_error.description", "An unexpected error occurred")
	}
}

// getCurrentUserID extracts the current authenticated user ID.
func getCurrentUserID(c *gin.Context) (uuid.UUID, error) {
	_, exists := middleware.GetUserFromContext(c)
	if !exists {
		return uuid.Nil, fmt.Errorf("user not authenticated")
	}

	// Parse subject as user ID
	// This is simplified - in production you'd have proper user ID extraction
	// For now return a placeholder UUID
	placeholderID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	return placeholderID, nil // Placeholder
}

// isAdmin checks if the current user has admin privileges via the "admin" scope.
func isAdmin(c *gin.Context) bool {
	userInfo, exists := middleware.GetUserFromContext(c)
	if !exists {
		return false
	}

	return userInfo.Scopes.Has("admin")
}

// contains checks if string contains substring
func contains(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
