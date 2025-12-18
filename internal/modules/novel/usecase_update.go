package novel

import (
	"context"
	"encoding/json"

	"system/internal/domain"
)

// updateNovelUseCase implements UpdateNovelUseCase
type updateNovelUseCase struct {
	novelService     NovelService
	embeddingService UCEmbeddingService
}

// NewUpdateNovelUseCase creates a new UpdateNovelUseCase instance
func NewUpdateNovelUseCase(novelService NovelService, embeddingService UCEmbeddingService) UpdateNovelUseCase {
	return &updateNovelUseCase{
		novelService:     novelService,
		embeddingService: embeddingService,
	}
}

// Execute updates an existing novel
func (uc *updateNovelUseCase) Execute(ctx context.Context, input UpdateNovelInput) (*domain.Novel, error) {
	// Convert input to service call
	var status *string
	if input.Status != "" {
		status = &input.Status
	}

	var metadataJSON *string
	if input.MetadataJSON != nil {
		metadataJSON = input.MetadataJSON
	}

	// Validate synopsis
	var synopsis json.RawMessage
	if len(input.Synopsis) > 0 {
		synopsis = input.Synopsis
	}

	result, err := uc.novelService.UpdateNovel(
		ctx,
		input.ID,
		input.Title,
		synopsis,
		input.CoverImageURL,
		input.ThumbnailURL,
		status,
		input.OriginalLanguage,
		input.OriginalTitle,
		metadataJSON,
		input.IsOneshot,
	)

	if err != nil {
		return nil, err
	}

	// Queue for re-embedding (content changed)
	if uc.embeddingService != nil {
		_ = uc.embeddingService.QueueNovelForEmbedding(ctx, result.ID)
	}

	return result, nil
}
