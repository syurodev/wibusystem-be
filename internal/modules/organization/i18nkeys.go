package organization

import "system/pkg/util/i18nutil"

// Re-export common keys from i18nutil
const (
	I18nValidationFailed  = i18nutil.ValidationFailed
	I18nAuthUnauthorized  = i18nutil.AuthUnauthorized
	I18nAuthInvalidUserID = i18nutil.AuthInvalidUserID
)

// i18n message keys for organization module
const (
	// Success messages
	I18nCreatedSuccess           = "organization.created_success"
	I18nUpdatedSuccess           = "organization.updated_success"
	I18nDeletedSuccess           = "organization.deleted_success"
	I18nGetSuccess               = "organization.get_success"
	I18nListSuccess              = "organization.list_success"
	I18nMemberInvitedSuccess     = "organization.member_invited_success"
	I18nMemberRemovedSuccess     = "organization.member_removed_success"
	I18nInviteApprovedSuccess    = "organization.invite_approved_success"
	I18nInviteRejectedSuccess    = "organization.invite_rejected_success"
	I18nReportCreatedSuccess     = "organization.report_created_success"
	I18nReportRespondedSuccess   = "organization.report_responded_success"
	I18nSettingsUpdatedSuccess   = "organization.settings_updated_success"
	I18nMemberRoleUpdatedSuccess = "organization.member_role_updated_success"

	// Error messages
	I18nCreateFailed  = "organization.create_failed"
	I18nUpdateFailed  = "organization.update_failed"
	I18nDeleteFailed  = "organization.delete_failed"
	I18nGetFailed     = "organization.get_failed"
	I18nListFailed    = "organization.list_failed"
	I18nInviteFailed  = "organization.invite_failed"
	I18nReportFailed  = "organization.report_failed"

	// Validation and business errors
	I18nNotFound               = "organization.not_found"
	I18nInvalidID              = "organization.invalid_id"
	I18nInvalidInput           = "organization.invalid_input"
	I18nSlugAlreadyExists      = "organization.slug_already_exists"
	I18nAlreadyOwner           = "organization.already_owner"
	I18nMaxMemberships         = "organization.max_memberships_reached"
	I18nCannotRemoveOwner      = "organization.cannot_remove_owner"
	I18nNotMember              = "organization.not_member"
	I18nInsufficientPermission = "organization.insufficient_permission"
	I18nUserAlreadyMember      = "organization.user_already_member"
	I18nUserAlreadyInvited     = "organization.user_already_invited"
	I18nInviteNotFound         = "organization.invite_not_found"
	I18nInviteExpired          = "organization.invite_expired"
	I18nCannotReportOwnOrg     = "organization.cannot_report_own"
	I18nAlreadyReported        = "organization.already_reported"
	I18nReportNotFound         = "organization.report_not_found"
	I18nReportAlreadyResponded = "organization.report_already_responded"
	I18nUserNotFound           = "organization.user_not_found"
	I18nCannotKickOwner        = "organization.cannot_kick_owner"
	I18nOwnerCannotLeave       = "organization.owner_cannot_leave"
)
