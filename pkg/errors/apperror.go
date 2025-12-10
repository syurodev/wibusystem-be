package errors

import (
	"fmt"
	"net/http"
)

// AppError là struct error thống nhất cho toàn bộ application
type AppError struct {
	StatusCode int    `json:"-"`
	ErrCode    string `json:"code"`
	I18nKey    string `json:"i18n_key"`
	Message    string `json:"message"`
}

func (e *AppError) Error() string {
	return e.Message
}

// Is implements errors.Is interface
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.StatusCode == t.StatusCode && e.ErrCode == t.ErrCode
}

// Helper functions để tạo các loại error phổ biến

// NotFound trả về 404 error
func NotFound(i18nKey, msg string) *AppError {
	return &AppError{
		StatusCode: http.StatusNotFound,
		ErrCode:    "NOT_FOUND",
		I18nKey:    i18nKey,
		Message:    msg,
	}
}

// NotFoundf trả về 404 error với format string
func NotFoundf(i18nKey, format string, args ...any) *AppError {
	return NotFound(i18nKey, fmt.Sprintf(format, args...))
}

// BadRequest trả về 400 error
func BadRequest(i18nKey, msg string) *AppError {
	return &AppError{
		StatusCode: http.StatusBadRequest,
		ErrCode:    "BAD_REQUEST",
		I18nKey:    i18nKey,
		Message:    msg,
	}
}

// BadRequestf trả về 400 error với format string
func BadRequestf(i18nKey, format string, args ...any) *AppError {
	return BadRequest(i18nKey, fmt.Sprintf(format, args...))
}

// Conflict trả về 409 error
func Conflict(i18nKey, msg string) *AppError {
	return &AppError{
		StatusCode: http.StatusConflict,
		ErrCode:    "CONFLICT",
		I18nKey:    i18nKey,
		Message:    msg,
	}
}

// Conflictf trả về 409 error với format string
func Conflictf(i18nKey, format string, args ...any) *AppError {
	return Conflict(i18nKey, fmt.Sprintf(format, args...))
}

// Forbidden trả về 403 error
func Forbidden(i18nKey, msg string) *AppError {
	return &AppError{
		StatusCode: http.StatusForbidden,
		ErrCode:    "FORBIDDEN",
		I18nKey:    i18nKey,
		Message:    msg,
	}
}

// Unauthorized trả về 401 error
func Unauthorized(i18nKey, msg string) *AppError {
	return &AppError{
		StatusCode: http.StatusUnauthorized,
		ErrCode:    "UNAUTHORIZED",
		I18nKey:    i18nKey,
		Message:    msg,
	}
}

// Internal trả về 500 error
func Internal(i18nKey, msg string) *AppError {
	return &AppError{
		StatusCode: http.StatusInternalServerError,
		ErrCode:    "INTERNAL_ERROR",
		I18nKey:    i18nKey,
		Message:    msg,
	}
}

// Internalf trả về 500 error với format string
func Internalf(i18nKey, format string, args ...any) *AppError {
	return Internal(i18nKey, fmt.Sprintf(format, args...))
}

// ValidationFailed trả về 422 error
func ValidationFailed(i18nKey, msg string) *AppError {
	return &AppError{
		StatusCode: http.StatusUnprocessableEntity,
		ErrCode:    "VALIDATION_FAILED",
		I18nKey:    i18nKey,
		Message:    msg,
	}
}

// AsAppError kiểm tra và convert error thành AppError nếu có thể
func AsAppError(err error) (*AppError, bool) {
	if err == nil {
		return nil, false
	}
	appErr, ok := err.(*AppError)
	return appErr, ok
}
