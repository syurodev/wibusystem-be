package i18n

import (
	"github.com/gin-gonic/gin"
)

// LocaleMiddleware creates a Gin middleware that automatically resolves language preferences
// and sets up the localizer for each request
func LocaleMiddleware(translator *Translator) gin.HandlerFunc {
	return func(c *gin.Context) {
		config := translator.GetConfig()

		// Extract language preferences from request
		languages := ExtractLanguagePreferences(c, config)

		// Find the best matching supported language
		var resolvedLanguage string
		for _, lang := range languages {
			if translator.IsLanguageSupported(lang) {
				resolvedLanguage = config.NormalizeLanguage(lang)
				break
			}
		}

		// Fallback to default if no supported language found
		if resolvedLanguage == "" {
			resolvedLanguage = config.DefaultLanguage
		}

		// Create localizer for this request
		localizer := translator.NewLocalizer(resolvedLanguage)

		// Store in context
		SetTranslator(c, translator)
		SetLocalizer(c, localizer)
		SetLanguage(c, resolvedLanguage)

		// Set language cookie if it's different from current
		if currentCookie, err := c.Cookie(config.CookieName); err != nil || currentCookie != resolvedLanguage {
			SetLanguageCookie(c, config, resolvedLanguage)
		}

		c.Next()
	}
}

// RequireLanguage creates a middleware that ensures a specific language is used for the request
func RequireLanguage(translator *Translator, requiredLanguage string) gin.HandlerFunc {
	return func(c *gin.Context) {
		config := translator.GetConfig()

		// Validate the required language
		if !translator.IsLanguageSupported(requiredLanguage) {
			requiredLanguage = config.DefaultLanguage
		}

		// Create localizer for the required language
		localizer := translator.NewLocalizer(requiredLanguage)

		// Store in context
		SetTranslator(c, translator)
		SetLocalizer(c, localizer)
		SetLanguage(c, requiredLanguage)

		c.Next()
	}
}
