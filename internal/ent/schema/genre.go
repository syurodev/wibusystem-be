package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gofrs/uuid/v5"
)

// Genre holds the schema definition for the Genre entity.
type Genre struct {
	ent.Schema
}

// Annotations of the Genre.
func (Genre) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "genres", Schema: "catalog"},
	}
}

// Fields of the Genre.
func (Genre) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.String("name").
			NotEmpty(),
		field.String("slug").
			Unique().
			NotEmpty(),
		field.String("description").
			Optional().
			Nillable(),

		// Parent genre for hierarchy
		field.UUID("parent_id", uuid.UUID{}).
			Optional().
			Nillable(),

		field.Bool("is_active").
			Default(true),

		// Statistics
		field.Int("novel_count").
			Default(0),
		field.Int("anime_count").
			Default(0),
		field.Int("manga_count").
			Default(0),
		field.Int64("active_readers").
			Default(0),
		field.Int64("total_views").
			Default(0),

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

// Edges of the Genre.
func (Genre) Edges() []ent.Edge {
	return nil
}
