package i18nkeys

// Author message keys
const (
	// Success messages
	AuthorCreatedSuccess = "author.created_success"
	AuthorUpdatedSuccess = "author.updated_success"
	AuthorDeletedSuccess = "author.deleted_success"
	AuthorGetSuccess     = "author.get_success"
	AuthorListSuccess    = "author.list_success"

	// Error messages
	AuthorCreateFailed = "author.create_failed"
	AuthorUpdateFailed = "author.update_failed"
	AuthorDeleteFailed = "author.delete_failed"
	AuthorGetFailed    = "author.get_failed"
	AuthorListFailed   = "author.list_failed"

	// Validation and business errors
	AuthorNotFound          = "author.not_found"
	AuthorInvalidID         = "author.invalid_id"
	AuthorInvalidInput      = "author.invalid_input"
	AuthorSlugAlreadyExists = "author.slug_already_exists"
	AuthorInUse             = "author.in_use"
)
