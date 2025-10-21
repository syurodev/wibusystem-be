package response

// ErrorDetail represents the error object used in standard responses.
type ErrorDetail struct {
	Code        string `json:"code,omitempty"`
	Description string `json:"description,omitempty"`
}

// StandardResponse is the unified response envelope for all APIs.
type StandardResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Data    any  `json:"data,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
	Meta    *Pagination  `json:"meta,omitempty"`
}

// Pagination represents pagination metadata, typically included in the 'meta' field.
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
	TotalItems int64 `json:"total_items"`
}

// NewSuccessResponse creates a new successful response.
func NewSuccessResponse(data any, message string, pagination *Pagination) StandardResponse {
	return StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    pagination,
	}
}

// NewErrorResponse creates a new error response.
func NewErrorResponse(code, description, message string) StandardResponse {
	return StandardResponse{
		Success: false,
		Message: message,
		Error: &ErrorDetail{
			Code:        code,
			Description: description,
		},
	}
}
