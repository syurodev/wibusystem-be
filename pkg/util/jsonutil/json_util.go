package jsonutil

import (
	"encoding/json"
)

// ToJSONRawMessage converts any value to json.RawMessage safely
func ToJSONRawMessage(v any) json.RawMessage {
	if v == nil {
		return nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return data
}

// ExtractTextField trích xuất field "text" từ JSON map
// Thường dùng để extract description từ biography JSONB
// Trả về nil nếu không có hoặc lỗi parse
func ExtractTextField(data json.RawMessage) *string {
	if len(data) == 0 {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil
	}
	if text, ok := m["text"].(string); ok && text != "" {
		return &text
	}
	return nil
}

// BuildTextJSON tạo JSON với format {"text": value}
// Trả về "{}" nếu value rỗng
// Error bị ignore vì input là string đơn giản
func BuildTextJSON(value string) json.RawMessage {
	if value == "" {
		return json.RawMessage("{}")
	}
	m := map[string]string{"text": value}
	data, _ := json.Marshal(m)
	return data
}

// ParseJSONOrDefault parses JSON string và validate
// Trả về default value nếu input empty
// Trả về error nếu JSON invalid
func ParseJSONOrDefault(jsonStr *string, defaultVal string) (json.RawMessage, error) {
	if jsonStr == nil || *jsonStr == "" {
		return json.RawMessage(defaultVal), nil
	}
	if !json.Valid([]byte(*jsonStr)) {
		return nil, ErrInvalidJSON
	}
	return json.RawMessage(*jsonStr), nil
}

// ErrInvalidJSON is returned when JSON parsing fails
var ErrInvalidJSON = &InvalidJSONError{}

// InvalidJSONError represents an invalid JSON error
type InvalidJSONError struct{}

func (e *InvalidJSONError) Error() string {
	return "invalid JSON"
}
