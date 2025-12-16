package creator

import (
	"context"

	"system/internal/domain"
)

type ListCreatorsInput struct {
	Filter domain.CreatorListFilter
}

type ListCreatorsUseCase interface {
	Execute(ctx context.Context, input ListCreatorsInput) (*domain.CreatorListResult, error)
}
