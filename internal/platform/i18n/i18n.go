package i18n

import (
	"embed"
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/zap"
	"golang.org/x/text/language"

	analytics_locale "system/internal/modules/analytics/locale"
	artist_locale "system/internal/modules/artist/locale"
	auth_locale "system/internal/modules/auth/locale"
	author_locale "system/internal/modules/author/locale"
	creator_locale "system/internal/modules/creator/locale"
	genre_locale "system/internal/modules/genre/locale"
	media_locale "system/internal/modules/media/locale"
	oauth2_locale "system/internal/modules/oauth2/locale"
	organization_locale "system/internal/modules/organization/locale"
	payment_locale "system/internal/modules/payment/locale"
	user_locale "system/internal/modules/user/locale"
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

	// Load centralized i18n files (shared across modules)
	// Modules moved to module-level embed: genre, artist, author, user(account), auth, oauth2, creator, analytics, media
	domains := []string{"common"}
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

	// Register module-level i18n
	if err := genre_locale.RegisterI18n(bundle); err != nil {
		log.Error("Failed to register genre i18n", zap.Error(err))
		return err
	}
	log.Info("Successfully loaded genre module i18n")

	if err := artist_locale.RegisterI18n(bundle); err != nil {
		log.Error("Failed to register artist i18n", zap.Error(err))
		return err
	}
	log.Info("Successfully loaded artist module i18n")

	if err := author_locale.RegisterI18n(bundle); err != nil {
		log.Error("Failed to register author i18n", zap.Error(err))
		return err
	}
	log.Info("Successfully loaded author module i18n")

	if err := user_locale.RegisterI18n(bundle); err != nil {
		log.Error("Failed to register user i18n", zap.Error(err))
		return err
	}
	log.Info("Successfully loaded user module i18n")

	if err := auth_locale.RegisterI18n(bundle); err != nil {
		log.Error("Failed to register auth i18n", zap.Error(err))
		return err
	}
	log.Info("Successfully loaded auth module i18n")

	if err := oauth2_locale.RegisterI18n(bundle); err != nil {
		log.Error("Failed to register oauth2 i18n", zap.Error(err))
		return err
	}
	log.Info("Successfully loaded oauth2 module i18n")

	if err := creator_locale.RegisterI18n(bundle); err != nil {
		log.Error("Failed to register creator i18n", zap.Error(err))
		return err
	}
	log.Info("Successfully loaded creator module i18n")

	if err := analytics_locale.RegisterI18n(bundle); err != nil {
		log.Error("Failed to register analytics i18n", zap.Error(err))
		return err
	}
	log.Info("Successfully loaded analytics module i18n")

	if err := media_locale.RegisterI18n(bundle); err != nil {
		log.Error("Failed to register media i18n", zap.Error(err))
		return err
	}
	log.Info("Successfully loaded media module i18n")

	if err := organization_locale.RegisterI18n(bundle); err != nil {
		log.Error("Failed to register organization i18n", zap.Error(err))
		return err
	}
	log.Info("Successfully loaded organization module i18n")

	if err := payment_locale.RegisterI18n(bundle); err != nil {
		log.Error("Failed to register payment i18n", zap.Error(err))
		return err
	}
	log.Info("Successfully loaded payment module i18n")

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
		// Ưu tiên lấy từ header x-locale-key
		lang := c.GetHeader("x-locale-key")
		if lang == "" {
			lang = c.GetHeader("Accept-Language")
		}
		
		localizer := i.GetLocalizerFromAcceptLanguage(lang)
		c.Set(LocalizerContextKey, &Localizer{localizer})
		c.Next()
	}
}
