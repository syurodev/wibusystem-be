package middleware

import (
	"github.com/gin-gonic/gin"

	"wibusystem/pkg/i18n"
)

// LocaleMiddleware wraps the shared locale resolver for the Catalog service.
type LocaleMiddleware struct {
	middleware gin.HandlerFunc
}

// NewLocaleMiddleware returns a locale middleware instance using shared i18n package.
func NewLocaleMiddleware(translator *i18n.Translator) *LocaleMiddleware {
	return &LocaleMiddleware{middleware: i18n.LocaleMiddleware(translator)}
}

// Handler exposes the gin.HandlerFunc to plug into router.Use.
func (lm *LocaleMiddleware) Handler() gin.HandlerFunc {
	return lm.middleware
}
