package i18nkeys

// Artist message keys
const (
	// Success messages
	ArtistCreatedSuccess = "artist.created_success"
	ArtistUpdatedSuccess = "artist.updated_success"
	ArtistDeletedSuccess = "artist.deleted_success"
	ArtistGetSuccess     = "artist.get_success"
	ArtistListSuccess    = "artist.list_success"

	// Error messages
	ArtistCreateFailed = "artist.create_failed"
	ArtistUpdateFailed = "artist.update_failed"
	ArtistDeleteFailed = "artist.delete_failed"
	ArtistGetFailed    = "artist.get_failed"
	ArtistListFailed   = "artist.list_failed"

	// Validation and business errors
	ArtistNotFound          = "artist.not_found"
	ArtistInvalidID         = "artist.invalid_id"
	ArtistInvalidInput      = "artist.invalid_input"
	ArtistSlugAlreadyExists = "artist.slug_already_exists"
	ArtistInUse             = "artist.in_use"
)
