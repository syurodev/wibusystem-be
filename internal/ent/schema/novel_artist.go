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

// NovelArtist holds the schema definition for the NovelArtist entity.
// Đây là junction table cho quan hệ many-to-many giữa Novel và Artist.
type NovelArtist struct {
	ent.Schema
}

// Annotations of the NovelArtist.
func (NovelArtist) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "novel_artists", Schema: "catalog"},
	}
}

// Fields of the NovelArtist.
func (NovelArtist) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.UUID("novel_id", uuid.UUID{}),
		field.UUID("artist_id", uuid.UUID{}),
		field.String("role").
			Default("cover_artist"), // "cover_artist", "illustrator", "character_designer"
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

// Edges of the NovelArtist.
func (NovelArtist) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("novel", Novel.Type).
			Field("novel_id").
			Unique().
			Required(),
		edge.To("artist", Artist.Type).
			Field("artist_id").
			Unique().
			Required(),
	}
}
