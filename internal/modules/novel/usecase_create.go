package novel

import (
	"context"
	"encoding/json"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
	db "system/internal/platform/database"
	pkgerrors "system/pkg/errors"
	"system/pkg/util/stringutil"
)

// createNovelUseCase orchestrates novel creation across multiple services
// It owns the transaction lifecycle
type createNovelUseCase struct {
	txManager        db.TransactionManager
	novelService     UCNovelService
	genreService     UCGenreService
	authorService    UCAuthorService
	artistService    UCArtistService
	creatorService   UCCreatorService
	embeddingService UCEmbeddingService
}

// NewCreateNovelUseCase creates a new CreateNovelUseCase instance
func NewCreateNovelUseCase(
	txManager db.TransactionManager,
	novelService UCNovelService,
	genreService UCGenreService,
	authorService UCAuthorService,
	artistService UCArtistService,
	creatorService UCCreatorService,
	embeddingService UCEmbeddingService,
) CreateNovelUseCase {
	return &createNovelUseCase{
		txManager:        txManager,
		novelService:     novelService,
		genreService:     genreService,
		authorService:    authorService,
		artistService:    artistService,
		creatorService:   creatorService,
		embeddingService: embeddingService,
	}
}

// Execute creates a novel with all its relations in a single transaction
func (uc *createNovelUseCase) Execute(ctx context.Context, input CreateNovelInput) (*domain.Novel, error) {
	// Validate input
	if input.Title == "" {
		return nil, pkgerrors.BadRequest(I18nInvalidInput, "title is required")
	}
	if !domain.NovelStatus(input.Status).IsValid() {
		return nil, pkgerrors.BadRequest(I18nInvalidStatus, "invalid novel status")
	}

	// Generate unique slug
	novelSlug, err := stringutil.GenerateUniqueSlug(input.Title)
	if err != nil {
		return nil, err
	}

	// Validate synopsis JSON
	synopsis := input.Synopsis
	if len(synopsis) == 0 || string(synopsis) == "null" {
		synopsis = json.RawMessage("{}")
	} else if !json.Valid(synopsis) {
		return nil, pkgerrors.BadRequest(I18nInvalidInput, "invalid synopsis JSON")
	}

	// Validate metadata JSON
	var metadata json.RawMessage
	if input.MetadataJSON != nil && *input.MetadataJSON != "" {
		if !json.Valid([]byte(*input.MetadataJSON)) {
			return nil, pkgerrors.BadRequest(I18nInvalidInput, "invalid metadata JSON")
		}
		metadata = json.RawMessage(*input.MetadataJSON)
	} else {
		metadata = json.RawMessage("{}")
	}

	// Generate ID
	id, err := uuid.NewV7()
	if err != nil {
		return nil, pkgerrors.Internal(I18nCreateFailed, "failed to generate ID")
	}

	// Build novel entity
	novel := &domain.Novel{
		ID:               id,
		Title:            input.Title,
		Slug:             novelSlug,
		Synopsis:         synopsis,
		CoverImageURL:    input.CoverImageURL,
		ThumbnailURL:     input.ThumbnailURL,
		Status:           domain.NovelStatus(input.Status),
		IsOneshot:        input.IsOneshot,
		OriginalLanguage: input.OriginalLanguage,
		OriginalTitle:    input.OriginalTitle,
		Metadata:         metadata,
		OwnerID:          input.OwnerID,
		OwnerType:        input.OwnerType,
		CreatedBy:        input.CreatedBy,
	}

	var result *domain.Novel

	// Execute all operations in a single transaction
	err = uc.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		// 1. Create novel
		if err := uc.novelService.CreateNovelEntity(txCtx, novel); err != nil {
			return err
		}

		// 2. Add genres
		if len(input.GenreIDs) > 0 {
			if err := uc.genreService.AddNovelGenres(txCtx, id, input.GenreIDs, input.OwnerID); err != nil {
				return err
			}
		}

		// 3. Add authors
		if len(input.AuthorIDs) > 0 {
			if err := uc.authorService.AddNovelAuthors(txCtx, id, input.AuthorIDs); err != nil {
				return err
			}
		}

		// 4. Add artists
		if len(input.ArtistIDs) > 0 {
			if err := uc.artistService.AddNovelArtists(txCtx, id, input.ArtistIDs); err != nil {
				return err
			}
		}

		// 5. Increment user novel count
		if input.OwnerType == "user" {
			if err := uc.creatorService.IncrementNovelCount(txCtx, input.OwnerID); err != nil {
				return err
			}
		}

		// 6. Get created novel
		result, err = uc.novelService.GetNovelByID(txCtx, id)
		return err
	})

	if err != nil {
		return nil, err
	}

	// Queue for embedding generation (fire and forget, don't block response)
	if uc.embeddingService != nil {
		_ = uc.embeddingService.QueueNovelForEmbedding(ctx, result.ID)
	}

	return result, nil
}
