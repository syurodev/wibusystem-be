// Package services contains stateless helpers used by handlers, such as
// input validation and sanitization utilities for common fields.
package services

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// ValidationService handles input validation and sanitization
type ValidationService struct {
	validator *validator.Validate
}

// NewValidationService creates a new validation service
func NewValidationService() *ValidationService {
	validate := validator.New()
	
	// Register custom validators
	validate.RegisterValidation("strong_password", validateStrongPassword)
	validate.RegisterValidation("username", validateUsername)
	validate.RegisterValidation("tenant_name", validateTenantName)
	
	return &ValidationService{
		validator: validate,
	}
}

// ValidateStruct validates a struct using tags
func (vs *ValidationService) ValidateStruct(s interface{}) error {
	return vs.validator.Struct(s)
}

// SanitizeString removes potentially dangerous characters and trims whitespace
func (vs *ValidationService) SanitizeString(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")
	
	// Trim whitespace
	input = strings.TrimSpace(input)
	
	// Remove potentially dangerous characters (adjust as needed)
	dangerousChars := []string{
		"<script", "</script>", 
		"javascript:", "data:",
		"vbscript:", "onload=", "onerror=",
	}
	
	inputLower := strings.ToLower(input)
	for _, dangerous := range dangerousChars {
		if strings.Contains(inputLower, dangerous) {
			input = strings.ReplaceAll(input, dangerous, "")
		}
	}
	
	return input
}

// SanitizeEmail validates and normalizes email addresses
func (vs *ValidationService) SanitizeEmail(email string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	
	// Basic email validation regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return "", fmt.Errorf("invalid email format")
	}
	
	return email, nil
}

// ValidatePassword checks password strength
func (vs *ValidationService) ValidatePassword(password string) []string {
	var errors []string
	
	if len(password) < 8 {
		errors = append(errors, "password must be at least 8 characters long")
	}
	
	if len(password) > 128 {
		errors = append(errors, "password must be less than 128 characters")
	}
	
	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false
	
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	
	if !hasUpper {
		errors = append(errors, "password must contain at least one uppercase letter")
	}
	
	if !hasLower {
		errors = append(errors, "password must contain at least one lowercase letter")
	}
	
	if !hasNumber {
		errors = append(errors, "password must contain at least one number")
	}
	
	if !hasSpecial {
		errors = append(errors, "password must contain at least one special character")
	}
	
	// Check for common weak passwords
	weakPasswords := []string{
		"password", "12345678", "qwerty", "abc123",
		"password123", "admin", "letmein", "welcome",
	}
	
	for _, weak := range weakPasswords {
		if strings.ToLower(password) == weak {
			errors = append(errors, "password is too common and weak")
			break
		}
	}
	
	return errors
}

// ValidateUsername checks if username meets requirements
func (vs *ValidationService) ValidateUsername(username string) []string {
	var errors []string
	
	if len(username) < 3 {
		errors = append(errors, "username must be at least 3 characters long")
	}
	
	if len(username) > 30 {
		errors = append(errors, "username must be less than 30 characters")
	}
	
	// Username can only contain alphanumeric characters, underscores, and hyphens
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validUsername.MatchString(username) {
		errors = append(errors, "username can only contain letters, numbers, underscores, and hyphens")
	}
	
	// Username cannot start with a number
	if len(username) > 0 && unicode.IsNumber(rune(username[0])) {
		errors = append(errors, "username cannot start with a number")
	}
	
	// Reserved usernames
	reserved := []string{
		"admin", "administrator", "root", "system", "api",
		"www", "mail", "ftp", "support", "help", "about",
		"contact", "info", "news", "blog", "forum",
	}
	
	for _, res := range reserved {
		if strings.ToLower(username) == res {
			errors = append(errors, "username is reserved and cannot be used")
			break
		}
	}
	
	return errors
}

// Custom validators for struct tags

// validateStrongPassword is a custom validator for strong passwords
func validateStrongPassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	vs := NewValidationService()
	errors := vs.ValidatePassword(password)
	return len(errors) == 0
}

// validateUsername is a custom validator for usernames
func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	vs := NewValidationService()
	errors := vs.ValidateUsername(username)
	return len(errors) == 0
}

// validateTenantName is a custom validator for tenant names
func validateTenantName(fl validator.FieldLevel) bool {
	name := fl.Field().String()
	
	if len(name) < 2 || len(name) > 50 {
		return false
	}
	
	// Tenant name can contain letters, numbers, spaces, and some special characters
	validName := regexp.MustCompile(`^[a-zA-Z0-9\s\-_.,&()]+$`)
	return validName.MatchString(name)
}

// SecurityValidationRules contains validation rules for security
type SecurityValidationRules struct {
	MaxLoginAttempts    int
	LoginWindowMinutes  int
	PasswordMinLength   int
	PasswordMaxLength   int
	UsernameMinLength   int
	UsernameMaxLength   int
	TenantNameMinLength int
	TenantNameMaxLength int
}

// DefaultSecurityRules returns default security validation rules
func DefaultSecurityRules() SecurityValidationRules {
	return SecurityValidationRules{
		MaxLoginAttempts:    5,
		LoginWindowMinutes:  15,
		PasswordMinLength:   8,
		PasswordMaxLength:   128,
		UsernameMinLength:   3,
		UsernameMaxLength:   30,
		TenantNameMinLength: 2,
		TenantNameMaxLength: 50,
	}
}

// InputSanitizer provides methods to sanitize different types of input
type InputSanitizer struct{}

// NewInputSanitizer creates a new input sanitizer
func NewInputSanitizer() *InputSanitizer {
	return &InputSanitizer{}
}

// SanitizeHTML removes HTML tags and dangerous content
func (is *InputSanitizer) SanitizeHTML(input string) string {
	// Remove HTML tags
	htmlTagRegex := regexp.MustCompile(`<[^>]*>`)
	input = htmlTagRegex.ReplaceAllString(input, "")
	
	// Remove potentially dangerous content
	input = strings.ReplaceAll(input, "javascript:", "")
	input = strings.ReplaceAll(input, "data:", "")
	input = strings.ReplaceAll(input, "vbscript:", "")
	
	return strings.TrimSpace(input)
}

// SanitizeSQL removes SQL injection attempts
func (is *InputSanitizer) SanitizeSQL(input string) string {
	// Note: This is basic protection. Proper prepared statements should be used
	sqlKeywords := []string{
		"DROP", "DELETE", "INSERT", "UPDATE", "SELECT", "CREATE", "ALTER",
		"EXEC", "EXECUTE", "UNION", "SCRIPT", "DECLARE", "CAST", "CONVERT",
		"--", "/*", "*/", ";", "'", "\"",
	}
	
	for _, keyword := range sqlKeywords {
		input = strings.ReplaceAll(strings.ToUpper(input), keyword, "")
	}
	
	return input
}

// SanitizeFileName removes dangerous characters from file names
func (is *InputSanitizer) SanitizeFileName(filename string) string {
	// Remove path traversal attempts
	filename = strings.ReplaceAll(filename, "../", "")
	filename = strings.ReplaceAll(filename, "..\\", "")
	filename = strings.ReplaceAll(filename, "./", "")
	filename = strings.ReplaceAll(filename, ".\\", "")
	
	// Remove dangerous characters
	dangerousChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\x00"}
	for _, char := range dangerousChars {
		filename = strings.ReplaceAll(filename, char, "")
	}
	
	// Trim whitespace and dots
	filename = strings.Trim(filename, " .")
	
	return filename
}
