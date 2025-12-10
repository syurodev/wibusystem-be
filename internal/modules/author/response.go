package author

import (
	dtoauthor "system/internal/dto/author"
)

// Re-export types from dto package for backward compatibility
type AuthorResponse = dtoauthor.AuthorResponse
type AuthorDetailResponse = dtoauthor.AuthorDetailResponse
type PreviewMergeAuthorResponse = dtoauthor.PreviewMergeAuthorResponse
type AffectedNovel = dtoauthor.AffectedNovel
