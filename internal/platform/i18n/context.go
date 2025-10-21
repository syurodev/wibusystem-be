package i18n

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
)

const (
	localizerContextKey  = "localizer"
	languageContextKey   = "locale"
	translatorContextKey = "translator"
)

// SetLocalizer stores a request-scoped localizer inside the Fiber context.
func SetLocalizer(c *fiber.Ctx, localizer *goi18n.Localizer) {
	if localizer != nil {
		c.Locals(localizerContextKey, localizer)
	}
}

// GetLocalizer retrieves the request localizer if present.
func GetLocalizer(c *fiber.Ctx) *goi18n.Localizer {
	v := c.Locals(localizerContextKey)
	if v == nil {
		return nil
	}
	localizer, ok := v.(*goi18n.Localizer)
	if !ok {
		return nil
	}
	return localizer
}

// SetLanguage persists the resolved language code for downstream consumers.
func SetLanguage(c *fiber.Ctx, lang string) {
	if lang != "" {
		c.Locals(languageContextKey, lang)
	}
}

// LanguageFromContext returns the resolved language code if available.
func LanguageFromContext(c *fiber.Ctx) string {
	if v := c.Locals(languageContextKey); v != nil {
		if lang, ok := v.(string); ok {
			return lang
		}
	}
	return ""
}

// SetTranslator stores the translator instance in the context for middleware use
func SetTranslator(c *fiber.Ctx, translator *Translator) {
	if translator != nil {
		c.Locals(translatorContextKey, translator)
	}
}

// GetTranslator retrieves the translator instance from context
func GetTranslator(c *fiber.Ctx) *Translator {
	v := c.Locals(translatorContextKey)
	if v == nil {
		return nil
	}
	translator, ok := v.(*Translator)
	if !ok {
		return nil
	}
	return translator
}

// T localizes the specified message with namespace support, falling back to the provided default.
func T(c *fiber.Ctx, messageID string, defaultMessage string, templateData map[string]any) string {
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
func Localize(c *fiber.Ctx, messageID, defaultMessage string) string {
	return T(c, messageID, defaultMessage, nil)
}

// LocalizeWithData mirrors T but keeps a clearer call-site name when passing template data.
func LocalizeWithData(c *fiber.Ctx, messageID, defaultMessage string, templateData map[string]any) string {
	return T(c, messageID, defaultMessage, templateData)
}

// Tn localizes a pluralized message using the provided count with namespace support.
func Tn(c *fiber.Ctx, messageID string, defaultMessage string, pluralCount any, templateData map[string]any) string {
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
func ExtractLanguagePreferences(c *fiber.Ctx, config Config) []string {
	var languages []string

	// 1. Check query parameter
	if queryLang := c.Query(config.QueryParam); queryLang != "" {
		languages = append(languages, queryLang)
	}

	// 2. Check cookie
	if cookieLang := c.Cookies(config.CookieName); cookieLang != "" {
		languages = append(languages, cookieLang)
	}

	// 3. Parse Accept-Language header
	if acceptLang := c.Get(config.HeaderName); acceptLang != "" {
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
func SetLanguageCookie(c *fiber.Ctx, config Config, language string) {
	cookie := &fiber.Cookie{
		Name:     config.CookieName,
		Value:    language,
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		Path:     "/",
		HTTPOnly: true,
		Secure:   false, // Set to true in production
	}
	c.Cookie(cookie)
}
