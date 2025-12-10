package artist

import (
	dtoartist "system/internal/dto/artist"
)

// Re-export types from dto package for backward compatibility
type ArtistResponse = dtoartist.ArtistResponse
type ArtistDetailResponse = dtoartist.ArtistDetailResponse
type PreviewMergeArtistResponse = dtoartist.PreviewMergeArtistResponse
type AffectedNovel = dtoartist.AffectedNovel
