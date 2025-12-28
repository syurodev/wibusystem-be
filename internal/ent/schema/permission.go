package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gofrs/uuid/v5"
)

// Permission holds the schema definition for the Permission entity.
type Permission struct {
	ent.Schema
}

// Annotations of the Permission.
func (Permission) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "permissions", Schema: "identify"},
	}
}

// Fields of the Permission.
func (Permission) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.String("name").
			Unique().
			NotEmpty().
			Comment("Permission name in format resource:action (e.g., user:create)"),
		field.Enum("scope").
			Values("global", "organization").
			Comment("Scope: global (system-wide) or organization (within org)"),
		field.String("description").
			Optional().
			Nillable(),
		field.String("resource").
			Optional().
			Nillable().
			Comment("Resource type this permission applies to (e.g., user, content, role)"),
		field.String("action").
			Optional().
			Nillable().
			Comment("Action allowed (view, create, update, delete, etc.)"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Indexes of the Permission.
func (Permission) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("scope"),
		index.Fields("resource"),
		index.Fields("name"),
	}
}

// Edges of the Permission.
func (Permission) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("roles", Role.Type).
			Ref("permissions").
			Through("role_permissions", RolePermission.Type),
	}
}
