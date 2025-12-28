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

// Organization holds the schema definition for the Organization entity.
type Organization struct {
	ent.Schema
}

// Annotations of the Organization.
func (Organization) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "organizations", Schema: "identify"},
	}
}

// Fields of the Organization.
func (Organization) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.String("name").
			NotEmpty(),
		field.String("slug").
			Unique().
			NotEmpty(),
		field.Enum("status").
			Values("active", "suspended", "deleted").
			Default("active"),
		field.JSON("description", json.RawMessage{}).
			Optional(),
		field.String("avatar_url").
			Optional().
			Nillable(),
		field.JSON("settings", json.RawMessage{}).
			Optional(),

		// Capabilities
		field.Bool("is_recruiting").
			Default(false),
		field.Bool("can_translate").
			Default(true),
		field.Bool("can_proofread").
			Default(false),
		field.Bool("can_edit").
			Default(false),

		// Statistics
		field.Int("member_count").
			Default(0),
		field.Int("active_projects").
			Default(0),
		field.Int("completed_translations").
			Default(0),
		field.Int("report_count").
			Default(0),

		// Metadata (JSONB)
		field.JSON("metadata", json.RawMessage{}).
			Optional(),

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

// Edges of the Organization.
func (Organization) Edges() []ent.Edge {
	return nil
}
