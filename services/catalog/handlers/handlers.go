// Package handlers groups HTTP handlers for the Catalog service.
package handlers

import (
	"wibusystem/pkg/i18n"
	"wibusystem/services/catalog/repositories"
)

// Handlers aggregates all HTTP handlers for dependency injection.
type Handlers struct {
	Health *HealthHandler
}

// NewHandlers wires handlers with their required dependencies.
func NewHandlers(repos *repositories.Repositories, translator *i18n.Translator) *Handlers {
	return &Handlers{
		Health: NewHealthHandler(repos, translator),
	}
}
