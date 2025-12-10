package artist

import (
	dtoartist "system/internal/dto/artist"
)

// Re-export request types from dto package
type CreateArtistRequest = dtoartist.CreateArtistRequest
type UpdateArtistRequest = dtoartist.UpdateArtistRequest
type MergeArtistRequest = dtoartist.MergeArtistRequest
type ListArtistsRequest = dtoartist.ListArtistsRequest
