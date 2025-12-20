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
