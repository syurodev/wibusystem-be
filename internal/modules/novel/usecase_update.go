package novel

import (
	"context"
	"encoding/json"

	"system/internal/domain"
)

// updateNovelUseCase implements UpdateNovelUseCase
type updateNovelUseCase struct {
	novelService NovelService
	// TODO: Add external services when relations update is needed
	// genreService  UCGenreService
	// authorService UCAuthorService
	// artistService UCArtistService
}

// NewUpdateNovelUseCase creates a new UpdateNovelUseCase instance
func NewUpdateNovelUseCase(novelService NovelService) UpdateNovelUseCase {
	return &updateNovelUseCase{
		novelService: novelService,
	}
}

// Execute updates an existing novel
// Note: Currently only updates basic novel info. Relations update to be added in future.
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

	return uc.novelService.UpdateNovel(
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

	// TODO: Future implementation for relations update:
	// 1. Begin transaction
	// 2. Update novel entity
	// 3. Remove old relations (genres, authors, artists)
	// 4. Add new relations
	// 5. Commit transaction
}
