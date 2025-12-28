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

// UserOrganizationRole holds the schema definition for user_organization_roles.
type UserOrganizationRole struct {
	ent.Schema
}

// Annotations of the UserOrganizationRole.
func (UserOrganizationRole) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "user_organization_roles", Schema: "identify"},
	}
}

// Fields of the UserOrganizationRole.
func (UserOrganizationRole) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("organization_id", uuid.UUID{}),
		field.UUID("role_id", uuid.UUID{}),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Indexes of the UserOrganizationRole.
func (UserOrganizationRole) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "organization_id", "role_id").Unique(),
		index.Fields("user_id"),
		index.Fields("organization_id"),
		index.Fields("role_id"),
	}
}

// Edges of the UserOrganizationRole.
func (UserOrganizationRole) Edges() []ent.Edge {
	return nil
}
