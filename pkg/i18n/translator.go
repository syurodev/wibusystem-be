package i18n

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// Translator wraps the go-i18n bundle and exposes helpers for matching and
// instantiating request-scoped localizers with multi-service namespace support.
type Translator struct {
	bundle    *goi18n.Bundle
	config    Config
	fallback  language.Tag
	supported map[string]struct{}
}

// NewTranslator creates a new translator instance with the given configuration.
// It loads translations from both common and service-specific directories.
func NewTranslator(config Config) (*Translator, error) {
	// Parse default language
	fallback, err := language.Parse(config.DefaultLanguage)
	if err != nil {
		return nil, fmt.Errorf("invalid default language %s: %w", config.DefaultLanguage, err)
	}

	// Create bundle
	bundle := goi18n.NewBundle(fallback)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Build supported languages map
	supported := make(map[string]struct{})
	for _, lang := range config.SupportedLanguages {
		supported[strings.ToLower(lang)] = struct{}{}
	}

	translator := &Translator{
		bundle:    bundle,
		config:    config,
		fallback:  fallback,
		supported: supported,
	}

	// Load translation files
	if err := translator.loadTranslations(); err != nil {
		return nil, fmt.Errorf("failed to load translations: %w", err)
	}

	return translator, nil
}

// loadTranslations loads translation files from configured paths
func (t *Translator) loadTranslations() error {
	paths := t.config.GetLocalePaths()

	for _, localePath := range paths {
		if err := t.loadTranslationsFromPath(localePath); err != nil {
			// Log warning but don't fail if a path doesn't exist
			continue
		}
	}

	return nil
}

// loadTranslationsFromPath loads translations from a specific directory path
func (t *Translator) loadTranslationsFromPath(localePath string) error {
	if _, err := os.Stat(localePath); os.IsNotExist(err) {
		return fmt.Errorf("locale path does not exist: %s", localePath)
	}

	return filepath.WalkDir(localePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		// Extract language from filename (e.g., "en.json" -> "en")
		filename := filepath.Base(path)
		lang := strings.TrimSuffix(filename, ".json")

		// Skip if language is not supported
		if _, supported := t.supported[strings.ToLower(lang)]; !supported {
			return nil
		}

		// Load the message file
		_, err = t.bundle.LoadMessageFile(path)
		if err != nil {
			return fmt.Errorf("failed to load message file %s: %w", path, err)
		}

		return nil
	})
}

// NewLocalizer creates a request-scoped localizer for the given language preferences.
// It attempts to match against supported languages and falls back appropriately.
func (t *Translator) NewLocalizer(languages ...string) *goi18n.Localizer {
	var tags []language.Tag

	// Process each language preference
	for _, lang := range languages {
		if lang == "" {
			continue
		}

		normalizedLang := t.config.NormalizeLanguage(lang)
		if tag, err := language.Parse(normalizedLang); err == nil {
			tags = append(tags, tag)
		}
	}

	// Always include fallback language
	tags = append(tags, t.fallback)

	// Convert tags to strings
	langStrings := getLanguageStrings(tags)
	return goi18n.NewLocalizer(t.bundle, langStrings...)
}

// GetSupportedLanguages returns the list of supported languages
func (t *Translator) GetSupportedLanguages() []string {
	var languages []string
	for lang := range t.supported {
		languages = append(languages, lang)
	}
	sort.Strings(languages)
	return languages
}

// IsLanguageSupported checks if a language is supported
func (t *Translator) IsLanguageSupported(lang string) bool {
	// Use strings.Cut for better performance (Go 1.18+)
	normalizedLang, _, _ := strings.Cut(strings.ToLower(lang), "-")
	_, supported := t.supported[normalizedLang]
	return supported
}

// GetConfig returns the translator configuration
func (t *Translator) GetConfig() Config {
	return t.config
}

// getLanguageStrings converts language tags to strings
func getLanguageStrings(tags []language.Tag) []string {
	var strings []string
	for _, tag := range tags {
		strings = append(strings, tag.String())
	}
	return strings
}
