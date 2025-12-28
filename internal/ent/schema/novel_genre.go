package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/gofrs/uuid/v5"
)

// NovelGenre holds the schema definition for the NovelGenre entity.
// Đây là junction table cho quan hệ many-to-many giữa Novel và Genre.
type NovelGenre struct {
	ent.Schema
}

// Annotations of the NovelGenre.
func (NovelGenre) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "novel_genres", Schema: "catalog"},
	}
}

// Fields of the NovelGenre.
func (NovelGenre) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.UUID("novel_id", uuid.UUID{}),
		field.UUID("genre_id", uuid.UUID{}),
		field.Int("display_order").
			Default(0),
		field.UUID("created_by", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the NovelGenre.
func (NovelGenre) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("novel", Novel.Type).
			Field("novel_id").
			Unique().
			Required(),
		edge.To("genre", Genre.Type).
			Field("genre_id").
			Unique().
			Required(),
	}
}
