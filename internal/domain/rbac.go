package domain

import (
	"context"

	"github.com/gofrs/uuid/v5"
)

// =====================================================
// Role Constants
// =====================================================

// Role represents a role name in the system
type Role string

const (
	// Global Roles
	RoleSuperAdmin Role = "SUPER_ADMIN"
	RoleAdmin      Role = "ADMIN"
	RoleModerator  Role = "MODERATOR"
	RoleCreator    Role = "CREATOR"
	RoleUser       Role = "USER"
	RoleGuest      Role = "GUEST"
)

// String returns the string representation of the role
func (r Role) String() string {
	return string(r)
}

// IsValid checks if the role is a valid global role
func (r Role) IsValid() bool {
	switch r {
	case RoleSuperAdmin, RoleAdmin, RoleModerator, RoleCreator, RoleUser, RoleGuest:
		return true
	default:
		return false
	}
}

// GetDefaultRole returns the default role for new users
func GetDefaultRole() Role {
	return RoleUser
}

// =====================================================
// Permission Scope
// =====================================================

// PermissionScope represents the scope of a permission
type PermissionScope string

const (
	PermissionScopeGlobal       PermissionScope = "global"
	PermissionScopeOrganization PermissionScope = "organization"
)

// =====================================================
// Global Permissions
// =====================================================

// Permission represents a permission name in the system
type Permission string

const (
	// Auth
	PermAuthLogin        Permission = "auth:login"
	PermAuthLogout       Permission = "auth:logout"
	PermAuthRefreshToken Permission = "auth:refresh_token"

	// User Self-Management
	PermUserViewSelf       Permission = "user:view_self"
	PermUserUpdateSelf     Permission = "user:update_self"
	PermUserDeleteSelf     Permission = "user:delete_self"
	PermUserChangePassword Permission = "user:change_password"
	PermUserTwoFAManage    Permission = "user:two_fa_manage"

	// Social / Community
	PermCommentCreate     Permission = "comment:create"
	PermCommentUpdateSelf Permission = "comment:update_self"
	PermCommentDeleteSelf Permission = "comment:delete_self"
	PermCommentReport     Permission = "comment:report"
	PermReactionAdd       Permission = "reaction:add"
	PermReviewCreate      Permission = "review:create"
	PermReviewUpdateSelf  Permission = "review:update_self"
	PermReviewDeleteSelf  Permission = "review:delete_self"
	PermFollowContent     Permission = "follow:content"
	PermFollowUser        Permission = "follow:user"

	// Translation
	PermTranslationSubmit     Permission = "translation:submit"
	PermTranslationUpdateSelf Permission = "translation:update_self"
	PermTranslationVote       Permission = "translation:vote"
	PermSubtitleContribute    Permission = "subtitle:contribute"

	// Reporting
	PermReportContent Permission = "report:content"

	// Content Viewing
	PermContentViewPublic    Permission = "content:view_public"
	PermContentViewPurchased Permission = "content:view_purchased"
	PermContentStreamAnime   Permission = "content:stream_anime"
	PermContentReadManga     Permission = "content:read_manga"
	PermContentReadNovel     Permission = "content:read_novel"

	// Master Data: Character
	PermCharacterView               Permission = "character:view"
	PermCharacterContribute         Permission = "character:contribute"
	PermCharacterContributeUpdateSelf Permission = "character:contribute_update_self"
	PermCharacterCreate             Permission = "character:create"
	PermCharacterApprove            Permission = "character:approve"
	PermCharacterReject             Permission = "character:reject"
	PermCharacterUpdate             Permission = "character:update"
	PermCharacterDelete             Permission = "character:delete"

	// Master Data: Creator
	PermCreatorView   Permission = "creator:view"
	PermCreatorCreate Permission = "creator:create"
	PermCreatorUpdate Permission = "creator:update"
	PermCreatorDelete Permission = "creator:delete"

	// Master Data: Genre
	PermGenreView   Permission = "genre:view"
	PermGenreCreate Permission = "genre:create"
	PermGenreUpdate Permission = "genre:update"
	PermGenreDelete Permission = "genre:delete"

	// Master Data: Relations
	PermRelationView   Permission = "relation:view"
	PermRelationCreate Permission = "relation:create"
	PermRelationUpdate Permission = "relation:update"
	PermRelationDelete Permission = "relation:delete"

	// Moderation & System
	PermModerationContentReview Permission = "moderation:content_review"
	PermModerationUserSuspend   Permission = "moderation:user_suspend"
	PermModerationBan           Permission = "moderation:ban"
	PermSystemConfigManage      Permission = "system:config_manage"
	PermSystemMetricsView       Permission = "system:metrics_view"
	PermSystemAuditView         Permission = "system:audit_view"
	PermSupportTicketManage     Permission = "support:ticket_manage"
)

