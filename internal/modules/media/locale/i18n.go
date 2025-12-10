package locale

import (
	"embed"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

//go:embed *.json
var localeFS embed.FS

// RegisterI18n registers the media module's i18n messages
func RegisterI18n(bundle *i18n.Bundle) error {
	files := []string{"en.json", "vi.json"}
	for _, file := range files {
		if _, err := bundle.LoadMessageFileFS(localeFS, file); err != nil {
			return err
		}
	}
	return nil
}
