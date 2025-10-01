// Package auth defines global roles and permissions constants
// Based on TypeScript constants from wibutime/lib/auth/constants.ts
package auth

// GlobalRole defines system-wide user roles
type GlobalRole string

const (
	// GlobalRole enum based on backend seed_roles.go
	RoleSuperAdmin GlobalRole = "SUPER_ADMIN"
	RoleAdmin      GlobalRole = "ADMIN"
	RoleModerator  GlobalRole = "MODERATOR"
	RoleUser       GlobalRole = "USER"
	RoleGuest      GlobalRole = "GUEST"
)

// AllGlobalRoles returns all valid global roles
func AllGlobalRoles() []GlobalRole {
	return []GlobalRole{
		RoleSuperAdmin,
		RoleAdmin,
		RoleModerator,
		RoleUser,
		RoleGuest,
	}
}

// IsValidGlobalRole checks if a role is valid
func IsValidGlobalRole(role string) bool {
	for _, validRole := range AllGlobalRoles() {
		if string(validRole) == role {
			return true
		}
	}
	return false
}

// IsAdminRole checks if the role has admin privileges
func IsAdminRole(role GlobalRole) bool {
	return role == RoleSuperAdmin || role == RoleAdmin || role == RoleModerator
}

// GlobalPermission defines system-wide permissions
type GlobalPermission string

const (
	// Auth & User
	PermAuthLogin        GlobalPermission = "auth:login"
	PermAuthLogout       GlobalPermission = "auth:logout"
	PermAuthRefreshToken GlobalPermission = "auth:refresh_token"
	PermUserViewSelf     GlobalPermission = "user:view_self"
	PermUserUpdateSelf   GlobalPermission = "user:update_self"
	PermUserDeleteSelf   GlobalPermission = "user:delete_self"
	PermUserChangePassword GlobalPermission = "user:change_password"
	PermUser2FAManage    GlobalPermission = "user:2fa_manage"

	// Social / Community
	PermCommentCreate     GlobalPermission = "comment:create"
	PermCommentUpdateSelf GlobalPermission = "comment:update_self"
	PermCommentDeleteSelf GlobalPermission = "comment:delete_self"
	PermCommentReport     GlobalPermission = "comment:report"
	PermReactionAdd       GlobalPermission = "reaction:add"
	PermReviewCreate      GlobalPermission = "review:create"
	PermReviewUpdateSelf  GlobalPermission = "review:update_self"
	PermReviewDeleteSelf  GlobalPermission = "review:delete_self"
	PermFollowContent     GlobalPermission = "follow:content"
	PermFollowUser        GlobalPermission = "follow:user"
	PermTranslationSubmit     GlobalPermission = "translation:submit"
	PermTranslationUpdateSelf GlobalPermission = "translation:update_self"
	PermTranslationVote       GlobalPermission = "translation:vote"
	PermSubtitleContribute    GlobalPermission = "subtitle:contribute"
	PermReportContent         GlobalPermission = "report:content"

	// Content viewing
	PermContentViewPublic    GlobalPermission = "content:view_public"
	PermContentViewPurchased GlobalPermission = "content:view_purchased"
	PermContentStreamAnime   GlobalPermission = "content:stream_anime"
	PermContentReadManga     GlobalPermission = "content:read_manga"
	PermContentReadNovel     GlobalPermission = "content:read_novel"

	// Master data: Character
	PermCharacterView                 GlobalPermission = "character:view"
	PermCharacterContribute           GlobalPermission = "character:contribute"
	PermCharacterContributeUpdateSelf GlobalPermission = "character:contribute_update_self"
	PermCharacterCreate               GlobalPermission = "character:create"
	PermCharacterApprove              GlobalPermission = "character:approve"
	PermCharacterReject               GlobalPermission = "character:reject"
	PermCharacterUpdate               GlobalPermission = "character:update"
	PermCharacterDelete               GlobalPermission = "character:delete"

	// Master data: Creator
	PermCreatorView   GlobalPermission = "creator:view"
	PermCreatorCreate GlobalPermission = "creator:create"
	PermCreatorUpdate GlobalPermission = "creator:update"
	PermCreatorDelete GlobalPermission = "creator:delete"

	// Master data: Genre
	PermGenreView   GlobalPermission = "genre:view"
	PermGenreCreate GlobalPermission = "genre:create"
	PermGenreUpdate GlobalPermission = "genre:update"
	PermGenreDelete GlobalPermission = "genre:delete"

	// Master data: Relations
	PermRelationView   GlobalPermission = "relation:view"
	PermRelationCreate GlobalPermission = "relation:create"
	PermRelationUpdate GlobalPermission = "relation:update"
	PermRelationDelete GlobalPermission = "relation:delete"

	// Moderation & System
	PermModerationContentReview GlobalPermission = "moderation:content_review"
	PermModerationUserSuspend   GlobalPermission = "moderation:user_suspend"
	PermModerationBan           GlobalPermission = "moderation:ban"
	PermSystemConfigManage      GlobalPermission = "system:config_manage"
	PermSystemMetricsView       GlobalPermission = "system:metrics_view"
	PermSystemAuditView         GlobalPermission = "system:audit_view"
	PermSupportTicketManage     GlobalPermission = "support:ticket_manage"
)

