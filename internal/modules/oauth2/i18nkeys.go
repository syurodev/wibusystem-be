package oauth2

import "system/pkg/util/i18nutil"

// Re-export common keys from i18nutil
const (
	I18nValidationFailed  = i18nutil.ValidationFailed
	I18nAuthUnauthorized  = i18nutil.AuthUnauthorized
	I18nAuthInvalidUserID = i18nutil.AuthInvalidUserID
)

// i18n message keys for OAuth2 module
const (
	// Client CRUD success messages
	I18nClientCreated           = "client.created"
	I18nClientUpdated           = "client.updated"
	I18nClientDeleted           = "client.deleted"
	I18nClientRetrieved         = "client.retrieved"
	I18nClientListed            = "client.listed"
	I18nClientSecretRegenerated = "client.secret_regenerated"

	// Client error messages
	I18nClientNotFound         = "client.not_found"
	I18nClientCreateFailed     = "client.create_failed"
	I18nClientUpdateFailed     = "client.update_failed"
	I18nClientDeleteFailed     = "client.delete_failed"
	I18nClientListFailed       = "client.list_failed"
	I18nClientSecretGenFailed  = "client.secret_generation_failed"
	I18nClientSecretHashFailed = "client.secret_hash_failed"
	I18nClientPublicNoSecret   = "client.public_no_secret"

	// Validation errors
	I18nInvalidGrantType    = "validation.invalid_grant_type"
	I18nInvalidResponseType = "validation.invalid_response_type"
	I18nInvalidClientID     = "validation.invalid_client_id"
	I18nInvalidTenantID     = "validation.invalid_tenant_id"
	I18nInvalidUserID       = "validation.invalid_user_id"

	// Auth errors
	I18nInvalidToken       = "auth.invalid_token"
	I18nInvalidCredentials = "auth.invalid_credentials"

	// Resource errors
	I18nUserNotFound = "resource.not_found"
)
