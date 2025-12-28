package schema

import (
	"encoding/json"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/gofrs/uuid/v5"
)

// NovelChapter holds the schema definition for the NovelChapter entity.
type NovelChapter struct {
	ent.Schema
}

// Annotations of the NovelChapter.
func (NovelChapter) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "novel_chapters", Schema: "catalog"},
	}
}

// Fields of the NovelChapter.
func (NovelChapter) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.UUID("novel_id", uuid.UUID{}),
		field.UUID("volume_id", uuid.UUID{}).
			Optional().
			Nillable(),

		field.Int("chapter_number"),
		field.String("title").
			NotEmpty(),
		field.String("slug").
			NotEmpty(),

		// Content (JSONB)
		field.JSON("content", json.RawMessage{}).
			Optional(),

		// Content metadata
		field.Int("word_count").
			Default(0),
		field.Int("character_count").
			Default(0),

		// Access control
		field.Bool("is_free").
			Default(true),
		field.Float("price").
			Optional().
			Nillable(),
		field.String("currency").
			Optional().
			Nillable(),

		// Status
		field.Enum("status").
			Values("draft", "published", "scheduled").
			Default("draft"),

		// Statistics
		field.Int64("view_count").
			Default(0),
		field.Int("like_count").
			Default(0),
		field.Int("comment_count").
			Default(0),

		// Display
		field.Int("display_order").
			Default(0),

		// Author notes (JSONB)
		field.JSON("author_notes", json.RawMessage{}).
			Optional(),

		// Dates
		field.Time("published_at").
			Optional().
			Nillable(),
		field.Time("scheduled_at").
			Optional().
			Nillable(),

		// Audit
		field.UUID("created_by", uuid.UUID{}),
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

// Edges of the NovelChapter.
func (NovelChapter) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("novel", Novel.Type).
			Ref("chapters").
			Field("novel_id").
			Unique().
			Required(),
		edge.From("volume", NovelVolume.Type).
			Ref("chapters").
			Field("volume_id").
			Unique(),
	}
}
