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

// TranslationContribution holds the schema for translation contributions.
type TranslationContribution struct {
	ent.Schema
}

// Annotations of the TranslationContribution.
func (TranslationContribution) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "translation_contributions", Schema: "community"},
	}
}

// Fields of the TranslationContribution.
func (TranslationContribution) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(NewUUIDV7),
		field.UUID("chapter_id", uuid.UUID{}),
		field.UUID("contributor_id", uuid.UUID{}),
		field.String("language").NotEmpty(),
		field.String("contribution_type").NotEmpty(), // full, partial, correction
		field.String("title").Optional().Nillable(),
		field.Text("content").NotEmpty(),
		field.Text("contributor_notes").Optional().Nillable(),
		field.String("status").Default("draft"), // draft, pending_review, approved, rejected
		field.UUID("reviewed_by", uuid.UUID{}).Optional().Nillable(),
		field.Time("reviewed_at").Optional().Nillable(),
		field.String("review_notes").Optional().Nillable(),
		field.UUID("official_translation_id", uuid.UUID{}).Optional().Nillable(),
		field.Int("credit_points").Default(0),
		field.Bool("is_credited").Default(false),
		field.Int("word_count").Default(0),
		field.Int("character_count").Default(0),
		field.Int("upvote_count").Default(0),
		field.Int("downvote_count").Default(0),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
		field.Time("deleted_at").Optional().Nillable(),
	}
}

// Indexes of the TranslationContribution.
func (TranslationContribution) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("chapter_id"),
		index.Fields("contributor_id"),
		index.Fields("status"),
	}
}

// Edges of the TranslationContribution.
func (TranslationContribution) Edges() []ent.Edge {
	return nil
}
