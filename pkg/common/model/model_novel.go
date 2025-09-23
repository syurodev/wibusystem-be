package model

import (
	"time"

	"github.com/google/uuid"
)

// Novel represents a novel series master table
type Novel struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	Status     string          `json:"status" db:"status"`                         // content_status enum
	CoverImage *string         `json:"cover_image,omitempty" db:"cover_image"`
	Summary    *ContentSummary `json:"summary,omitempty" db:"summary"`             // JSONB field
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at" db:"updated_at"`
}

// NovelVolume represents volumes of a novel
type NovelVolume struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	NovelID            uuid.UUID `json:"novel_id" db:"novel_id"`
	VolumeNumber       int       `json:"volume_number" db:"volume_number"`
	VolumeTitle        *string   `json:"volume_title,omitempty" db:"volume_title"`
	CoverImage         *string   `json:"cover_image,omitempty" db:"cover_image"`
	Description        *string   `json:"description,omitempty" db:"description"`
	PriceCoins         *int      `json:"price_coins,omitempty" db:"price_coins"`
	RentalPriceCoins   *int      `json:"rental_price_coins,omitempty" db:"rental_price_coins"`
	RentalDurationDays *int      `json:"rental_duration_days,omitempty" db:"rental_duration_days"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
}

// NovelChapter represents chapters of a novel
type NovelChapter struct {
	ID            uuid.UUID     `json:"id" db:"id"`
	VolumeID      uuid.UUID     `json:"volume_id" db:"volume_id"`
	ChapterNumber int           `json:"chapter_number" db:"chapter_number"`
	Title         *string       `json:"title,omitempty" db:"title"`
	Content       *NovelContent `json:"content,omitempty" db:"content"`           // JSONB field for rich content
	PublishedAt   *time.Time    `json:"published_at,omitempty" db:"published_at"`
	IsPublic      bool          `json:"is_public" db:"is_public"`
	PriceCoins    *int          `json:"price_coins,omitempty" db:"price_coins"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
}