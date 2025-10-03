package i18n

import (
	"strings"

	"github.com/gin-gonic/gin"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
)

const (
	localizerContextKey  = "localizer"
	languageContextKey   = "locale"
	translatorContextKey = "translator"
)

// SetLocalizer stores a request-scoped localizer inside the Gin context.
func SetLocalizer(c *gin.Context, localizer *goi18n.Localizer) {
	if localizer != nil {
		c.Set(localizerContextKey, localizer)
	}
}

// GetLocalizer retrieves the request localizer if present.
func GetLocalizer(c *gin.Context) *goi18n.Localizer {
	v, exists := c.Get(localizerContextKey)
	if !exists {
		return nil
	}
	localizer, ok := v.(*goi18n.Localizer)
	if !ok {
		return nil
	}
	return localizer
}

// SetLanguage persists the resolved language code for downstream consumers.
func SetLanguage(c *gin.Context, lang string) {
	if lang != "" {
		c.Set(languageContextKey, lang)
	}
}

// LanguageFromContext returns the resolved language code if available.
func LanguageFromContext(c *gin.Context) string {
	if v, exists := c.Get(languageContextKey); exists {
		if lang, ok := v.(string); ok {
			return lang
		}
	}
	return ""
}

// SetTranslator stores the translator instance in the context for middleware use
func SetTranslator(c *gin.Context, translator *Translator) {
	if translator != nil {
		c.Set(translatorContextKey, translator)
	}
}

// GetTranslator retrieves the translator instance from context
func GetTranslator(c *gin.Context) *Translator {
	v, exists := c.Get(translatorContextKey)
	if !exists {
		return nil
	}
	translator, ok := v.(*Translator)
	if !ok {
		return nil
	}
	return translator
}

// T localizes the specified message with namespace support, falling back to the provided default.
// Message IDs can be namespaced (e.g., "common.validation.required" or "identify.auth.login_failed")
func T(c *gin.Context, messageID string, defaultMessage string, templateData map[string]any) string {
	localizer := GetLocalizer(c)
	if localizer == nil {
		if defaultMessage != "" {
			return defaultMessage
		}
		return messageID
	}

	cfg := &goi18n.LocalizeConfig{MessageID: messageID}
	if defaultMessage != "" {
		cfg.DefaultMessage = &goi18n.Message{ID: messageID, Other: defaultMessage}
	}
	if len(templateData) > 0 {
		cfg.TemplateData = templateData
	}

	value, err := localizer.Localize(cfg)
	if err != nil {
		if defaultMessage != "" {
			return defaultMessage
		}
		return messageID
	}
	return value
}

// Localize is a convenience wrapper around T without template data.
func Localize(c *gin.Context, messageID, defaultMessage string) string {
	return T(c, messageID, defaultMessage, nil)
}

// LocalizeWithData mirrors T but keeps a clearer call-site name when passing template data.
func LocalizeWithData(c *gin.Context, messageID, defaultMessage string, templateData map[string]any) string {
	return T(c, messageID, defaultMessage, templateData)
}

// Tn localizes a pluralized message using the provided count with namespace support.
func Tn(c *gin.Context, messageID string, defaultMessage string, pluralCount any, templateData map[string]any) string {
	localizer := GetLocalizer(c)
	if localizer == nil {
		if defaultMessage != "" {
			return defaultMessage
		}
		return messageID
	}

	cfg := &goi18n.LocalizeConfig{MessageID: messageID, PluralCount: pluralCount}
	if defaultMessage != "" {
		cfg.DefaultMessage = &goi18n.Message{ID: messageID, Other: defaultMessage}
	}
	if len(templateData) > 0 {
		cfg.TemplateData = templateData
	}

	value, err := localizer.Localize(cfg)
	if err != nil {
		if defaultMessage != "" {
			return defaultMessage
		}
		return messageID
	}
	return value
}

// ExtractLanguagePreferences extracts language preferences from HTTP request in order of priority:
// 1. Query parameter (e.g., ?lang=en)
// 2. Cookie
// 3. Accept-Language header
func ExtractLanguagePreferences(c *gin.Context, config Config) []string {
	var languages []string

	// 1. Check query parameter
	if queryLang := c.Query(config.QueryParam); queryLang != "" {
		languages = append(languages, queryLang)
	}

	// 2. Check cookie
	if cookieLang, err := c.Cookie(config.CookieName); err == nil && cookieLang != "" {
		languages = append(languages, cookieLang)
	}

	// 3. Parse Accept-Language header
	if acceptLang := c.GetHeader(config.HeaderName); acceptLang != "" {
		headerLangs := parseAcceptLanguage(acceptLang)
		languages = append(languages, headerLangs...)
	}

	// 4. Fallback to default
	languages = append(languages, config.DefaultLanguage)

	return languages
}

// parseAcceptLanguage parses the Accept-Language header and returns languages in preference order
func parseAcceptLanguage(header string) []string {
	var languages []string

	// Use manual iteration to avoid modernize warning about ranging over strings.Split
	remaining := header
	for remaining != "" {
		var part string
		if idx := strings.Index(remaining, ","); idx >= 0 {
			part = remaining[:idx]
			remaining = remaining[idx+1:]
		} else {
			part = remaining
			remaining = ""
		}
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Remove quality factor if present (e.g., "en-US;q=0.9" -> "en-US")
		if idx := strings.Index(part, ";"); idx >= 0 {
			part = part[:idx]
		}

		part = strings.TrimSpace(part)
		if part != "" {
			languages = append(languages, part)
		}
	}

	return languages
}

// SetLanguageCookie sets the language preference cookie
func SetLanguageCookie(c *gin.Context, config Config, language string) {
	c.SetCookie(
		config.CookieName,
		language,
		60*60*24*30, // 30 days
		"/",
		"",
		false, // secure
		true,  // httpOnly
	)
}
