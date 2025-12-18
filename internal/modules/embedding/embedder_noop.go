package embedding

import (
	"context"
	"fmt"
	"math/rand"
)

// NoopEmbedder is a placeholder embedder for development/testing.
// It generates random vectors instead of actual embeddings.
// Use HugotEmbedder (with ORT build tag) for production.
type NoopEmbedder struct {
	dimensions int
}

// NewNoopEmbedder creates a new noop embedder.
func NewNoopEmbedder(dimensions int) *NoopEmbedder {
	return &NoopEmbedder{dimensions: dimensions}
}

// Embed returns a random vector (for testing purposes).
func (e *NoopEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("empty text")
	}

	// Generate deterministic random vector based on text hash
	// This ensures same text gives same embedding (for testing)
	seed := int64(0)
	for _, c := range text {
		seed += int64(c)
	}
	r := rand.New(rand.NewSource(seed))

	vector := make([]float32, e.dimensions)
	for i := range vector {
		vector[i] = r.Float32()*2 - 1 // Random value between -1 and 1
	}

	return vector, nil
}

// Close is a no-op for this embedder.
func (e *NoopEmbedder) Close() error {
	return nil
}
