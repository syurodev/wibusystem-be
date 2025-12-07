package i18nkeys

const (
	// Account Success
	AccountProfileFetchedSuccess = "account.profile_fetched_success"
	AccountProfileUpdatedSuccess = "account.profile_updated_success"
	AccountPasswordChangedSuccess = "account.password_changed_success"
	AccountSessionsFetchedSuccess = "account.sessions_fetched_success"
	AccountSessionDeletedSuccess  = "account.session_deleted_success"

	// Account Errors
	AccountInvalidUserID       = "account.invalid_user_id"
	AccountFetchUserError      = "account.fetch_user_error"
	AccountInvalidRequest      = "account.invalid_request"
	AccountUsernameLengthInvalid = "account.username_length_invalid"
	AccountUsernameTaken       = "account.username_taken"
	AccountUsernameImmutable   = "account.username_immutable"
	AccountUpdateProfileError  = "account.update_profile_error"
	AccountWeakPassword        = "auth.weak_password" // Reusing auth key if appropriate, or create new
	AccountInvalidPassword     = "auth.invalid_password"
	AccountHashError           = "account.hash_error"
	AccountUpdatePasswordError = "account.update_password_error"
	AccountFetchSessionsError  = "account.fetch_sessions_error"
	AccountInvalidSessionID    = "account.invalid_session_id"
	AccountDeleteSessionError  = "account.delete_session_error"
)
