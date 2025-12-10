package i18nutil

import (
	"embed"
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// RegisterModuleI18n loads i18n files from an embedded filesystem into the bundle.
// Parameters:
//   - bundle: the shared i18n bundle to register messages to
//   - fs: embedded filesystem containing the JSON files
//   - moduleName: name of the module (used for unique file paths in bundle)
//   - languages: list of language codes to load (e.g., []string{"en", "vi"})
func RegisterModuleI18n(bundle *i18n.Bundle, fs embed.FS, moduleName string, languages []string) error {
	for _, lang := range languages {
		filePath := fmt.Sprintf("%s.json", lang)
		data, err := fs.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("%s: failed to read i18n file %s: %w", moduleName, filePath, err)
		}

		// Use a unique file path to avoid conflicts with other modules
		bundlePath := fmt.Sprintf("%s/%s.json", moduleName, lang)
		if _, err := bundle.ParseMessageFileBytes(data, bundlePath); err != nil {
			return fmt.Errorf("%s: failed to parse i18n file %s: %w", moduleName, filePath, err)
		}
	}

	return nil
}

// DefaultLanguages returns the default language list used in the project.
func DefaultLanguages() []string {
	return []string{"en", "vi"}
}
