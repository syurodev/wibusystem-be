package interfaces

import (
	"context"

	d "wibusystem/pkg/common/dto"
)

// ChapterServiceInterface defines business logic operations for chapter management.
// This follows the API design spec from /services/catalog/api-design/novel.md sections 3.1-3.7
type ChapterServiceInterface interface {
	// CreateChapter creates a new chapter within a volume.
	// Parameters:
	//   - volumeID: UUID string of the parent volume
	//   - req: Chapter creation request payload
	// Returns the created chapter or an error if creation fails.
	CreateChapter(ctx context.Context, volumeID string, req d.CreateChapterRequest) (*d.CreateChapterResponse, error)

	// GetChapterByID retrieves a specific chapter by its ID.
	// Parameters:
	//   - id: UUID string of the chapter
	//   - includeContent: Flag to include chapter content in the response
	// Returns chapter details or an error if not found.
	GetChapterByID(ctx context.Context, id string, includeContent bool) (*d.ChapterResponse, error)

	// ListChaptersByVolumeID retrieves a paginated list of chapters in a volume.
	// Parameters:
	//   - volumeID: UUID string of the parent volume
	//   - req: List request with pagination and content inclusion options
	// Returns paginated chapter list or an error if the operation fails.
	ListChaptersByVolumeID(ctx context.Context, volumeID string, req d.ListChaptersRequest) (*d.PaginatedChaptersResponse, error)

	// UpdateChapter updates an existing chapter's information.
	// Parameters:
	//   - id: UUID string of the chapter to update
	//   - req: Update request with fields to modify
	// Returns updated chapter information or an error if update fails.
	UpdateChapter(ctx context.Context, id string, req d.UpdateChapterRequest) (*d.UpdateChapterResponse, error)

	// DeleteChapter soft-deletes a chapter.
	// Parameters:
	//   - id: UUID string of the chapter to delete
	// Returns an error if deletion fails (e.g., chapter has purchases).
	DeleteChapter(ctx context.Context, id string) error

	// PublishChapter publishes a chapter, making it publicly accessible.
	// Parameters:
	//   - id: UUID string of the chapter to publish
	//   - req: Publish request with optional publish time
	// Returns updated publish status or an error if operation fails.
	PublishChapter(ctx context.Context, id string, req d.PublishChapterRequest) (*d.PublishChapterResponse, error)

	// UnpublishChapter unpublishes a chapter, removing it from public access.
	// Parameters:
	//   - id: UUID string of the chapter to unpublish
	// Returns updated publish status or an error if operation fails.
	UnpublishChapter(ctx context.Context, id string) (*d.PublishChapterResponse, error)
}