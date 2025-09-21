// Package i18n provides internationalization support for the wibusystem monorepo.
// It supports multiple services with namespaced translations and various locale resolution strategies.
package i18n

import (
	"path/filepath"
	"strings"
)

// Config describes how the translator bundle should be constructed for multi-service support.
type Config struct {
	// BundlePath is the root path containing locale directories
	BundlePath string `json:"bundle_path"`

	// DefaultLanguage is the fallback language when requested locale is not available
	DefaultLanguage string `json:"default_language"`

	// SupportedLanguages is the list of languages this service supports
	SupportedLanguages []string `json:"supported_languages"`

	// FallbackToDefault determines if we should fallback to default language for missing translations
	FallbackToDefault bool `json:"fallback_to_default"`

	// ServiceNamespace is the service-specific namespace for translations (e.g., "identify", "catalog")
	ServiceNamespace string `json:"service_namespace"`

	// LoadCommonTranslations determines if common translations should be loaded alongside service-specific ones
	LoadCommonTranslations bool `json:"load_common_translations"`

	// QueryParam is the query parameter name to extract language preference (e.g., "lang")
	QueryParam string `json:"query_param"`

	// HeaderName is the header name to extract language preference (e.g., "Accept-Language")
	HeaderName string `json:"header_name"`

	// CookieName is the cookie name to store/retrieve language preference
	CookieName string `json:"cookie_name"`
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() Config {
	return Config{
		BundlePath:             "locales",
		DefaultLanguage:        "en",
		SupportedLanguages:     []string{"en", "vi"},
		FallbackToDefault:      true,
		ServiceNamespace:       "",
		LoadCommonTranslations: true,
		QueryParam:             "lang",
		HeaderName:             "Accept-Language",
		CookieName:             "locale",
	}
}

// GetLocalePaths returns the paths where translation files should be loaded from
func (c Config) GetLocalePaths() []string {
	var paths []string

	// Always load common translations first (if enabled)
	if c.LoadCommonTranslations {
		paths = append(paths, filepath.Join(c.BundlePath, "common"))
	}

	// Add service-specific translations
	if c.ServiceNamespace != "" {
		paths = append(paths, filepath.Join(c.BundlePath, c.ServiceNamespace))
	}

	return paths
}

// ValidateLanguage checks if the given language is supported
func (c Config) ValidateLanguage(lang string) bool {
	if lang == "" {
		return false
	}

	// Normalize language code (handle cases like "en-US" -> "en")
	normalizedLang, _, _ := strings.Cut(strings.ToLower(lang), "-")

	for _, supported := range c.SupportedLanguages {
		if strings.ToLower(supported) == normalizedLang {
			return true
		}
	}
	return false
}

// NormalizeLanguage returns the normalized language code
func (c Config) NormalizeLanguage(lang string) string {
	if lang == "" {
		return c.DefaultLanguage
	}

	normalizedLang, _, _ := strings.Cut(strings.ToLower(lang), "-")

	if c.ValidateLanguage(normalizedLang) {
		return normalizedLang
	}

	if c.FallbackToDefault {
		return c.DefaultLanguage
	}

	return normalizedLang
}
