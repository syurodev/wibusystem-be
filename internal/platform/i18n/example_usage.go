package i18n

// File này chứa các ví dụ sử dụng i18n trong application
// KHÔNG build file này - chỉ dùng để tham khảo

/*
import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"

	platformI18n "system/internal/platform/i18n"
	"system/pkg/util/response"
)

// ===============================================
// Example 1: Sử dụng i18n trong Gin Handler
// ===============================================

func ExampleHandler(c *gin.Context) {
	// Lấy localizer từ context (được set bởi i18n middleware)
	localizer := c.MustGet("localizer").(*i18n.Localizer)

	// Translate message đơn giản
	message := localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "general.success",
	})

	c.JSON(http.StatusOK, gin.H{
		"message": message, // "Operation completed successfully." hoặc "Thao tác được thực hiện thành công."
	})
}

// ===============================================
// Example 2: Translate message với template data
// ===============================================

func ExampleValidationError(c *gin.Context) {
	localizer := c.MustGet("localizer").(*i18n.Localizer)

	// Translate với template data
	message := localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "validation.required",
		TemplateData: map[string]interface{}{
			"field": "email",
		},
	})

	c.JSON(http.StatusBadRequest, gin.H{
		"error": message, // "The field 'email' is required." hoặc "Trường 'email' là bắt buộc."
	})
}

// ===============================================
// Example 3: Sử dụng với StandardResponse
// ===============================================

func ExampleStandardResponse(c *gin.Context) {
	localizer := c.MustGet("localizer").(*i18n.Localizer)

	// Success response
	response.Success(c, http.StatusOK, "user.created", nil, nil)

	// Error response với template data
	response.Error(c, http.StatusBadRequest, "validation.required", map[string]interface{}{
		"field": "password",
	})
}

// ===============================================
// Example 4: Tạo localizer từ Accept-Language header
// ===============================================

func ExampleGetLocalizerFromHeader(c *gin.Context) {
	// Lấy Accept-Language từ header
	acceptLang := c.GetHeader("Accept-Language") // Example: "vi-VN,vi;q=0.9,en;q=0.8"

	// Tạo localizer
	localizer := platformI18n.GetLocalizerFromAcceptLanguage(acceptLang)

	message := localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "auth.unauthorized",
	})

	c.JSON(http.StatusUnauthorized, gin.H{
		"error": message,
	})
}

// ===============================================
// Example 5: Tạo localizer từ language tag
// ===============================================

func ExampleGetLocalizerFromTag() {
	// Tạo localizer cho tiếng Việt
	localizerVi := platformI18n.GetLocalizer(language.Vietnamese)
	messageVi := localizerVi.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "general.success",
	})
	// Output: "Thao tác được thực hiện thành công."

	// Tạo localizer cho tiếng Anh
	localizerEn := platformI18n.GetLocalizer(language.English)
	messageEn := localizerEn.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "general.success",
	})
	// Output: "Operation completed successfully."
}

// ===============================================
// Example 6: Handle missing translation
// ===============================================

func ExampleHandleMissingTranslation(c *gin.Context) {
	localizer := c.MustGet("localizer").(*i18n.Localizer)

	// Localize với fallback
	message, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    "some.missing.key",
		DefaultMessage: &i18n.Message{
			ID:    "some.missing.key",
			Other: "Default message if translation not found",
		},
	})

	if err != nil {
		// Translation không tồn tại, sử dụng default message
		c.JSON(http.StatusOK, gin.H{
			"message": message, // "Default message if translation not found"
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": message,
	})
}

// ===============================================
// Example 7: Pluralization (nếu cần trong tương lai)
// ===============================================

// Trong file i18n JSON, định nghĩa như sau:
// {
//   "item.count": {
//     "one": "You have {{.Count}} item",
//     "other": "You have {{.Count}} items"
//   }
// }

func ExamplePluralization(c *gin.Context) {
	localizer := c.MustGet("localizer").(*i18n.Localizer)

	count := 5

	message := localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "item.count",
		PluralCount: count,
		TemplateData: map[string]interface{}{
			"Count": count,
		},
	})

	c.JSON(http.StatusOK, gin.H{
		"message": message, // "You have 5 items"
	})
}

// ===============================================
// Example 8: Sử dụng trong service layer
// ===============================================

type UserService struct {
	localizer *i18n.Localizer
}

func (s *UserService) CreateUser(email string) (string, error) {
	// Validate email
	if email == "" {
		// Return translated error message
		return "", fmt.Errorf(s.localizer.MustLocalize(&i18n.LocalizeConfig{
			MessageID: "validation.required",
			TemplateData: map[string]interface{}{
				"field": "email",
			},
		}))
	}

	// Create user...

	// Return success message
	return s.localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "user.created",
	}), nil
}

// ===============================================
// Example 9: Rate limit message với dynamic data
// ===============================================

func ExampleRateLimitMessage(c *gin.Context) {
	localizer := c.MustGet("localizer").(*i18n.Localizer)

	retryAfter := 60 // seconds

	message := localizer.MustLocalize(&i18n.LocalizeConfig{
		MessageID: "rate_limit.exceeded",
		TemplateData: map[string]interface{}{
			"seconds": retryAfter,
		},
	})

	c.JSON(http.StatusTooManyRequests, gin.H{
		"error":       message, // "Too many requests. Please try again in 60 seconds."
		"retry_after": retryAfter,
	})
}

// ===============================================
// NOTES
// ===============================================

// 1. Message IDs trong JSON files:
//    - Sử dụng dot notation: "domain.action" (e.g., "user.created", "validation.required")
//    - Consistent naming giúp dễ tìm kiếm và maintain

// 2. Template Data:
//    - Sử dụng {{.FieldName}} trong JSON
//    - Pass data qua TemplateData map

// 3. Fallback Language:
//    - Default language là English (set trong InitI18n)
//    - Nếu translation không tồn tại, sẽ fallback về English

// 4. Thêm domain mới:
//    - Tạo folder: internal/platform/i18n/locales/{domain}/
//    - Tạo files: en.json, vi.json
//    - Thêm domain vào slice trong i18n.go: domains := []string{"common", "new_domain"}

// 5. Best Practices:
//    - Luôn sử dụng i18n cho user-facing messages
//    - Internal error messages (logs) có thể giữ English
//    - Test với cả Accept-Language: en và vi
//    - Sync message IDs giữa en.json và vi.json
*/
