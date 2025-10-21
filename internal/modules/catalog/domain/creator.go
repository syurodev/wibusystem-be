package domain

import (
	"time"

	"github.com/google/uuid"
)

// Creator represents an author or artist.
type Creator struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Bio          string    `json:"bio"`
	ProfileImage string    `json:"profile_image"`
	SocialLinks  []string  `json:"social_links"`
	NovelCount   int       `json:"novel_count"`
	Followers    int       `json:"followers"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// NewCreator creates a new Creator instance.
func NewCreator(name, bio, profileImage string) (*Creator, error) {
	id := uuid.New()
	creator := &Creator{
		ID:           id,
		Name:         name,
		Bio:          bio,
		ProfileImage: profileImage,
		SocialLinks:  []string{},
		NovelCount:   0,
		Followers:    0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := creator.Validate(); err != nil {
		return nil, err
	}

	return creator, nil
}

// Validate checks if the Creator's fields are valid.
func (c *Creator) Validate() error {
	if c.Name == "" {
		return ErrCreatorNameRequired
	}
	if len(c.Name) > 150 {
		return ErrCreatorNameTooLong
	}
	return nil
}

// UpdateProfile updates the creator's profile information.
func (c *Creator) UpdateProfile(name, bio, profileImage string) error {
	c.Name = name
	c.Bio = bio
	c.ProfileImage = profileImage
	c.UpdatedAt = time.Now()
	return c.Validate()
}

// AddSocialLink adds a new social media link.
func (c *Creator) AddSocialLink(link string) {
	if !contains(c.SocialLinks, link) {
		c.SocialLinks = append(c.SocialLinks, link)
		c.UpdatedAt = time.Now()
	}
}

// RemoveSocialLink removes a social media link.
func (c *Creator) RemoveSocialLink(link string) {
	c.SocialLinks = remove(c.SocialLinks, link)
	c.UpdatedAt = time.Now()
}

// IncrementFollowers increases the follower count.
func (c *Creator) IncrementFollowers() {
	c.Followers++
	c.UpdatedAt = time.Now()
}

// DecrementFollowers decreases the follower count.
func (c *Creator) DecrementFollowers() {
	if c.Followers > 0 {
		c.Followers--
		c.UpdatedAt = time.Now()
	}
}

// IncrementNovelCount increases the novel count.
func (c *Creator) IncrementNovelCount() {
	c.NovelCount++
	c.UpdatedAt = time.Now()
}

// DecrementNovelCount decreases the novel count.
func (c *Creator) DecrementNovelCount() {
	if c.NovelCount > 0 {
		c.NovelCount--
		c.UpdatedAt = time.Now()
	}
}

// Clone creates a deep copy of the Creator.
func (c *Creator) Clone() *Creator {
	newCreator := *c
	newCreator.SocialLinks = make([]string, len(c.SocialLinks))
	copy(newCreator.SocialLinks, c.SocialLinks)
	return &newCreator
}

// contains checks if a string is present in a slice.
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

// remove removes a string from a slice.
func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
