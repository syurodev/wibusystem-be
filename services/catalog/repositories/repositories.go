// Package repositories contains data access interfaces for the Catalog service.
package repositories

import "github.com/jackc/pgx/v5/pgxpool"

// Repositories aggregates repository interfaces used by handlers.
type Repositories struct {
	Health     HealthRepository
	Genre      GenreRepository
	Character  CharacterRepository
	Creator    CreatorRepository
	Novel      NovelRepository
	NovelQuery NovelQueryRepository // CQRS: Query-side repository for complex reads
	Volume     VolumeRepository     // Volume management repository
	Chapter    ChapterRepository    // Chapter management repository
}

// NewRepositories instantiates concrete repository implementations.
func NewRepositories(pool *pgxpool.Pool) *Repositories {
	return &Repositories{
		Health:     NewHealthRepository(pool),
		Genre:      NewGenreRepository(pool),
		Character:  NewCharacterRepository(pool),
		Creator:    NewCreatorRepository(pool),
		Novel:      NewNovelRepository(pool),
		NovelQuery: NewNovelQueryRepository(pool),
		Volume:     NewVolumeRepository(pool),
		Chapter:    NewChapterRepository(pool),
	}
}
