package stringutil

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateRandomString tạo một chuỗi random an toàn với độ dài n bytes
// Sử dụng crypto/rand để đảm bảo tính ngẫu nhiên an toàn
func GenerateRandomString(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GenerateSessionID tạo một session ID an toàn (32 bytes)
func GenerateSessionID() (string, error) {
	return GenerateRandomString(32)
}

// GenerateRandomAlphaNumeric tạo chuỗi ngẫu nhiên chỉ gồm chữ và số
func GenerateRandomAlphaNumeric(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b), nil
}