// String returns the string representation of the permission
func (p Permission) String() string {
	return string(p)
}

// =====================================================
// Organization Permissions
// =====================================================

const (
	// Tenant Management
	PermTenantManageMember     Permission = "tenant:manage_member"
	PermTenantAssignPermission Permission = "tenant:assign_permission"
	PermTenantUpdateInfo       Permission = "tenant:update_info"
	PermTenantViewStats        Permission = "tenant:view_stats"
	PermTenantBillingManage    Permission = "tenant:billing_manage"

	// Content Management (Anime)
	PermContentCreateAnime Permission = "content:create_anime"
	PermContentUpdateAnime Permission = "content:update_anime"
	PermContentDeleteAnime Permission = "content:delete_anime"

	// Content Management (Manga)
	PermContentCreateManga Permission = "content:create_manga"
	PermContentUpdateManga Permission = "content:update_manga"
	PermContentDeleteManga Permission = "content:delete_manga"

	// Content Management (Novel)
	PermContentCreateNovel Permission = "content:create_novel"
	PermContentUpdateNovel Permission = "content:update_novel"
	PermContentDeleteNovel Permission = "content:delete_novel"

	// Anime Management
	PermAnimeEpisodeCreate Permission = "anime:episode_create"
	PermAnimeEpisodeUpdate Permission = "anime:episode_update"
	PermAnimeEpisodeDelete Permission = "anime:episode_delete"
	PermAnimeSeasonCreate  Permission = "anime:season_create"
	PermAnimeSeasonUpdate  Permission = "anime:season_update"
	PermAnimeSeasonDelete  Permission = "anime:season_delete"

	// Manga Management
	PermMangaChapterCreate Permission = "manga:chapter_create"
	PermMangaChapterUpdate Permission = "manga:chapter_update"
	PermMangaChapterDelete Permission = "manga:chapter_delete"
	PermMangaVolumeCreate  Permission = "manga:volume_create"
	PermMangaVolumeUpdate  Permission = "manga:volume_update"
	PermMangaVolumeDelete  Permission = "manga:volume_delete"

	// Novel Management
	PermNovelChapterCreate Permission = "novel:chapter_create"
	PermNovelChapterUpdate Permission = "novel:chapter_update"
	PermNovelChapterDelete Permission = "novel:chapter_delete"
	PermNovelVolumeCreate  Permission = "novel:volume_create"
	PermNovelVolumeUpdate  Permission = "novel:volume_update"
	PermNovelVolumeDelete  Permission = "novel:volume_delete"

	// Master Data Management in Tenant
	PermCharacterManage Permission = "character:manage"
	PermCreatorManage   Permission = "creator:manage"
	PermGenreManage     Permission = "genre:manage"
	PermRelationManage  Permission = "relation:manage"

	// Content Publishing
	PermContentPublish   Permission = "content:publish"
	PermContentUnpublish Permission = "content:unpublish"
	PermAnalyticsView    Permission = "analytics:view"
)

// =====================================================
// Role Repository Interface
// =====================================================

// RoleRepository defines the interface for role data access
type RoleRepository interface {
	// GetRoleIDByName gets the role ID by name
	GetRoleIDByName(ctx context.Context, name Role) (uuid.UUID, error)

	// AssignGlobalRole assigns a global role to a user
	AssignGlobalRole(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error

	// RemoveGlobalRole removes a global role from a user
	RemoveGlobalRole(ctx context.Context, userID uuid.UUID, roleID uuid.UUID) error

	// AssignOrganizationRole assigns a role to a user within an organization
	AssignOrganizationRole(ctx context.Context, userID, organizationID, roleID uuid.UUID) error

	// RemoveOrganizationRole removes a role from a user within an organization
	RemoveOrganizationRole(ctx context.Context, userID, organizationID, roleID uuid.UUID) error

	// HasGlobalRole checks if a user has a specific global role
	HasGlobalRole(ctx context.Context, userID uuid.UUID, roleName Role) (bool, error)

	// HasOrganizationRole checks if a user has a specific role in an organization
	HasOrganizationRole(ctx context.Context, userID, organizationID uuid.UUID, roleName Role) (bool, error)
}
