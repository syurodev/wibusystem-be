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

// UnitProgress holds the schema definition for the UnitProgress entity.
type UnitProgress struct {
	ent.Schema
}

// Annotations of the UnitProgress.
func (UnitProgress) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "unit_progress", Schema: "catalog"},
	}
}

// Fields of the UnitProgress.
func (UnitProgress) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.UUID("user_id", uuid.UUID{}),
		field.String("media_type").
			NotEmpty(),
		field.UUID("media_id", uuid.UUID{}),
		field.UUID("unit_id", uuid.UUID{}),
		field.Enum("status").
			Values("in_progress", "completed").
			Default("in_progress"),
		field.JSON("position", json.RawMessage{}).
			Optional(),
		field.Time("started_at").
			Default(time.Now),
		field.Time("completed_at").
			Optional().
			Nillable(),
		field.Time("last_accessed_at").
			Default(time.Now),
	}
}

// Edges of the UnitProgress.
func (UnitProgress) Edges() []ent.Edge {
	return nil
}
