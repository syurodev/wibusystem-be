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

// NovelAuthor holds the schema definition for the NovelAuthor entity.
// Đây là junction table cho quan hệ many-to-many giữa Novel và Author.
type NovelAuthor struct {
	ent.Schema
}

// Annotations of the NovelAuthor.
func (NovelAuthor) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "novel_authors", Schema: "catalog"},
	}
}

// Fields of the NovelAuthor.
func (NovelAuthor) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.UUID("novel_id", uuid.UUID{}),
		field.UUID("author_id", uuid.UUID{}),
		field.String("role").
			Default("original_author"), // "original_author", "co_author"
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

// Edges of the NovelAuthor.
func (NovelAuthor) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("novel", Novel.Type).
			Field("novel_id").
			Unique().
			Required(),
		edge.To("author", Author.Type).
			Field("author_id").
			Unique().
			Required(),
	}
}
