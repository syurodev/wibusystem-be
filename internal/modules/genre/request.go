package genre

import (
	dtogenre "system/internal/dto/genre"
)

// Re-export request types from dto package
type CreateGenreRequest = dtogenre.CreateGenreRequest
type UpdateGenreRequest = dtogenre.UpdateGenreRequest
type MergeGenreRequest = dtogenre.MergeGenreRequest
type ListGenresRequest = dtogenre.ListGenresRequest
