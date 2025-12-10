package stringutil

import (
	"fmt"

	"github.com/gosimple/slug"
)

// GenerateUniqueSlug tạo slug unique từ input string bằng cách thêm random suffix
// Input: "Sakuragi Sakura" -> Output: "sakuragi-sakura-a1b2c3d4"
func GenerateUniqueSlug(input string) (string, error) {
	return GenerateUniqueSlugWithLength(input, 8)
}

// GenerateUniqueSlugWithLength tạo slug unique với độ dài random suffix tùy chỉnh
// Input: ("Sakuragi Sakura", 6) -> Output: "sakuragi-sakura-a1b2c3"
func GenerateUniqueSlugWithLength(input string, suffixLength int) (string, error) {
	// Generate base slug from input
	baseSlug := slug.Make(input)
	if baseSlug == "" {
		baseSlug = "item" // Fallback nếu input không thể tạo slug
	}

	// Generate random suffix
	suffix, err := GenerateRandomAlphaNumeric(suffixLength)
	if err != nil {
		return "", fmt.Errorf("failed to generate random suffix: %w", err)
	}

	// Combine base slug with random suffix
	return fmt.Sprintf("%s-%s", baseSlug, suffix), nil
}

