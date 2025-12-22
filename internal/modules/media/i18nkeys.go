package media

import "system/pkg/util/i18nutil"

// Re-export common keys from i18nutil
const (
	I18nValidationFailed  = i18nutil.ValidationFailed
	I18nAuthUnauthorized  = i18nutil.AuthUnauthorized
	I18nAuthInvalidUserID = i18nutil.AuthInvalidUserID
)

// I18n key constants cho Media module
const (
	I18nMediaGetHomeSuccess = "media.get_home_success"
	I18nMediaGetHomeFailed  = "media.get_home_failed"
	I18nMediaGetTopSuccess  = "media.get_top_success"

	// Analytics-related keys used in media handler
	I18nAnalyticsGetTrendingSuccess  = "analytics.get_trending_success"
	I18nAnalyticsGetTrendingFailed   = "analytics.get_trending_failed"
	I18nAnalyticsSerializationFailed = "analytics.serialization_failed"
	I18nAnalyticsMappingFailed       = "analytics.mapping_failed"
)
