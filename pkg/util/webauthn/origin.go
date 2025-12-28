package webauthn

import (
	"net/url"
)

// ExtractOrigin extracts the origin (scheme://host:port) from a given URL string.
// Returns empty string if URL is invalid.
//
// Example:
//
//	ExtractOrigin("http://localhost:3000/api/auth/callback") → "http://localhost:3000"
//	ExtractOrigin("https://example.com:8443/path") → "https://example.com:8443"
func ExtractOrigin(rawURL string) string {
	if rawURL == "" {
		return ""
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	// Origin = scheme + "://" + host
	// host already includes port if present
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return ""
	}

	return parsedURL.Scheme + "://" + parsedURL.Host
}
