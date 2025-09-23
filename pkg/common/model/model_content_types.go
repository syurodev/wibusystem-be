package model

import (
	"database/sql/driver"
	"encoding/json"
)

// ContentSummary represents structured summary information stored as JSONB
// Used for anime, manga, and novel summary fields
type ContentSummary struct {
	Text        string            `json:"text,omitempty"`        // Main summary text
	Synopsis    string            `json:"synopsis,omitempty"`    // Brief synopsis
	Tags        []string          `json:"tags,omitempty"`        // Content tags
	Rating      string            `json:"rating,omitempty"`      // Age rating (PG, R, etc.)
	Duration    string            `json:"duration,omitempty"`    // Duration info (for anime)
	Chapters    int               `json:"chapters,omitempty"`    // Total chapters (for manga/novel)
	Volumes     int               `json:"volumes,omitempty"`     // Total volumes (for manga/novel)
	Episodes    int               `json:"episodes,omitempty"`    // Total episodes (for anime)
	Metadata    map[string]string `json:"metadata,omitempty"`    // Additional metadata
}

// Value implements driver.Valuer interface for database storage
func (cs ContentSummary) Value() (driver.Value, error) {
	return json.Marshal(cs)
}

// Scan implements sql.Scanner interface for database retrieval
func (cs *ContentSummary) Scan(value interface{}) error {
	if value == nil {
		*cs = ContentSummary{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), cs)
	}
	return json.Unmarshal(bytes, cs)
}

// NovelContent represents structured content for novel chapters stored as JSONB
type NovelContent struct {
	Text        string                   `json:"text,omitempty"`        // Main content text
	Formatting  NovelContentFormatting   `json:"formatting,omitempty"`  // Text formatting info
	Sections    []NovelContentSection    `json:"sections,omitempty"`    // Content sections
	Images      []NovelContentImage      `json:"images,omitempty"`      // Embedded images
	Footnotes   []NovelContentFootnote   `json:"footnotes,omitempty"`   // Footnotes
	Metadata    map[string]string        `json:"metadata,omitempty"`    // Additional metadata
}

// NovelContentFormatting represents formatting options for novel content
type NovelContentFormatting struct {
	FontFamily string `json:"font_family,omitempty"` // Font family
	FontSize   string `json:"font_size,omitempty"`   // Font size
	LineHeight string `json:"line_height,omitempty"` // Line height
	Alignment  string `json:"alignment,omitempty"`   // Text alignment
}

// NovelContentSection represents a section within novel content
type NovelContentSection struct {
	Title   string `json:"title,omitempty"`   // Section title
	Content string `json:"content"`           // Section content
	Type    string `json:"type,omitempty"`    // Section type (chapter, prologue, epilogue, etc.)
	Order   int    `json:"order"`             // Order within chapter
}

// NovelContentImage represents an image within novel content
type NovelContentImage struct {
	URL         string `json:"url"`                    // Image URL
	Alt         string `json:"alt,omitempty"`          // Alt text
	Caption     string `json:"caption,omitempty"`      // Image caption
	Position    string `json:"position,omitempty"`     // Position (top, middle, bottom, inline)
	Width       string `json:"width,omitempty"`        // Image width
	Height      string `json:"height,omitempty"`       // Image height
}

// NovelContentFootnote represents a footnote within novel content
type NovelContentFootnote struct {
	ID      string `json:"id"`                // Footnote ID
	Text    string `json:"text"`              // Footnote text
	Anchor  string `json:"anchor,omitempty"`  // Anchor text in main content
}

// Value implements driver.Valuer interface for database storage
func (nc NovelContent) Value() (driver.Value, error) {
	return json.Marshal(nc)
}

// Scan implements sql.Scanner interface for database retrieval
func (nc *NovelContent) Scan(value interface{}) error {
	if value == nil {
		*nc = NovelContent{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return json.Unmarshal([]byte(value.(string)), nc)
	}
	return json.Unmarshal(bytes, nc)
}