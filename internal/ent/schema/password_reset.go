package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gofrs/uuid/v5"
)

// PasswordReset holds the schema definition for the PasswordReset entity.
type PasswordReset struct {
	ent.Schema
}

// Annotations of the PasswordReset.
func (PasswordReset) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "password_reset_tokens", Schema: "identify"},
	}
}

// Fields of the PasswordReset.
func (PasswordReset) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.UUID("user_id", uuid.UUID{}),
		field.String("token").
			Unique().
			NotEmpty(),
		field.Time("expires_at"),
		field.Time("used_at").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the PasswordReset.
func (PasswordReset) Edges() []ent.Edge {
	return nil
}
