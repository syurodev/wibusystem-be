package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gofrs/uuid/v5"
)

// OrgReport holds the schema definition for the OrganizationReport entity.
type OrgReport struct {
	ent.Schema
}

// Annotations of the OrgReport.
func (OrgReport) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "organization_reports", Schema: "identify"},
	}
}

// Fields of the OrgReport.
func (OrgReport) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.UUID("organization_id", uuid.UUID{}),
		field.UUID("reporter_id", uuid.UUID{}),
		field.String("reason").
			NotEmpty(),
		field.String("description").
			Optional().
			Nillable(),
		field.String("org_response").
			Optional().
			Nillable(),
		field.UUID("org_responded_by", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Time("org_responded_at").
			Optional().
			Nillable(),
		field.Enum("status").
			Values("pending", "org_responded", "reviewing", "resolved", "dismissed").
			Default("pending"),
		field.UUID("resolved_by", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Time("resolved_at").
			Optional().
			Nillable(),
		field.String("resolution_note").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the OrgReport.
func (OrgReport) Edges() []ent.Edge {
	return nil
}
