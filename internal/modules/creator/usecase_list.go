package creator

import (
	"context"

	"system/internal/domain"
)

type listCreatorsUseCase struct {
	creatorService CreatorService
}

func NewListCreatorsUseCase(creatorService CreatorService) ListCreatorsUseCase {
	return &listCreatorsUseCase{
		creatorService: creatorService,
	}
}

func (uc *listCreatorsUseCase) Execute(ctx context.Context, input ListCreatorsInput) (*domain.CreatorListResult, error) {
	return uc.creatorService.ListCreators(ctx, input.Filter)
}
