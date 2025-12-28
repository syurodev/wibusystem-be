package schema

import (
	"encoding/json"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gofrs/uuid/v5"
)

// OrgMember holds the schema definition for the OrganizationMembership entity.
type OrgMember struct {
	ent.Schema
}

// Annotations of the OrgMember.
func (OrgMember) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "organization_members", Schema: "identify"},
	}
}

// Fields of the OrgMember.
func (OrgMember) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("organization_id", uuid.UUID{}),
		field.String("status").
			Default("active"),
		field.Enum("role").
			Values("owner", "admin", "moderator", "member").
			Default("member"),
		field.Bool("is_active").
			Default(true),

		// Statistics
		field.Int("contribution_count").
			Default(0),
		field.Float("quality_score").
			Default(0),

		// Metadata
		field.JSON("metadata", json.RawMessage{}).
			Optional(),

		// Invitation
		field.UUID("invited_by", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Time("invited_at").
			Optional().
			Nillable(),
		field.Time("joined_at").
			Optional().
			Nillable(),
		field.Time("left_at").
			Optional().
			Nillable(),

		// Audit
		field.UUID("created_by", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("updated_by", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("deleted_by", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Int("version").
			Default(1),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}

// Edges of the OrgMember.
func (OrgMember) Edges() []ent.Edge {
	return nil
}
