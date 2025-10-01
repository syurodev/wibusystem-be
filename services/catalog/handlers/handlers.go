// Package handlers groups HTTP handlers for the Catalog service.
package handlers

import (
	"wibusystem/pkg/i18n"
	"wibusystem/services/catalog/repositories"
	"wibusystem/services/catalog/services"
)

// Handlers aggregates all HTTP handlers for dependency injection.
type Handlers struct {
	Health    *HealthHandler
	Genre     *GenreHandler
	Character *CharacterHandler
	Creator   *CreatorHandler
	Novel     *NovelHandler
	Volume    *VolumeHandler
	Chapter   *ChapterHandler
}

// NewHandlers wires handlers with their required dependencies.
func NewHandlers(repos *repositories.Repositories, services *services.Services, translator *i18n.Translator) *Handlers {
	return &Handlers{
		Health:    NewHealthHandler(repos, translator),
		Genre:     NewGenreHandler(services.Genre, translator),
		Character: NewCharacterHandler(services.Character, translator),
		Creator:   NewCreatorHandler(services.Creator, translator),
		Novel:     NewNovelHandler(services.Novel, translator),
		Volume:    NewVolumeHandler(services.Volume, translator),
		Chapter:   NewChapterHandler(services.Chapter, translator),
	}
}
