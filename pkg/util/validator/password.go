package validator

import pkgerrors "system/pkg/errors"

// I18n keys for validator
const (
	I18nWeakPassword = "validation.weak_password"
)

// ValidatePasswordStrength kiểm tra độ mạnh của password.
// Password phải có ít nhất 8 ký tự và chứa ít nhất 2 trong 3 loại ký tự:
// - Chữ hoa (A-Z)
// - Chữ thường (a-z)
// - Số (0-9)
func ValidatePasswordStrength(password string) error {
	if len(password) < 8 {
		return pkgerrors.BadRequest(I18nWeakPassword, "password must be at least 8 characters")
	}

	hasUpper := false
	hasLower := false
	hasDigit := false

	for _, char := range password {
		switch {
		case 'A' <= char && char <= 'Z':
			hasUpper = true
		case 'a' <= char && char <= 'z':
			hasLower = true
		case '0' <= char && char <= '9':
			hasDigit = true
		}
	}

	// Require at least 2 of 3 conditions (more flexible)
	conditions := 0
	if hasUpper {
		conditions++
	}
	if hasLower {
		conditions++
	}
	if hasDigit {
		conditions++
	}

	if conditions < 2 {
		return pkgerrors.BadRequest(I18nWeakPassword, "password must contain at least 2 of: uppercase, lowercase, digit")
	}

	return nil
}
