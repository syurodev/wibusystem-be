package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gofrs/uuid/v5"
)

// WebAuthnSession holds the schema definition for the WebAuthnSession entity.
type WebAuthnSession struct {
	ent.Schema
}

// Annotations of the WebAuthnSession.
func (WebAuthnSession) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "webauthn_sessions", Schema: "identify"},
	}
}

// Fields of the WebAuthnSession.
func (WebAuthnSession) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.UUID("user_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.String("challenge").
			Unique().
			NotEmpty(),
		field.Enum("session_type").
			Values("registration", "authentication"),
		field.String("user_agent").
			Optional().
			Nillable(),
		field.String("ip_address").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("expires_at"),
	}
}

// Edges of the WebAuthnSession.
func (WebAuthnSession) Edges() []ent.Edge {
	return nil
}
