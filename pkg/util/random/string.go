package random

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
