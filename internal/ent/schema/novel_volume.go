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

// NovelVolume holds the schema definition for the NovelVolume entity.
type NovelVolume struct {
	ent.Schema
}

// Annotations of the NovelVolume.
func (NovelVolume) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "novel_volumes", Schema: "catalog"},
	}
}

// Fields of the NovelVolume.
func (NovelVolume) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.UUID("novel_id", uuid.UUID{}),

		field.Int("volume_number"),
		field.String("title").
			NotEmpty(),
		field.String("slug").
			NotEmpty(),
		field.String("description").
			Optional().
			Nillable(),

		field.String("cover_image_url").
			Optional().
			Nillable(),

		// Statistics
		field.Int("chapter_count").
			Default(0),
		field.Int64("word_count").
			Default(0),

		// Display & Status
		field.Int("display_order").
			Default(0),
		field.Bool("is_published").
			Default(false),
		field.Time("published_at").
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

// Edges of the NovelVolume.
func (NovelVolume) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("novel", Novel.Type).
			Ref("volumes").
			Field("novel_id").
			Unique().
			Required(),
		edge.To("chapters", NovelChapter.Type),
	}
}
