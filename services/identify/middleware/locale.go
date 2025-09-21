package middleware

import (
	"github.com/gin-gonic/gin"

	"wibusystem/pkg/i18n"
)

// LocaleMiddleware resolves the preferred language for a request and exposes a
// go-i18n localizer via the Gin context.
type LocaleMiddleware struct {
	middleware gin.HandlerFunc
}

// NewLocaleMiddleware constructs a locale middleware instance using the new i18n package.
func NewLocaleMiddleware(translator *i18n.Translator, queryParam string) *LocaleMiddleware {
	// Use the built-in middleware from pkg/i18n
	middleware := i18n.LocaleMiddleware(translator)

	return &LocaleMiddleware{
		middleware: middleware,
	}
}

// Handler returns the Gin middleware that resolves language preferences.
func (lm *LocaleMiddleware) Handler() gin.HandlerFunc {
	return lm.middleware
}
