package schema

import (
	"encoding/json"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gofrs/uuid/v5"
)

// OAuth2Session holds the schema definition for the OAuth2Session entity.
type OAuth2Session struct {
	ent.Schema
}

// Annotations of the OAuth2Session.
func (OAuth2Session) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "oauth2_sessions", Schema: "identify"},
	}
}

// Fields of the OAuth2Session.
func (OAuth2Session) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.String("signature").
			Unique().
			NotEmpty(),
		field.String("request_id").
			NotEmpty(),
		field.String("session_type").
			NotEmpty(),
		field.JSON("session_data", json.RawMessage{}),
		field.Time("expires_at"),
		field.String("client_id").
			NotEmpty(),
		field.String("subject_id").
			NotEmpty(),
		field.Bool("active").
			Default(true),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Indexes of the OAuth2Session.
func (OAuth2Session) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("subject_id", "active"),
		index.Fields("client_id"),
	}
}

// Edges of the OAuth2Session.
func (OAuth2Session) Edges() []ent.Edge {
	return nil
}
