package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gofrs/uuid/v5"
)

// Consent holds the schema definition for the OAuth2Consent entity.
type Consent struct {
	ent.Schema
}

// Annotations of the Consent.
func (Consent) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "oauth2_consents", Schema: "identify"},
	}
}

// Fields of the Consent.
func (Consent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("client_id", uuid.UUID{}),
		field.Other("granted_scopes", Strings{}).
			Optional().
			SchemaType(map[string]string{
				dialect.Postgres: "text[]",
			}),
		field.Bool("revoked").
			Default(false),
		field.Time("granted_at").
			Default(time.Now),
		field.Time("revoked_at").
			Optional().
			Nillable(),
		field.Time("last_used_at").
			Optional().
			Nillable(),
		field.Time("expires_at").
			Optional().
			Nillable(),
		field.Enum("consent_method").
			Values("explicit", "implicit", "remembered").
			Default("explicit"),
		field.String("ip_address").
			Optional().
			Nillable(),
		field.String("user_agent").
			Optional().
			Nillable(),
	}
}

// Edges of the Consent.
func (Consent) Edges() []ent.Edge {
	return nil
}
