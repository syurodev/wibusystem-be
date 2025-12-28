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

// RolePermission holds the schema definition for the RolePermission entity.
// This is a junction table linking roles to permissions.
type RolePermission struct {
	ent.Schema
}

// Annotations of the RolePermission.
func (RolePermission) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "role_permissions", Schema: "identify"},
	}
}

// Fields of the RolePermission.
func (RolePermission) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.UUID("role_id", uuid.UUID{}).
			Comment("Reference to role"),
		field.UUID("permission_id", uuid.UUID{}).
			Comment("Reference to permission"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Indexes of the RolePermission.
func (RolePermission) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("role_id", "permission_id").Unique(),
		index.Fields("role_id"),
		index.Fields("permission_id"),
	}
}

// Edges of the RolePermission.
func (RolePermission) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("role", Role.Type).
			Required().
			Unique().
			Field("role_id"),
		edge.To("permission", Permission.Type).
			Required().
			Unique().
			Field("permission_id"),
	}
}
