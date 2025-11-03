package crypto

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost là cost factor mặc định cho bcrypt
	// Cost 12 cân bằng giữa security và performance
	DefaultCost = 12
)

// HashPassword tạo bcrypt hash từ plain password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// VerifyPassword so sánh plain password với hashed password
// Trả về true nếu password đúng, false nếu sai
func VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
