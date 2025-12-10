package author

import (
	dtoauthor "system/internal/dto/author"
)

// Re-export request types from dto package
type CreateAuthorRequest = dtoauthor.CreateAuthorRequest
type UpdateAuthorRequest = dtoauthor.UpdateAuthorRequest
type MergeAuthorRequest = dtoauthor.MergeAuthorRequest
type ListAuthorsRequest = dtoauthor.ListAuthorsRequest
