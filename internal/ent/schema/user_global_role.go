package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gofrs/uuid/v5"
)

// UserGlobalRole holds the schema definition for user_global_roles.
type UserGlobalRole struct {
	ent.Schema
}

// Annotations of the UserGlobalRole.
func (UserGlobalRole) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "user_global_roles", Schema: "identify"},
	}
}

// Fields of the UserGlobalRole.
func (UserGlobalRole) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("role_id", uuid.UUID{}),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Indexes of the UserGlobalRole.
func (UserGlobalRole) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "role_id").Unique(),
		index.Fields("user_id"),
		index.Fields("role_id"),
	}
}

// Edges of the UserGlobalRole.
func (UserGlobalRole) Edges() []ent.Edge {
	return nil
}
