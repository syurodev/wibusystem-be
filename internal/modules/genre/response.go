package genre

import (
	dtogenre "system/internal/dto/genre"
)

// Re-export types from dto package for backward compatibility
type GenreResponse = dtogenre.GenreResponse
type GenreDetailResponse = dtogenre.GenreDetailResponse
type PreviewMergeGenreResponse = dtogenre.PreviewMergeGenreResponse
type AffectedNovel = dtogenre.AffectedNovel
