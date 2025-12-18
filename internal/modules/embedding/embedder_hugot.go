//go:build ORT || ALL
// +build ORT ALL

package embedding

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/knights-analytics/hugot"
	"github.com/knights-analytics/hugot/pipelines"
)

// HugotEmbedder implements Embedder using the hugot library for local inference.
type HugotEmbedder struct {
	session  *hugot.Session
	pipeline *pipelines.FeatureExtractionPipeline
}

// NewHugotEmbedder creates a new embedder using hugot with the specified model.
// modelName should be a HuggingFace model name like "sentence-transformers/all-MiniLM-L6-v2"
// modelsDir is the directory where models will be downloaded/cached
func NewHugotEmbedder(modelName, modelsDir string) (*HugotEmbedder, error) {
	// Create a new hugot session with ONNX Runtime backend
	session, err := hugot.NewORTSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create hugot ORT session: %w", err)
	}

	// Download the model if not already cached
	modelPath, err := hugot.DownloadModel(modelName, modelsDir, hugot.NewDownloadOptions())
	if err != nil {
		session.Destroy()
		return nil, fmt.Errorf("failed to download model %s: %w", modelName, err)
	}

	// Create feature extraction pipeline configuration
	config := hugot.FeatureExtractionConfig{
		ModelPath: modelPath,
		Name:      filepath.Base(modelPath),
	}

	// Create the feature extraction pipeline
	pipeline, err := hugot.NewPipeline(session, config)
	if err != nil {
		session.Destroy()
		return nil, fmt.Errorf("failed to create feature extraction pipeline: %w", err)
	}

	return &HugotEmbedder{
		session:  session,
		pipeline: pipeline,
	}, nil
}

// Embed generates an embedding for the given text.
func (e *HugotEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	// Run feature extraction on single text
	result, err := e.pipeline.RunPipeline([]string{text})
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	if len(result.Embeddings) == 0 || len(result.Embeddings[0]) == 0 {
		return nil, fmt.Errorf("no embeddings generated")
	}

	return result.Embeddings[0], nil
}

// Close releases resources used by the embedder.
func (e *HugotEmbedder) Close() error {
	if e.session != nil {
		return e.session.Destroy()
	}
	return nil
}
