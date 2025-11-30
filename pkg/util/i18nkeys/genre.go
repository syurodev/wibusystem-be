package i18nkeys

// Genre message keys
const (
	// Success messages
	GenreCreatedSuccess = "genre.created_success"
	GenreUpdatedSuccess = "genre.updated_success"
	GenreDeletedSuccess = "genre.deleted_success"
	GenreGetSuccess     = "genre.get_success"
	GenreListSuccess    = "genre.list_success"

	// Error messages
	GenreCreateFailed = "genre.create_failed"
	GenreUpdateFailed = "genre.update_failed"
	GenreDeleteFailed = "genre.delete_failed"
	GenreGetFailed    = "genre.get_failed"
	GenreListFailed   = "genre.list_failed"

	// Validation and business errors
	GenreNotFound          = "genre.not_found"
	GenreInvalidID         = "genre.invalid_id"
	GenreInvalidInput      = "genre.invalid_input"
	GenreSlugAlreadyExists = "genre.slug_already_exists"
	GenreParentNotFound    = "genre.parent_not_found"
	GenreInvalidParentID   = "genre.invalid_parent_id"
	GenreInUse             = "genre.in_use"
	GenreCircularReference = "genre.circular_reference"
	GenreHasChildren       = "genre.has_children"
)
