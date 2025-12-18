package worker

import (
	"context"
	"time"

	"go.uber.org/zap"

	"system/configs"
	"system/internal/modules/embedding"
	"system/internal/modules/novel"
)

// EmbeddingWorker handles background generation of novel embeddings.
// Follows the same pattern as ViewSyncWorker for consistency.
type EmbeddingWorker struct {
	embeddingService *embedding.Service
	novelService     novel.NovelService
	logger           *zap.Logger
	config           *configs.EmbeddingConfig
	stopChan         chan struct{}
	doneChan         chan struct{}
}

// NewEmbeddingWorker creates a new background embedding worker.
func NewEmbeddingWorker(
	embeddingService *embedding.Service,
	novelService novel.NovelService,
	logger *zap.Logger,
	config *configs.EmbeddingConfig,
) *EmbeddingWorker {
	return &EmbeddingWorker{
		embeddingService: embeddingService,
		novelService:     novelService,
		logger:           logger,
		config:           config,
		stopChan:         make(chan struct{}),
		doneChan:         make(chan struct{}),
	}
}

// Start begins the background worker goroutine.
func (w *EmbeddingWorker) Start(ctx context.Context) {
	if !w.config.WorkerEnabled {
		w.logger.Info("Embedding worker is disabled by configuration")
		close(w.doneChan)
		return
	}

	w.logger.Info("Starting embedding worker",
		zap.Int("interval_minutes", w.config.SyncIntervalMins),
		zap.Int("batch_size", w.config.BatchSize))

	go w.run(ctx)
}

// Stop gracefully stops the worker.
func (w *EmbeddingWorker) Stop(ctx context.Context) error {
	w.logger.Info("Stopping embedding worker...")

	close(w.stopChan)

	select {
	case <-w.doneChan:
		w.logger.Info("Embedding worker stopped gracefully")
		return nil
	case <-ctx.Done():
		w.logger.Warn("Embedding worker stop timed out")
		return ctx.Err()
	}
}

// run is the main worker loop.
func (w *EmbeddingWorker) run(ctx context.Context) {
	defer close(w.doneChan)

	ticker := time.NewTicker(time.Duration(w.config.SyncIntervalMins) * time.Minute)
	defer ticker.Stop()

	// Run immediately on start
	w.processQueue(ctx)

	for {
		select {
		case <-ticker.C:
			w.processQueue(ctx)

		case <-w.stopChan:
			w.logger.Info("Worker received stop signal, performing final sync...")
			w.processQueue(ctx)
			return

		case <-ctx.Done():
			w.logger.Warn("Worker context cancelled")
			return
		}
	}
}

// processQueue processes pending novel embeddings from the Redis queue.
func (w *EmbeddingWorker) processQueue(ctx context.Context) {
	w.logger.Debug("Embedding sync cycle started")
	startTime := time.Now()

	// Create timeout context
	syncCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Pop pending novel IDs from Redis
	novelIDs, err := w.embeddingService.PopPendingNovelIDs(syncCtx, w.config.BatchSize)
	if err != nil {
		w.logger.Error("Failed to pop pending novel IDs", zap.Error(err))
		return
	}

	if len(novelIDs) == 0 {
		w.logger.Debug("No pending novels to embed")
		return
	}

	w.logger.Info("Processing pending embeddings", zap.Int("count", len(novelIDs)))

	successCount := 0
	errorCount := 0

	for _, novelID := range novelIDs {
		// Fetch novel data
		novelData, err := w.novelService.GetNovelByID(syncCtx, novelID)
		if err != nil {
			w.logger.Warn("Failed to fetch novel for embedding",
				zap.String("novel_id", novelID.String()),
				zap.Error(err))
			errorCount++
			continue
		}

		// Get genre names
		var genreNames []string
		if novelData.Genres != nil {
			for _, g := range novelData.Genres {
				genreNames = append(genreNames, g.Name)
			}
		}

		// Extract synopsis text (simplified - assumes synopsis is text)
		synopsisText := ""
		if novelData.Synopsis != nil {
			synopsisText = string(novelData.Synopsis)
		}

		// Build embedding input
		input := &embedding.NovelEmbeddingInput{
			NovelID:       novelID,
			Title:         novelData.Title,
			OriginalTitle: novelData.OriginalTitle,
			Synopsis:      synopsisText,
			Genres:        genreNames,
		}

		// Generate and store embedding
		if err := w.embeddingService.GenerateAndStoreEmbedding(syncCtx, input); err != nil {
			w.logger.Error("Failed to generate embedding",
				zap.String("novel_id", novelID.String()),
				zap.Error(err))
			errorCount++
			continue
		}

		successCount++
	}

	duration := time.Since(startTime)
	w.logger.Info("Embedding sync cycle completed",
		zap.Duration("duration", duration),
		zap.Int("success", successCount),
		zap.Int("errors", errorCount))
}
