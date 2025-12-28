package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gofrs/uuid/v5"
)

// EmailVerification holds the schema definition for the EmailVerification entity.
type EmailVerification struct {
	ent.Schema
}

// Annotations of the EmailVerification.
func (EmailVerification) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "email_verification_tokens", Schema: "identify"},
	}
}

// Fields of the EmailVerification.
func (EmailVerification) Fields() []ent.Field {
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

// Edges of the EmailVerification.
func (EmailVerification) Edges() []ent.Edge {
	return nil
}
