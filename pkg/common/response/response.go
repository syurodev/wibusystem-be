package response

// ErrorDetail represents the error object used in standard responses.
type ErrorDetail struct {
	Code        string `json:"code"`
	Description string `json:"description,omitempty"`
}

// StandardResponse is the unified response envelope for all APIs.
//
// Example:
//
//	{
//	  "success": true,
//	  "message": "Request processed successfully",
//	  "data": {},
//	  "error": null,
//	  "meta": {}
//	}
type StandardResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Data    interface{}            `json:"data"`
	Error   *ErrorDetail           `json:"error"`
	Meta    map[string]interface{} `json:"meta"`
}

// Deprecated: use StandardResponse instead.
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// Deprecated: use StandardResponse.Error instead.
type ErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}
