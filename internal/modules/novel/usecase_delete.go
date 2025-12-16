package novel

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"

	db "system/internal/platform/database"
	pkgerrors "system/pkg/errors"
)

// deleteNovelUseCase implements DeleteNovelUseCase
type deleteNovelUseCase struct {
	txManager      db.TransactionManager
	novelService   UCNovelService
	genreService   UCGenreService
	creatorService UCCreatorService
}

// NewDeleteNovelUseCase creates a new DeleteNovelUseCase instance
func NewDeleteNovelUseCase(
	txManager db.TransactionManager,
	novelService UCNovelService,
	genreService UCGenreService,
	creatorService UCCreatorService,
) DeleteNovelUseCase {
	return &deleteNovelUseCase{
		txManager:      txManager,
		novelService:   novelService,
		genreService:   genreService,
		creatorService: creatorService,
	}
}

// Execute deletes a novel with transaction
func (uc *deleteNovelUseCase) Execute(ctx context.Context, input DeleteNovelInput) error {
	return uc.txManager.RunInTx(ctx, func(txCtx context.Context) error {
		// 1. Get novel to check existence and owner
		novel, err := uc.novelService.GetNovelByID(txCtx, input.ID)
		if err != nil {
			return err
		}
		if novel == nil {
			return pkgerrors.NotFound(I18nNotFound, "novel not found")
		}

		// 2. Get genres to decrement counts
		genres, err := uc.genreService.GetNovelGenres(txCtx, input.ID)
		if err != nil {
			// Log error but continue? Or fail? The previous service logic just logged it.
			// In transaction, we should probably fail if critical?
			// However, previous service logic: "Failed to get novel genres... if err != nil print"
			// But then "if len(genres) > 0" it decrements.
			// Safe approach: fail if error, to ensure data consistency.
			return fmt.Errorf("failed to get novel genres: %w", err)
		}

		// 3. Delete novel entity (soft delete)
		if err := uc.novelService.DeleteNovelEntity(txCtx, input.ID); err != nil {
			return err
		}

		// 4. Decrement genre counts
		if len(genres) > 0 {
			increments := make(map[uuid.UUID]int)
			for _, g := range genres {
				increments[g.ID] = -1
			}
			if err := uc.genreService.BatchIncrementNovelCount(txCtx, increments); err != nil {
				return fmt.Errorf("failed to decrement genre counts: %w", err)
			}
		}

		// 5. Decrement user novel count
		if novel.OwnerType == "user" {
			if err := uc.creatorService.DecrementNovelCount(txCtx, novel.OwnerID); err != nil {
				return err
			}
		}

		return nil
	})
}
