package payment

import "system/pkg/util/i18nutil"

// Re-export common keys from i18nutil
const (
	I18nValidationFailed  = i18nutil.ValidationFailed
	I18nAuthUnauthorized  = i18nutil.AuthUnauthorized
	I18nAuthInvalidUserID = i18nutil.AuthInvalidUserID
	I18nInternalError     = i18nutil.InternalError
)

// i18n message keys for payment module - Config
const (
	// Config Success
	I18nConfigListSuccess   = "payment.config.list_success"
	I18nConfigGetSuccess    = "payment.config.get_success"
	I18nConfigUpdateSuccess = "payment.config.update_success"
	I18nConfigCreateSuccess = "payment.config.create_success"
	I18nConfigDeleteSuccess = "payment.config.delete_success"

	// Config Errors
	I18nConfigNotFound      = "payment.config.not_found"
	I18nConfigAlreadyExists = "payment.config.already_exists"
	I18nConfigInvalidValue  = "payment.config.invalid_value"
	I18nConfigListFailed    = "payment.config.list_failed"
	I18nConfigGetFailed     = "payment.config.get_failed"
	I18nConfigUpdateFailed  = "payment.config.update_failed"
	I18nConfigCreateFailed  = "payment.config.create_failed"
	I18nConfigDeleteFailed  = "payment.config.delete_failed"
)

// i18n message keys for payment module - Wallet
const (
	// Wallet Success
	I18nWalletGetSuccess     = "payment.wallet.get_success"
	I18nWalletHistorySuccess = "payment.wallet.history_success"

	// Wallet Errors
	I18nWalletNotFound       = "payment.wallet.not_found"
	I18nWalletGetFailed      = "payment.wallet.get_failed"
	I18nInsufficientBalance  = "payment.wallet.insufficient_balance"
)

// i18n message keys for payment module - Package
const (
	// Package Success
	I18nPackageListSuccess = "payment.package.list_success"

	// Package Errors
	I18nPackageNotFound    = "payment.package.not_found"
	I18nPackageListFailed  = "payment.package.list_failed"
	I18nPackageInactive    = "payment.package.inactive"
)

// i18n message keys for payment module - Topup
const (
	// Topup Success
	I18nTopupCreateSuccess  = "payment.topup.create_success"
	I18nTopupGetSuccess     = "payment.topup.get_success"
	I18nTopupListSuccess    = "payment.topup.list_success"
	I18nTopupCancelSuccess  = "payment.topup.cancel_success"
	I18nTopupCompleteSuccess = "payment.topup.complete_success"
	I18nTopupSuccessNotification = "payment.topup.notification.success"

	// Topup Errors
	I18nTopupNotFound       = "payment.topup.not_found"
	I18nTopupCreateFailed   = "payment.topup.create_failed"
	I18nTopupAlreadyPending = "payment.topup.already_pending"
	I18nTopupExpired        = "payment.topup.expired"
	I18nTopupAlreadyCompleted = "payment.topup.already_completed"
	I18nTopupInvalidWebhook = "payment.topup.invalid_webhook"
	I18nTopupAmountMismatch = "payment.topup.amount_mismatch"
)

// i18n message keys for payment module - Transaction
const (
	// Transaction Success
	I18nTransactionListSuccess = "payment.transaction.list_success"

	// Transaction Errors
	I18nTransactionListFailed = "payment.transaction.list_failed"
)
