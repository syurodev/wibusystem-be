package response

import (
	"system/internal/platform/i18n"

	"github.com/gin-gonic/gin"
	go_i18n "github.com/nicksnyder/go-i18n/v2/i18n"
)

// StandardResponse là cấu trúc phản hồi chuẩn cho mọi API endpoint.
type StandardResponse struct {
	Success bool            `json:"success"`
	Code    string          `json:"code,omitempty"`
	Message string          `json:"message"` // Thông điệp đã được dịch
	Data    any             `json:"data,omitempty"`
	Meta    *PaginationMeta `json:"meta,omitempty"`
}

// PaginationMeta chỉ chứa thông tin phân trang
type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

// ----------------------------------------------------
// INTERNAL HELPER: Dịch tin nhắn
// ----------------------------------------------------

// getTranslatedMessage truy xuất Localizer từ Gin Context và dịch MessageID.
func getTranslatedMessage(c *gin.Context, msgID string, templateData any) string {
	localizerInterface, exists := c.Get(i18n.LocalizerContextKey)
	if !exists {
		return msgID // Fallback: Trả về ID thô nếu Localizer không tồn tại
	}

	localizer, ok := localizerInterface.(*i18n.Localizer)
	if !ok {
		return msgID // Fallback nếu cast thất bại
	}

	// Cấu hình dịch với MessageID và TemplateData
	translatedMsg, err := localizer.Localize(&go_i18n.LocalizeConfig{
		MessageID:    msgID,
		TemplateData: templateData,
	})

	if err != nil {
		return msgID // Nếu dịch thất bại, trả về ID thô
	}

	return translatedMsg
}

// ----------------------------------------------------
// PUBLIC RESPONSE FUNCTIONS
// ----------------------------------------------------

// Success tạo và gửi phản hồi thành công.
// msgID: ID của thông điệp cần dịch (ví dụ: "user_created_success").
func Success(c *gin.Context, httpStatus int, msgID string, data any, meta *PaginationMeta) {
	translatedMsg := getTranslatedMessage(c, msgID, nil)

	response := StandardResponse{
		Success: true,
		Message: translatedMsg,
		Data:    data,
		Meta:    meta,
	}
	c.JSON(httpStatus, response)
}

// Error tạo và gửi phản hồi lỗi nghiệp vụ.
// msgID: ID của thông điệp lỗi cần dịch.
// details: Dữ liệu chi tiết về lỗi (ví dụ: lỗi validation).
func Error(c *gin.Context, httpStatus int, customCode string, msgID string, details any) {
	// Dịch thông báo lỗi (details có thể được dùng làm template data cho thông báo lỗi)
	translatedMsg := getTranslatedMessage(c, msgID, details)

	response := StandardResponse{
		Success: false,
		Code:    customCode,
		Message: translatedMsg,
		Data:    details,
	}
	c.JSON(httpStatus, response)
}

// AbortWithError là hàm tiện ích cho Gin, dùng khi middleware cần dừng request.
func AbortWithError(c *gin.Context, httpStatus int, customCode string, msgID string) {
	translatedMsg := getTranslatedMessage(c, msgID, nil)

	response := StandardResponse{
		Success: false,
		Code:    customCode,
		Message: translatedMsg,
	}
	c.AbortWithStatusJSON(httpStatus, response)
}
