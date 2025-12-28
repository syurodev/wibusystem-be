package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/gofrs/uuid/v5"
	"github.com/pgvector/pgvector-go"
)

// NovelEmbedding holds the schema definition for novel vector embeddings.
type NovelEmbedding struct {
	ent.Schema
}

// Annotations of the NovelEmbedding.
func (NovelEmbedding) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "novel_embeddings", Schema: "catalog"},
	}
}

// Fields of the NovelEmbedding.
func (NovelEmbedding) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(NewUUIDV7),
		field.UUID("novel_id", uuid.UUID{}).Unique(),
		field.Other("embedding", pgvector.Vector{}).
			SchemaType(map[string]string{
				dialect.Postgres: "vector(384)", // Dimension for embedding model
			}),
		field.String("model_version").NotEmpty(),
		field.String("source_hash").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Indexes of the NovelEmbedding.
func (NovelEmbedding) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("novel_id"),
		// HNSW index for similarity search
		index.Fields("embedding").
			Annotations(
				entsql.IndexType("hnsw"),
				entsql.OpClass("vector_cosine_ops"),
			),
	}
}

// Edges of the NovelEmbedding.
func (NovelEmbedding) Edges() []ent.Edge {
	return nil
}
