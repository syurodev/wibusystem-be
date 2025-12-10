package locale

import (
	"embed"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	"system/pkg/util/i18nutil"
)

//go:embed *.json
var i18nFS embed.FS

// RegisterI18n registers genre module's i18n messages to the shared bundle.
func RegisterI18n(bundle *i18n.Bundle) error {
	return i18nutil.RegisterModuleI18n(bundle, i18nFS, "genre", i18nutil.DefaultLanguages())
}
