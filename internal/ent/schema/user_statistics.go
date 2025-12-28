package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/gofrs/uuid/v5"
)

// UserStatistics holds the schema definition for the UserStatistics entity.
type UserStatistics struct {
	ent.Schema
}

// Annotations of the UserStatistics.
func (UserStatistics) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "user_statistics", Schema: "identify"},
	}
}

// Fields of the UserStatistics.
func (UserStatistics) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(NewUUIDV7),
		field.UUID("user_id", uuid.UUID{}).
			Unique(),
		field.Int("follower_count").
			Default(0),
		field.Int("following_count").
			Default(0),
		field.Int("novel_count").
			Default(0),
		field.Int("manga_count").
			Default(0),
		field.Int("anime_count").
			Default(0),
		field.Time("last_content_updated_at").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the UserStatistics.
func (UserStatistics) Edges() []ent.Edge {
	return nil
}
