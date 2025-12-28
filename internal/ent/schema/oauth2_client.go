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

// OAuth2Client holds the schema definition for the OAuth2Client entity.
type OAuth2Client struct {
	ent.Schema
}

// Annotations of the OAuth2Client.
func (OAuth2Client) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "oauth2_clients", Schema: "identify"},
	}
}

// Fields of the OAuth2Client.
func (OAuth2Client) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.String("client_name").
			NotEmpty(),
		field.String("secret_hash").
			Optional().
			Sensitive(),
		field.Other("redirect_uris", Strings{}).
			Optional().
			SchemaType(map[string]string{
				dialect.Postgres: "text[]",
			}),
		field.Other("grant_types", Strings{}).
			Optional().
			SchemaType(map[string]string{
				dialect.Postgres: "text[]",
			}),
		field.Other("response_types", Strings{}).
			Optional().
			SchemaType(map[string]string{
				dialect.Postgres: "text[]",
			}),
		field.Other("scopes", Strings{}).
			Optional().
			SchemaType(map[string]string{
				dialect.Postgres: "text[]",
			}),
		field.Bool("is_public").
			Default(false),
		field.Bool("is_internal").
			Default(false),
		field.String("token_endpoint_auth_method").
			Default("client_secret_basic"),
		field.UUID("organization_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("owner_user_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.String("client_uri").
			Optional().
			Nillable(),
		field.String("logo_url").
			Optional().
			Nillable(),
		field.String("terms_of_service_url").
			Optional().
			Nillable(),
		field.String("policy_url").
			Optional().
			Nillable(),
		field.Bool("active").
			Default(true),

		// Audit
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the OAuth2Client.
func (OAuth2Client) Edges() []ent.Edge {
	return nil
}
