// Package schema contains Ent schema custom types.
package schema

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

// Strings is a custom type that represents a PostgreSQL TEXT[] array.
// It implements the sql.Scanner and driver.Valuer interfaces for proper
// serialization/deserialization with PostgreSQL native array format.
type Strings []string

// Value implements the driver.Valuer interface.
// Converts the Strings slice to PostgreSQL array format: {val1,val2,val3}
func (s Strings) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	if len(s) == 0 {
		return "{}", nil
	}

	// Escape and quote each element
	escaped := make([]string, len(s))
	for i, v := range s {
		// Escape backslashes and double quotes
		v = strings.ReplaceAll(v, `\`, `\\`)
		v = strings.ReplaceAll(v, `"`, `\"`)
		escaped[i] = `"` + v + `"`
	}

	return "{" + strings.Join(escaped, ",") + "}", nil
}

// Scan implements the sql.Scanner interface.
// Parses PostgreSQL array format: {val1,val2,val3} or {"val1","val2"}
func (s *Strings) Scan(src interface{}) error {
	if src == nil {
		*s = nil
		return nil
	}

	var source string
	switch v := src.(type) {
	case []byte:
		source = string(v)
	case string:
		source = v
	default:
		return fmt.Errorf("unsupported type for Strings: %T", src)
	}

	// Handle empty array
	if source == "{}" || source == "" {
		*s = Strings{}
		return nil
	}

	// Remove the curly braces
	source = strings.TrimPrefix(source, "{")
	source = strings.TrimSuffix(source, "}")

	// Parse the array elements
	result, err := parsePostgresArray(source)
	if err != nil {
		return fmt.Errorf("failed to parse postgres array: %w", err)
	}

	*s = result
	return nil
}

// parsePostgresArray parses the inner content of a PostgreSQL array string.
// Handles both quoted ("value") and unquoted (value) elements.
func parsePostgresArray(s string) ([]string, error) {
	if s == "" {
		return []string{}, nil
	}

	var result []string
	var current strings.Builder
	inQuotes := false
	escaped := false

	for i := 0; i < len(s); i++ {
		c := s[i]

		if escaped {
			current.WriteByte(c)
			escaped = false
			continue
		}

		switch c {
		case '\\':
			escaped = true
		case '"':
			inQuotes = !inQuotes
		case ',':
			if inQuotes {
				current.WriteByte(c)
			} else {
				result = append(result, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(c)
		}
	}

	// Add the last element
	if current.Len() > 0 || len(result) > 0 {
		result = append(result, current.String())
	}

	return result, nil
}
