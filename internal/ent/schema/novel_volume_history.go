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

// NovelVolumeHistory holds the schema for volume history/audit logs.
type NovelVolumeHistory struct {
	ent.Schema
}

// Annotations of the NovelVolumeHistory.
func (NovelVolumeHistory) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "novel_volume_histories", Schema: "catalog"},
	}
}

// Fields of the NovelVolumeHistory.
func (NovelVolumeHistory) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(NewUUIDV7),
		field.UUID("volume_id", uuid.UUID{}),
		field.UUID("novel_id", uuid.UUID{}),
		field.Int("version_number").Default(1),
		field.String("action").NotEmpty(), // created, updated, published, unpublished, deleted
		field.String("title").Optional().Nillable(),
		field.String("slug").Optional().Nillable(),
		field.Int("volume_number").Optional().Nillable(),
		field.Bool("is_published").Optional().Nillable(),
		field.Int("chapter_count").Optional().Nillable(),
		field.Int("word_count").Optional().Nillable(),
		field.JSON("changed_fields", json.RawMessage{}).Optional(),
		field.String("change_summary").Optional().Nillable(),
		field.UUID("changed_by", uuid.UUID{}),
		field.String("request_id").Optional().Nillable(),
		field.String("ip_address").Optional().Nillable(),
		field.String("user_agent").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

// Indexes of the NovelVolumeHistory.
func (NovelVolumeHistory) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("volume_id"),
		index.Fields("novel_id"),
		index.Fields("changed_by"),
	}
}

// Edges of the NovelVolumeHistory.
func (NovelVolumeHistory) Edges() []ent.Edge {
	return nil
}
