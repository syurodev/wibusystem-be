package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gofrs/uuid/v5"
)

// OrgPendingInvite holds the schema definition for the OrganizationPendingInvite entity.
type OrgPendingInvite struct {
	ent.Schema
}

// Annotations of the OrgPendingInvite.
func (OrgPendingInvite) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "organization_pending_invites", Schema: "identify"},
	}
}

// Fields of the OrgPendingInvite.
func (OrgPendingInvite) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.UUID("organization_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("invited_by", uuid.UUID{}),
		field.Enum("status").
			Values("pending", "approved", "rejected", "expired").
			Default("pending"),
		field.UUID("approved_by", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Time("processed_at").
			Optional().
			Nillable(),
		field.Time("expires_at"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the OrgPendingInvite.
func (OrgPendingInvite) Edges() []ent.Edge {
	return nil
}
