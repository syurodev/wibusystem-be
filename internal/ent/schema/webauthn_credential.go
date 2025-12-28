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

// WebAuthnCredential holds the schema definition for the WebAuthnCredential entity.
type WebAuthnCredential struct {
	ent.Schema
}

// Annotations of the WebAuthnCredential.
func (WebAuthnCredential) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "webauthn_credentials", Schema: "identify"},
	}
}

// Fields of the WebAuthnCredential.
func (WebAuthnCredential) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.UUID("user_id", uuid.UUID{}),
		field.String("credential_id").
			Unique().
			NotEmpty(),
		field.Bytes("public_key"),
		field.Enum("attestation_type").
			Values("none", "indirect", "direct").
			Default("none"),
		field.Bytes("aaguid").
			Optional(),
		field.Int32("sign_count").
			Default(0),
		field.Other("transports", Strings{}).
			Optional().
			SchemaType(map[string]string{
				dialect.Postgres: "text[]",
			}),
		field.Bool("backup_eligible").
			Default(false),
		field.Bool("backup_state").
			Default(false),
		field.String("credential_name").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("last_used_at").
			Optional().
			Nillable(),
	}
}

// Edges of the WebAuthnCredential.
func (WebAuthnCredential) Edges() []ent.Edge {
	return nil
}
