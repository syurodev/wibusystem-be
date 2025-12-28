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

// MediaProgress holds the schema definition for the MediaProgress entity.
type MediaProgress struct {
	ent.Schema
}

// Annotations of the MediaProgress.
func (MediaProgress) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "media_progress", Schema: "catalog"},
	}
}

// Fields of the MediaProgress.
func (MediaProgress) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.UUID("user_id", uuid.UUID{}),
		field.String("media_type").
			NotEmpty(), // "novel", "manga", "anime"
		field.UUID("media_id", uuid.UUID{}),
		field.UUID("current_unit_id", uuid.UUID{}),
		field.JSON("position", json.RawMessage{}).
			Optional(),
		field.Int("total_units").
			Default(0),
		field.Int("completed_units").
			Default(0),
		field.Float("progress_percentage").
			Default(0),
		field.Time("last_accessed_at").
			Default(time.Now),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the MediaProgress.
func (MediaProgress) Edges() []ent.Edge {
	return nil
}
