package domain

import (
	"time"

	"github.com/google/uuid"
)

// Genre represents a category for novels.
type Genre struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Slug        string     `json:"slug"`
	ParentID    *uuid.UUID `json:"parent_id"`
	NovelCount  int        `json:"novel_count"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// NewGenre creates a new Genre instance.
func NewGenre(name, description string) (*Genre, error) {
	id := uuid.New()
	slug, err := GenerateSlug(name)
	if err != nil {
		return nil, err
	}

	genre := &Genre{
		ID:          id,
		Name:        name,
		Description: description,
		Slug:        slug,
		NovelCount:  0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := genre.Validate(); err != nil {
		return nil, err
	}

	return genre, nil
}

// Validate checks if the Genre's fields are valid.
func (g *Genre) Validate() error {
	if g.Name == "" {
		return ErrGenreNameRequired
	}
	if len(g.Name) > 100 {
		return ErrGenreNameTooLong
	}
	if g.Slug == "" {
		return ErrInvalidSlug
	}
	return nil
}

// UpdateName updates the genre's name and regenerates the slug.
func (g *Genre) UpdateName(name, description string) error {
	newSlug, err := GenerateSlug(name)
	if err != nil {
		return err
	}
	g.Name = name
	g.Description = description
	g.Slug = newSlug
	g.UpdatedAt = time.Now()
	return g.Validate()
}

// SetParent sets the parent genre.
func (g *Genre) SetParent(parentID *uuid.UUID) {
	g.ParentID = parentID
	g.UpdatedAt = time.Now()
}

// IncrementNovelCount increases the novel count for this genre.
func (g *Genre) IncrementNovelCount() {
	g.NovelCount++
	g.UpdatedAt = time.Now()
}

// DecrementNovelCount decreases the novel count for this genre.
func (g *Genre) DecrementNovelCount() {
	if g.NovelCount > 0 {
		g.NovelCount--
		g.UpdatedAt = time.Now()
	}
}

// Clone creates a deep copy of the Genre.
func (g *Genre) Clone() *Genre {
	newGenre := *g
	if g.ParentID != nil {
		newParentID := *g.ParentID
		newGenre.ParentID = &newParentID
	}
	return &newGenre
}
