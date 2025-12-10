package novel

import (
	dtonovel "system/internal/dto/novel"
)

// Re-export request types from dto package
type CreateNovelRequest = dtonovel.CreateNovelRequest
type UpdateNovelRequest = dtonovel.UpdateNovelRequest
type ListNovelsRequest = dtonovel.ListNovelsRequest
