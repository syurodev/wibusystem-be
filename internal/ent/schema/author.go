package schema

import (
	"encoding/json"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gofrs/uuid/v5"
)

// Author holds the schema definition for the Author entity.
type Author struct {
	ent.Schema
}

// Annotations of the Author.
func (Author) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "authors", Schema: "catalog"},
	}
}

// Fields of the Author.
func (Author) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.UUID("user_id", uuid.UUID{}).
			Optional().
			Nillable(),

		field.String("name").
			NotEmpty(),
		field.String("slug").
			Unique().
			NotEmpty(),

		// Biography (JSONB)
		field.JSON("biography", json.RawMessage{}).
			Optional(),
		field.String("avatar_url").
			Optional().
			Nillable(),
		// Social links (JSONB)
		field.JSON("social_links", json.RawMessage{}).
			Optional(),

		// Statistics
		field.Int("novel_count").
			Default(0),
		field.Int("total_chapters").
			Default(0),
		field.Int64("total_views").
			Default(0),
		field.Int("follower_count").
			Default(0),

		field.Bool("is_verified").
			Default(false),

		// Metadata (JSONB)
		field.JSON("metadata", json.RawMessage{}).
			Optional(),

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

// Edges of the Author.
func (Author) Edges() []ent.Edge {
	return nil
}