// TenantPermission defines tenant-scoped permissions
type TenantPermission string

const (
	// Tenant Management
	PermTenantManageMember      TenantPermission = "tenant:manage_member"
	PermTenantAssignPermission  TenantPermission = "tenant:assign_permission"
	PermTenantUpdateInfo        TenantPermission = "tenant:update_info"
	PermTenantViewStats         TenantPermission = "tenant:view_stats"
	PermTenantBillingManage     TenantPermission = "tenant:billing_manage"

	// Content Management
	PermContentCreateAnime TenantPermission = "content:create_anime"
	PermContentUpdateAnime TenantPermission = "content:update_anime"
	PermContentDeleteAnime TenantPermission = "content:delete_anime"
	PermContentCreateManga TenantPermission = "content:create_manga"
	PermContentUpdateManga TenantPermission = "content:update_manga"
	PermContentDeleteManga TenantPermission = "content:delete_manga"
	PermContentCreateNovel TenantPermission = "content:create_novel"
	PermContentUpdateNovel TenantPermission = "content:update_novel"
	PermContentDeleteNovel TenantPermission = "content:delete_novel"

	// Anime Management
	PermAnimeEpisodeCreate TenantPermission = "anime:episode_create"
	PermAnimeEpisodeUpdate TenantPermission = "anime:episode_update"
	PermAnimeEpisodeDelete TenantPermission = "anime:episode_delete"
	PermAnimeSeasonCreate  TenantPermission = "anime:season_create"
	PermAnimeSeasonUpdate  TenantPermission = "anime:season_update"
	PermAnimeSeasonDelete  TenantPermission = "anime:season_delete"

	// Manga Management
	PermMangaChapterCreate TenantPermission = "manga:chapter_create"
	PermMangaChapterUpdate TenantPermission = "manga:chapter_update"
	PermMangaChapterDelete TenantPermission = "manga:chapter_delete"
	PermMangaVolumeCreate  TenantPermission = "manga:volume_create"
	PermMangaVolumeUpdate  TenantPermission = "manga:volume_update"
	PermMangaVolumeDelete  TenantPermission = "manga:volume_delete"

	// Novel Management
	PermNovelChapterCreate TenantPermission = "novel:chapter_create"
	PermNovelChapterUpdate TenantPermission = "novel:chapter_update"
	PermNovelChapterDelete TenantPermission = "novel:chapter_delete"
	PermNovelVolumeCreate  TenantPermission = "novel:volume_create"
	PermNovelVolumeUpdate  TenantPermission = "novel:volume_update"
	PermNovelVolumeDelete  TenantPermission = "novel:volume_delete"

	// Master Data in Tenant Scope
	PermCharacterManage TenantPermission = "character:manage"
	PermCreatorManage   TenantPermission = "creator:manage"
	PermGenreManage     TenantPermission = "genre:manage"
	PermRelationManage  TenantPermission = "relation:manage"

	// Content Publishing
	PermContentPublish   TenantPermission = "content:publish"
	PermContentUnpublish TenantPermission = "content:unpublish"
	PermAnalyticsView    TenantPermission = "analytics:view"
)

// Permission groups for easier management
var (
	// AdminRoles contains roles that have system management access
	AdminRoles = []GlobalRole{
		RoleSuperAdmin,
		RoleAdmin,
		RoleModerator,
	}

	// MasterDataPermissions groups permissions by entity
	MasterDataPermissions = map[string][]GlobalPermission{
		"GENRE": {
			PermGenreView,
			PermGenreCreate,
			PermGenreUpdate,
			PermGenreDelete,
		},
		"CHARACTER": {
			PermCharacterView,
			PermCharacterCreate,
			PermCharacterUpdate,
			PermCharacterDelete,
			PermCharacterApprove,
			PermCharacterReject,
		},
		"CREATOR": {
			PermCreatorView,
			PermCreatorCreate,
			PermCreatorUpdate,
			PermCreatorDelete,
		},
	}

	// TranslationPermissions groups translation-related permissions
	TranslationPermissions = []GlobalPermission{
		PermTranslationSubmit,
		PermTranslationUpdateSelf,
		PermTranslationVote,
	}
)

// Helper functions

// HasPermission checks if a permission is in a slice of permissions
func HasPermission(permissions []GlobalPermission, permission GlobalPermission) bool {
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasTenantPermission checks if a tenant permission is in a slice of tenant permissions
func HasTenantPermission(permissions []TenantPermission, permission TenantPermission) bool {
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// GetPermissionsForEntity returns permissions for a specific master data entity
func GetPermissionsForEntity(entity string) ([]GlobalPermission, bool) {
	perms, exists := MasterDataPermissions[entity]
	return perms, exists
}