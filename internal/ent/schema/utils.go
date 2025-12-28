// Package schema contains Ent schema utilities.
package schema

import "github.com/gofrs/uuid/v5"

// NewUUIDV7 generates a new UUID v7.
// Use this as the default value for UUID fields in all Ent schemas.
//
// Example:
//
//	field.UUID("id", uuid.UUID{}).Default(NewUUIDV7)
func NewUUIDV7() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}
