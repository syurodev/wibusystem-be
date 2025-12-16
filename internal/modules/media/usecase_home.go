package media

import (
	"context"

	mediadto "system/internal/dto/media"
)

type getHomeDataUseCase struct {
	mediaService MediaService
}

func NewGetHomeDataUseCase(mediaService MediaService) GetHomeDataUseCase {
	return &getHomeDataUseCase{
		mediaService: mediaService,
	}
}

func (uc *getHomeDataUseCase) Execute(ctx context.Context, input GetHomeDataInput) (*mediadto.HomeData, error) {
	return uc.mediaService.GetHomeData(ctx)
}
