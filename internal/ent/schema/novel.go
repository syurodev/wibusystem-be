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

// Novel holds the schema definition for the Novel entity.
type Novel struct {
	ent.Schema
}

// Annotations of the Novel.
func (Novel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "novels", Schema: "catalog"},
	}
}

// Fields of the Novel.
func (Novel) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.String("title").
			NotEmpty(),
		field.String("slug").
			Unique().
			NotEmpty(),

		// Owner
		field.UUID("owner_id", uuid.UUID{}),
		field.String("owner_type").
			Default("user"), // "user" or "organization"

		// Synopsis (JSONB)
		field.JSON("synopsis", json.RawMessage{}).
			Optional(),

		field.String("cover_image_url").
			Optional().
			Nillable(),
		field.String("thumbnail_url").
			Optional().
			Nillable(),

		field.Enum("status").
			Values("draft", "ongoing", "completed", "hiatus", "dropped").
			Default("draft"),
		field.Bool("is_oneshot").
			Default(false),

		// Original info
		field.String("original_language").
			Optional().
			Nillable(),
		field.String("original_title").
			Optional().
			Nillable(),

		// Statistics
		field.Int("total_volumes").
			Default(0),
		field.Int("total_chapters").
			Default(0),
		field.Int64("total_words").
			Default(0),
		field.Int64("view_count").
			Default(0),
		field.Int("favorite_count").
			Default(0),
		field.Float("rating_average").
			Default(0),
		field.Int("rating_count").
			Default(0),

		// Metadata (JSONB)
		field.JSON("metadata", json.RawMessage{}).
			Optional(),

		// Dates
		field.Time("first_published_at").
			Optional().
			Nillable(),
		field.Time("last_chapter_at").
			Optional().
			Nillable(),
		field.Time("completed_at").
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

// Edges of the Novel.
func (Novel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("novels").
			Field("owner_id").
			Unique().
			Required(),
		edge.To("volumes", NovelVolume.Type),
		edge.To("chapters", NovelChapter.Type),
	}
}
