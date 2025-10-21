package domain

import (
	"regexp"
	"strings"
)

var (
	nonAlphaNumRegex = regexp.MustCompile(`[^a-z0-9]+`)
	dashesRegex      = regexp.MustCompile(`-{2,}`)
)

func GenerateSlug(s string) (string, error) {
	if s == "" {
		return "", ErrInvalidTitleForSlug
	}

	slug := strings.ToLower(s)
	slug = nonAlphaNumRegex.ReplaceAllString(slug, "-")
	slug = dashesRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	if slug == "" {
		return "", ErrInvalidTitleForSlug
	}

	return slug, nil
}
