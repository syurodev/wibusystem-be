// Package repositories contains data access interfaces for the Catalog service.
package repositories

import "github.com/jackc/pgx/v5/pgxpool"

// Repositories aggregates repository interfaces used by handlers.
type Repositories struct {
	Health    HealthRepository
	Genre     GenreRepository
	Character CharacterRepository
	Creator   CreatorRepository
}

// NewRepositories instantiates concrete repository implementations.
func NewRepositories(pool *pgxpool.Pool) *Repositories {
	return &Repositories{
		Health:    NewHealthRepository(pool),
		Genre:     NewGenreRepository(pool),
		Character: NewCharacterRepository(pool),
		Creator:   NewCreatorRepository(pool),
	}
}
