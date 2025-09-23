// Package services contains business logic implementations for the Catalog service.
package services

import (
	"wibusystem/services/catalog/repositories"
	"wibusystem/services/catalog/services/interfaces"
)

// Services aggregates service interfaces used by handlers.
type Services struct {
	Genre     interfaces.GenreServiceInterface
	Character interfaces.CharacterServiceInterface
	Creator   interfaces.CreatorServiceInterface
}

// NewServices instantiates concrete service implementations.
func NewServices(repos *repositories.Repositories) *Services {
	return &Services{
		Genre:     NewGenreService(repos),
		Character: NewCharacterService(repos),
		Creator:   NewCreatorService(repos),
	}
}