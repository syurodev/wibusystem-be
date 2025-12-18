package embedding

import (
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/pgvector/pgvector-go"
)

// NovelEmbedding represents a vector embedding for a novel.
type NovelEmbedding struct {
	ID           uuid.UUID       `db:"id"`
	NovelID      uuid.UUID       `db:"novel_id"`
	Embedding    pgvector.Vector `db:"embedding"`
	ModelVersion string          `db:"model_version"`
	SourceHash   *string         `db:"source_hash"`
	CreatedAt    time.Time       `db:"created_at"`
	UpdatedAt    time.Time       `db:"updated_at"`
}

// NovelEmbeddingInput represents the data needed to generate an embedding.
type NovelEmbeddingInput struct {
	NovelID       uuid.UUID
	Title         string
	OriginalTitle *string
	Synopsis      string   // Extracted text from synopsis JSONB
	Genres        []string // Genre names
}

// BuildEmbeddingText builds the text to be embedded from the input.
func (i *NovelEmbeddingInput) BuildEmbeddingText() string {
	text := i.Title
	if i.OriginalTitle != nil && *i.OriginalTitle != "" {
		text += " | " + *i.OriginalTitle
	}
	if i.Synopsis != "" {
		text += " | " + i.Synopsis
	}
	if len(i.Genres) > 0 {
		for _, g := range i.Genres {
			text += " | " + g
		}
	}
	return text
}

// SimilarNovel represents a novel with its similarity distance.
type SimilarNovel struct {
	NovelID  uuid.UUID `db:"novel_id"`
	Distance float32   `db:"distance"`
}
