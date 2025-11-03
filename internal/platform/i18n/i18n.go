package i18n

import (
	"embed"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/zap"
	"golang.org/x/text/language"
)

// Dùng 'embed' để nhúng các file i18n vào binary
// Lưu ý: Đường dẫn relative to file này (internal/platform/i18n/)
// Cấu trúc: locales/{domain}/{lang}.json
//
//go:embed locales/*/*.json
var fs embed.FS

var (
	instance *I18n
)

// I18n là struct chính quản lý bundle và các localizer.
type I18n struct {
	bundle *i18n.Bundle
	log    *zap.Logger
}

// InitI18n khởi tạo singleton instance và tải tất cả các file bản dịch.
func InitI18n(log *zap.Logger) error {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	instance = &I18n{
		bundle: bundle,
		log:    log,
	}

	domains := []string{"common", "oauth2", "auth"}
	languages := []string{"en", "vi"}

	for _, domain := range domains {
		for _, lang := range languages {
			filePath := fmt.Sprintf("locales/%s/%s.json", domain, lang)
			data, err := fs.ReadFile(filePath)
			if err != nil {
				log.Error("Failed to read i18n file",
					zap.String("path", filePath),
					zap.Error(err))
				return fmt.Errorf("failed to read i18n file %s: %w", filePath, err)
			}

			if _, err := bundle.ParseMessageFileBytes(data, filePath); err != nil {
				log.Error("Failed to parse i18n file",
					zap.String("path", filePath),
					zap.Error(err))
				return fmt.Errorf("failed to parse i18n file %s: %w", filePath, err)
			}

			log.Info("Successfully loaded i18n message file", zap.String("path", filePath))
		}
	}

	return nil
}

// GetInstance trả về singleton instance của I18n.
func GetInstance() *I18n {
	return instance
}

// GetLocalizer lấy Localizer từ Bundle dựa trên tag ngôn ngữ.
func (i *I18n) GetLocalizer(tag language.Tag) *i18n.Localizer {
	return i18n.NewLocalizer(i.bundle, tag.String())
}

// GetLocalizerFromAcceptLanguage lấy Localizer dựa trên header Accept-Language.
func (i *I18n) GetLocalizerFromAcceptLanguage(acceptLanguage string) *i18n.Localizer {
	return i18n.NewLocalizer(i.bundle, acceptLanguage)
}

// Localizer là một wrapper quanh i18n.Localizer để cung cấp các hàm tiện ích.
type Localizer struct {
	*i18n.Localizer
}

// LocalizerContextKey là key để lưu trữ localizer trong Gin context.
const LocalizerContextKey = "localizer"

// GinI18n là một middleware của Gin để inject localizer vào context.
func GinI18n(i *I18n) gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetHeader("Accept-Language")
		localizer := i.GetLocalizerFromAcceptLanguage(lang)
		c.Set(LocalizerContextKey, &Localizer{localizer})
		c.Next()
	}
}
