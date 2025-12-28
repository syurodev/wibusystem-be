package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gofrs/uuid/v5"
)

// NovelChapterTranslation holds the schema for chapter translations.
type NovelChapterTranslation struct {
	ent.Schema
}

// Annotations of the NovelChapterTranslation.
func (NovelChapterTranslation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "novel_chapter_translations", Schema: "community"},
	}
}

// Fields of the NovelChapterTranslation.
func (NovelChapterTranslation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(NewUUIDV7),
		field.UUID("chapter_id", uuid.UUID{}),
		field.String("language").NotEmpty(),
		field.String("title").NotEmpty(),
		field.Text("content").NotEmpty(), // Changed to Text for large content
		field.Text("translator_notes").Optional().Nillable(),
		field.UUID("organization_id", uuid.UUID{}).Optional().Nillable(),
		field.Int("version").Default(1),
		field.String("status").Default("draft"), // draft, pending_review, published
		field.Int("word_count").Default(0),
		field.Int("character_count").Default(0),
		field.Float("quality_score").Optional().Nillable(),
		field.Float("reviewer_rating").Optional().Nillable(),
		// Use Int64 for counts to match domain int64 (or at least provide capacity)
		field.Int64("view_count").Default(0),
		field.Int("like_count").Default(0),
		field.Int("comment_count").Default(0),
		field.Int("contribution_count").Default(0),
		field.UUID("reviewed_by", uuid.UUID{}).Optional().Nillable(),
		field.String("review_notes").Optional().Nillable(),
		field.Time("reviewed_at").Optional().Nillable(),
		field.Time("published_at").Optional().Nillable(),
		field.UUID("created_by", uuid.UUID{}).Optional().Nillable(),
		field.UUID("updated_by", uuid.UUID{}).Optional().Nillable(),
		field.UUID("deleted_by", uuid.UUID{}).Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

// Indexes of the NovelChapterTranslation.
func (NovelChapterTranslation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("chapter_id", "language").Unique(),
		index.Fields("chapter_id"),
		index.Fields("organization_id"),
		index.Fields("created_by"),
	}
}

// Edges of the NovelChapterTranslation.
func (NovelChapterTranslation) Edges() []ent.Edge {
	return nil
}
