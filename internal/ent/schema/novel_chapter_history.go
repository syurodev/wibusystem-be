package schema

import (
	"encoding/json"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gofrs/uuid/v5"
)

// NovelChapterHistory holds the schema for chapter history/audit logs.
type NovelChapterHistory struct {
	ent.Schema
}

// Annotations of the NovelChapterHistory.
func (NovelChapterHistory) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "novel_chapter_histories", Schema: "catalog"},
	}
}

// Fields of the NovelChapterHistory.
func (NovelChapterHistory) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(NewUUIDV7),
		field.UUID("chapter_id", uuid.UUID{}),
		field.UUID("volume_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("novel_id", uuid.UUID{}),
		field.Int("version_number").Default(1),
		field.String("action").NotEmpty(), // created, updated, published, deleted
		field.String("title").Optional().Nillable(),
		field.String("slug").Optional().Nillable(),
		field.Int("chapter_number").Optional().Nillable(),
		field.String("status").Optional().Nillable(),
		field.Int("word_count").Optional().Nillable(),
		field.Int("character_count").Optional().Nillable(),
		field.JSON("changed_fields", json.RawMessage{}).Optional(),
		field.String("change_summary").Optional().Nillable(),
		field.Bool("content_changed").Default(false),
		field.UUID("changed_by", uuid.UUID{}),
		field.String("request_id").Optional().Nillable(),
		field.String("ip_address").Optional().Nillable(),
		field.String("user_agent").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

// Indexes of the NovelChapterHistory.
func (NovelChapterHistory) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("chapter_id"),
		index.Fields("novel_id"),
		index.Fields("changed_by"),
	}
}

// Edges of the NovelChapterHistory.
func (NovelChapterHistory) Edges() []ent.Edge {
	return nil
}
