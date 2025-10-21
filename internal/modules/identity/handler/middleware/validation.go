// Package middleware contains HTTP middleware for the Identity module.
package middleware

import (
	"regexp"
	"strings"

	"wibusystem/internal/modules/identity/dto"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var (
	// validate is the global validator instance
	validate *validator.Validate

	// emailRegex is a simple email validation regex
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	// slugRegex validates slug format (lowercase alphanumeric with hyphens)
	slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
)

func init() {
	validate = validator.New()

	// Register custom validators
	_ = validate.RegisterValidation("slug", validateSlug)
	_ = validate.RegisterValidation("password_strength", validatePasswordStrength)
}

// ValidateRequest validates the request body against the provided struct
func ValidateRequest(c *fiber.Ctx, req any) error {
	// Parse the request body
	if err := c.BodyParser(req); err != nil {
		return NewAppError(
			ErrBadRequest,
			"Invalid request body format",
			fiber.StatusBadRequest,
		).WithCode("INVALID_JSON")
	}

	// Validate the request
	if err := validate.Struct(req); err != nil {
		validationErrors := formatValidationErrors(err)
		return c.Status(fiber.StatusUnprocessableEntity).JSON(dto.ValidationErrorResponse{
			Error:   "validation_error",
			Message: "Request validation failed",
			Errors:  validationErrors,
		})
	}

	return nil
}

// ValidateQuery validates query parameters against the provided struct
func ValidateQuery(c *fiber.Ctx, req any) error {
	// Parse query parameters
	if err := c.QueryParser(req); err != nil {
		return NewAppError(
			ErrBadRequest,
			"Invalid query parameters",
			fiber.StatusBadRequest,
		).WithCode("INVALID_QUERY")
	}

	// Validate the query
	if err := validate.Struct(req); err != nil {
		validationErrors := formatValidationErrors(err)
		return c.Status(fiber.StatusUnprocessableEntity).JSON(dto.ValidationErrorResponse{
			Error:   "validation_error",
			Message: "Query validation failed",
			Errors:  validationErrors,
		})
	}

	return nil
}

// formatValidationErrors formats validator errors into a user-friendly format
func formatValidationErrors(err error) []dto.ValidationError {
	var validationErrors []dto.ValidationError

	if validatorErrs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validatorErrs {
			validationErrors = append(validationErrors, dto.ValidationError{
				Field:   toSnakeCase(e.Field()),
				Message: getErrorMessage(e),
				Tag:     e.Tag(),
			})
		}
	}

	return validationErrors
}

// getErrorMessage returns a user-friendly error message for validation errors
func getErrorMessage(e validator.FieldError) string {
	field := toSnakeCase(e.Field())

	switch e.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "min":
		return field + " must be at least " + e.Param() + " characters long"
	case "max":
		return field + " must not exceed " + e.Param() + " characters"
	case "alphanum":
		return field + " must contain only alphanumeric characters"
	case "slug":
		return field + " must be a valid slug (lowercase letters, numbers, and hyphens)"
	case "oneof":
		return field + " must be one of: " + e.Param()
	case "uuid":
		return field + " must be a valid UUID"
	case "url":
		return field + " must be a valid URL"
	case "gte":
		return field + " must be greater than or equal to " + e.Param()
	case "lte":
		return field + " must be less than or equal to " + e.Param()
	case "password_strength":
		return field + " must contain at least one uppercase letter, one lowercase letter, one number, and one special character"
	default:
		return field + " is invalid"
	}
}

// Custom validators

// validateSlug validates that a string is a valid slug
func validateSlug(fl validator.FieldLevel) bool {
	slug := fl.Field().String()
	return slugRegex.MatchString(slug)
}

// validatePasswordStrength validates password strength
func validatePasswordStrength(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	// Require at least 3 out of 4 character types
	count := 0
	if hasUpper {
		count++
	}
	if hasLower {
		count++
	}
	if hasNumber {
		count++
	}
	if hasSpecial {
		count++
	}

	return count >= 3
}

// toSnakeCase converts a string from PascalCase/camelCase to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, char := range s {
		if char >= 'A' && char <= 'Z' {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(char + 32) // Convert to lowercase
		} else {
			result.WriteRune(char)
		}
	}
	return result.String()
}

// SanitizeInput sanitizes user input to prevent XSS
func SanitizeInput(input string) string {
	// Remove leading/trailing whitespace
	input = strings.TrimSpace(input)

	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	return input
}

// ValidateEmail validates an email address
func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// ValidatePassword validates password requirements
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return NewAppError(
			ErrValidation,
			"Password must be at least 8 characters long",
			fiber.StatusUnprocessableEntity,
		)
	}

	if len(password) > 72 {
		return NewAppError(
			ErrValidation,
			"Password must not exceed 72 characters",
			fiber.StatusUnprocessableEntity,
		)
	}

	return nil
}
