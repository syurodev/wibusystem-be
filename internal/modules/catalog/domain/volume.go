// Package domain contains business entities for the Catalog module.
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Volume represents a volume in a novel series
type Volume struct {
	// Core identification
	ID           uuid.UUID `json:"id"`
	NovelID      uuid.UUID `json:"novel_id"`
	VolumeNumber int       `json:"volume_number"`

	// Content fields
	VolumeTitle string `json:"volume_title,omitempty"`
	CoverImage  string `json:"cover_image,omitempty"`
	Description string `json:"description,omitempty"`

	// User tracking
	LastModifiedByUserID *uuid.UUID `json:"last_modified_by_user_id,omitempty"`

	// Publishing and status
	PublishedAt *time.Time `json:"published_at,omitempty"`
	IsDeleted   bool       `json:"is_deleted"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	IsAvailable bool       `json:"is_available"`

	// Pricing
	PriceCoins         int `json:"price_coins,omitempty"`
	RentalPriceCoins   int `json:"rental_price_coins,omitempty"`
	RentalDurationDays int `json:"rental_duration_days,omitempty"`

	// Content metadata
	PageCount            int `json:"page_count,omitempty"`
	WordCount            int `json:"word_count,omitempty"`
	ChapterCount         int `json:"chapter_count"`
	EstimatedReadingTime int `json:"estimated_reading_time,omitempty"`

	// Audit timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Chapter represents a chapter in a volume
type Chapter struct {
	// Core identification
	ID            uuid.UUID `json:"id"`
	VolumeID      uuid.UUID `json:"volume_id"`
	ChapterNumber int       `json:"chapter_number"`

	// Content fields
	Title   string          `json:"title,omitempty"`
	Content json.RawMessage `json:"content,omitempty"`

	// User tracking
	LastModifiedByUserID *uuid.UUID `json:"last_modified_by_user_id,omitempty"`

	// Publishing workflow
	PublishedAt        *time.Time `json:"published_at,omitempty"`
	ScheduledPublishAt *time.Time `json:"scheduled_publish_at,omitempty"`
	IsDraft            bool       `json:"is_draft"`
	IsPublic           bool       `json:"is_public"`
	IsDeleted          bool       `json:"is_deleted"`
	DeletedAt          *time.Time `json:"deleted_at,omitempty"`
	Version            int        `json:"version"`

	// Content warnings
	ContentWarnings  json.RawMessage `json:"content_warnings,omitempty"`
	HasMatureContent bool            `json:"has_mature_content"`

	// Pricing
	PriceCoins int `json:"price_coins,omitempty"`

	// Content metadata
	WordCount          int `json:"word_count,omitempty"`
	CharacterCount     int `json:"character_count,omitempty"`
	ReadingTimeMinutes int `json:"reading_time_minutes,omitempty"`

	// Analytics
	ViewCount    int64 `json:"view_count"`
	LikeCount    int64 `json:"like_count"`
	CommentCount int64 `json:"comment_count"`

	// Audit timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewVolume creates a new volume entity
func NewVolume(novelID uuid.UUID, volumeNumber int) (*Volume, error) {
	if novelID == uuid.Nil {
		return nil, ErrInvalidVolumeNovelID
	}
	if volumeNumber < 1 {
		return nil, ErrInvalidVolumeNumber
	}

	now := time.Now()
	volume := &Volume{
		ID:           uuid.New(),
		NovelID:      novelID,
		VolumeNumber: volumeNumber,
		IsAvailable:  false,
		IsDeleted:    false,
		ChapterCount: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return volume, nil
}

// Validate validates the volume entity
func (v *Volume) Validate() error {
	if v.ID == uuid.Nil {
		return ErrInvalidVolumeID
	}
	if v.NovelID == uuid.Nil {
		return ErrInvalidVolumeNovelID
	}
	if v.VolumeNumber < 1 {
		return ErrInvalidVolumeNumber
	}
	if len(v.VolumeTitle) > 500 {
		return ErrVolumeTitleTooLong
	}
	if v.PriceCoins < 0 || v.RentalPriceCoins < 0 || v.RentalDurationDays < 0 {
		return ErrNegativePrice
	}
	return nil
}

// UpdateTitle updates the volume title
func (v *Volume) UpdateTitle(title string) error {
	if v.IsDeleted {
		return ErrCannotModifyDeleted
	}
	if len(title) > 500 {
		return ErrVolumeTitleTooLong
	}
	v.VolumeTitle = title
	v.UpdatedAt = time.Now()
	return nil
}

// Publish marks the volume as published
func (v *Volume) Publish() error {
	if v.IsDeleted {
		return ErrCannotModifyDeleted
	}
	now := time.Now()
	v.PublishedAt = &now
	v.IsAvailable = true
	v.UpdatedAt = now
	return nil
}

// Delete soft deletes the volume
func (v *Volume) Delete() error {
	if v.IsDeleted {
		return ErrVolumeDeleted
	}
	now := time.Now()
	v.IsDeleted = true
	v.DeletedAt = &now
	v.UpdatedAt = now
	return nil
}

// UpdateMetadata updates content metadata
func (v *Volume) UpdateMetadata(chapterCount, pageCount, wordCount, readingTime int) error {
	if v.IsDeleted {
		return ErrCannotModifyDeleted
	}
	if chapterCount < 0 || pageCount < 0 || wordCount < 0 || readingTime < 0 {
		return errors.New("metadata values cannot be negative")
	}
	v.ChapterCount = chapterCount
	v.PageCount = pageCount
	v.WordCount = wordCount
	v.EstimatedReadingTime = readingTime
	v.UpdatedAt = time.Now()
	return nil
}

// Clone creates a deep copy of the volume
func (v *Volume) Clone() *Volume {
	clone := *v
	if v.PublishedAt != nil {
		t := *v.PublishedAt
		clone.PublishedAt = &t
	}
	if v.DeletedAt != nil {
		t := *v.DeletedAt
		clone.DeletedAt = &t
	}
	if v.LastModifiedByUserID != nil {
		id := *v.LastModifiedByUserID
		clone.LastModifiedByUserID = &id
	}
	return &clone
}

// NewChapter creates a new chapter entity
func NewChapter(volumeID uuid.UUID, chapterNumber int) (*Chapter, error) {
	if volumeID == uuid.Nil {
		return nil, ErrInvalidChapterVolumeID
	}
	if chapterNumber < 1 {
		return nil, ErrInvalidChapterNumber
	}

	now := time.Now()
	chapter := &Chapter{
		ID:               uuid.New(),
		VolumeID:         volumeID,
		ChapterNumber:    chapterNumber,
		IsDraft:          true,
		IsPublic:         false,
		IsDeleted:        false,
		Version:          1,
		HasMatureContent: false,
		ViewCount:        0,
		LikeCount:        0,
		CommentCount:     0,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	return chapter, nil
}

// Validate validates the chapter entity
func (c *Chapter) Validate() error {
	if c.ID == uuid.Nil {
		return ErrInvalidChapterID
	}
	if c.VolumeID == uuid.Nil {
		return ErrInvalidChapterVolumeID
	}
	if c.ChapterNumber < 1 {
		return ErrInvalidChapterNumber
	}
	if len(c.Title) > 500 {
		return ErrChapterTitleTooLong
	}
	if c.Version < 1 {
		return ErrInvalidChapterVersion
	}
	if c.PriceCoins < 0 {
		return ErrNegativePrice
	}
	return nil
}

// UpdateTitle updates the chapter title
func (c *Chapter) UpdateTitle(title string) error {
	if c.IsDeleted {
		return ErrCannotModifyDeleted
	}
	if len(title) > 500 {
		return ErrChapterTitleTooLong
	}
	c.Title = title
	c.UpdatedAt = time.Now()
	return nil
}

// UpdateContent updates the chapter content
func (c *Chapter) UpdateContent(content json.RawMessage) error {
	if c.IsDeleted {
		return ErrCannotModifyDeleted
	}
	c.Content = content
	c.Version++
	c.UpdatedAt = time.Now()
	return nil
}

// Publish publishes the chapter
func (c *Chapter) Publish() error {
	if c.IsDeleted {
		return ErrCannotModifyDeleted
	}
	now := time.Now()
	c.PublishedAt = &now
	c.IsPublic = true
	c.IsDraft = false
	c.UpdatedAt = now
	return nil
}

// SchedulePublish schedules the chapter for publishing
func (c *Chapter) SchedulePublish(scheduleAt time.Time) error {
	if c.IsDeleted {
		return ErrCannotModifyDeleted
	}
	c.ScheduledPublishAt = &scheduleAt
	c.UpdatedAt = time.Now()
	return nil
}

// Delete soft deletes the chapter
func (c *Chapter) Delete() error {
	if c.IsDeleted {
		return ErrChapterDeleted
	}
	now := time.Now()
	c.IsDeleted = true
	c.DeletedAt = &now
	c.UpdatedAt = now
	return nil
}

// IncrementViewCount increments the view count
func (c *Chapter) IncrementViewCount() {
	c.ViewCount++
	c.UpdatedAt = time.Now()
}

// IncrementLikeCount increments the like count
func (c *Chapter) IncrementLikeCount() {
	c.LikeCount++
	c.UpdatedAt = time.Now()
}

// DecrementLikeCount decrements the like count
func (c *Chapter) DecrementLikeCount() {
	if c.LikeCount > 0 {
		c.LikeCount--
		c.UpdatedAt = time.Now()
	}
}

// UpdateMetadata updates content metadata
func (c *Chapter) UpdateMetadata(wordCount, charCount, readingTime int) error {
	if c.IsDeleted {
		return ErrCannotModifyDeleted
	}
	if wordCount < 0 || charCount < 0 || readingTime < 0 {
		return errors.New("metadata values cannot be negative")
	}
	c.WordCount = wordCount
	c.CharacterCount = charCount
	c.ReadingTimeMinutes = readingTime
	c.UpdatedAt = time.Now()
	return nil
}

// IsPublished checks if the chapter is published
func (c *Chapter) IsPublished() bool {
	return c.PublishedAt != nil && !c.IsDraft && c.IsPublic
}

// IsScheduled checks if the chapter is scheduled for publishing
func (c *Chapter) IsScheduled() bool {
	if c.ScheduledPublishAt == nil {
		return false
	}
	return c.ScheduledPublishAt.After(time.Now())
}

// Clone creates a deep copy of the chapter
func (c *Chapter) Clone() *Chapter {
	clone := *c
	if c.PublishedAt != nil {
		t := *c.PublishedAt
		clone.PublishedAt = &t
	}
	if c.ScheduledPublishAt != nil {
		t := *c.ScheduledPublishAt
		clone.ScheduledPublishAt = &t
	}
	if c.DeletedAt != nil {
		t := *c.DeletedAt
		clone.DeletedAt = &t
	}
	if c.LastModifiedByUserID != nil {
		id := *c.LastModifiedByUserID
		clone.LastModifiedByUserID = &id
	}
	return &clone
}

// String returns a string representation of the chapter
func (c *Chapter) String() string {
	return fmt.Sprintf("Chapter{ID: %s, Number: %d, Title: %s, IsDraft: %v}",
		c.ID, c.ChapterNumber, c.Title, c.IsDraft)
}
