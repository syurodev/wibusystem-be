package domain

import (
	"time"

	"github.com/google/uuid"
)

// Character represents a character in a novel.
type Character struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	Traits      []string  `json:"traits"`
	NovelID     uuid.UUID `json:"novel_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewCharacter creates a new Character instance.
func NewCharacter(name, description, imageURL string, novelID uuid.UUID) (*Character, error) {
	id := uuid.New()
	character := &Character{
		ID:          id,
		Name:        name,
		Description: description,
		ImageURL:    imageURL,
		Traits:      []string{},
		NovelID:     novelID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := character.Validate(); err != nil {
		return nil, err
	}

	return character, nil
}

// Validate checks if the Character's fields are valid.
func (c *Character) Validate() error {
	if c.Name == "" {
		return ErrCharacterNameRequired
	}
	if len(c.Name) > 150 {
		return ErrCharacterNameTooLong
	}
	if c.NovelID == uuid.Nil {
		return ErrInvalidNovelID
	}
	return nil
}

// UpdateProfile updates the character's profile information.
func (c *Character) UpdateProfile(name, description, imageURL string) error {
	c.Name = name
	c.Description = description
	c.ImageURL = imageURL
	c.UpdatedAt = time.Now()
	return c.Validate()
}

// AddTrait adds a new trait to the character.
func (c *Character) AddTrait(trait string) {
	if !contains(c.Traits, trait) {
		c.Traits = append(c.Traits, trait)
		c.UpdatedAt = time.Now()
	}
}

// RemoveTrait removes a trait from the character.
func (c *Character) RemoveTrait(trait string) {
	c.Traits = remove(c.Traits, trait)
	c.UpdatedAt = time.Now()
}

// Clone creates a deep copy of the Character.
func (c *Character) Clone() *Character {
	newCharacter := *c
	newCharacter.Traits = make([]string, len(c.Traits))
	copy(newCharacter.Traits, c.Traits)
	return &newCharacter
}
