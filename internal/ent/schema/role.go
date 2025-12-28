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

// Role holds the schema definition for the Role entity.
type Role struct {
	ent.Schema
}

// Annotations of the Role.
func (Role) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "roles", Schema: "identify"},
	}
}

// Fields of the Role.
func (Role) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.String("name").
			Unique().
			NotEmpty().
			Comment("Role name (UPPER_SNAKE_CASE)"),
		field.String("slug").
			Unique().
			Optional().
			Nillable().
			Comment("URL-friendly identifier"),
		field.Enum("scope").
			Values("global", "organization").
			Optional().
			Nillable().
			Comment("Scope: global (system-wide) or organization (within org)"),
		field.String("description").
			Optional().
			Nillable(),
		field.Bool("is_system").
			Default(false).
			Comment("Whether this is a system-defined role that cannot be deleted"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Indexes of the Role.
func (Role) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name"),
		index.Fields("slug"),
		index.Fields("scope"),
		index.Fields("is_system"),
	}
}

// Edges of the Role.
func (Role) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("permissions", Permission.Type).
			Through("role_permissions", RolePermission.Type),
	}
}

