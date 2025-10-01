// Package services contains business logic implementations for the Catalog service.
package services

import (
	"wibusystem/services/catalog/grpc"
	"wibusystem/services/catalog/repositories"
	"wibusystem/services/catalog/services/interfaces"
)

// Services aggregates service interfaces used by handlers.
type Services struct {
	Genre     interfaces.GenreServiceInterface
	Character interfaces.CharacterServiceInterface
	Creator   interfaces.CreatorServiceInterface
	Novel     interfaces.NovelServiceInterface
	Volume    interfaces.VolumeServiceInterface
	Chapter   interfaces.ChapterServiceInterface
}

// NewServices instantiates concrete service implementations.
func NewServices(repos *repositories.Repositories, grpcClients *grpc.ClientManager) *Services {
	return &Services{
		Genre:     NewGenreService(repos),
		Character: NewCharacterService(repos),
		Creator:   NewCreatorService(repos),
		Novel:     NewNovelService(repos, grpcClients),
		Volume:    NewVolumeService(repos),
		Chapter:   NewChapterService(repos),
	}
}
