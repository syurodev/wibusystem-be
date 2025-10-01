package identify

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SeedGlobalRolesAndPermissions seeds global roles and permissions into the database.
// Bao gồm các nhóm quyền: Auth, User, Social, Content Viewing, Master Data, Moderation, System.
func SeedGlobalRolesAndPermissions(ctx context.Context, pool *pgxpool.Pool) error {
	type perm struct {
		key         string
		description string
	}

	permissions := []perm{
		// --- Auth & User ---
		{"auth:login", "Allow user to log into the platform"},
		{"auth:logout", "Allow user to logout"},
		{"auth:refresh_token", "Issue refresh tokens"},
		{"user:view_self", "View own profile"},
		{"user:update_self", "Update own profile"},
		{"user:delete_self", "Delete own account"},
		{"user:change_password", "Change account password"},
		{"user:2fa_manage", "Manage two-factor authentication"},

		// --- Social / Community ---
		{"comment:create", "Create comments"},
		{"comment:update_self", "Edit own comments"},
		{"comment:delete_self", "Delete own comments"},
		{"comment:report", "Report abusive comments"},
		{"reaction:add", "React to content"},
		{"review:create", "Create reviews"},
		{"review:update_self", "Edit own reviews"},
		{"review:delete_self", "Delete own reviews"},
		{"follow:content", "Follow content updates"},
		{"follow:user", "Follow other users"},
		{"translation:submit", "Submit community translations"},
		{"translation:update_self", "Edit own translations"},
		{"translation:vote", "Vote translations"},
		{"subtitle:contribute", "Contribute subtitles"},
		{"report:content", "Report inappropriate content"},

		// --- Content viewing ---
		{"content:view_public", "View public content"},
		{"content:view_purchased", "View purchased/authorized content"},
		{"content:stream_anime", "Stream anime"},
		{"content:read_manga", "Read manga"},
		{"content:read_novel", "Read novel"},

		// --- Master data: Character ---
		{"character:view", "View characters"},
		{"character:contribute", "Contribute edits/suggestions to any character"},
		{"character:contribute_update_self", "Update or delete own character contributions"},
		{"character:create", "Create new characters (requires approval if not staff)"},
		{"character:approve", "Approve new characters or contributions"},
		{"character:reject", "Reject character contributions"},
		{"character:update", "Update characters directly (staff only)"},
		{"character:delete", "Delete characters (staff only)"},

		// --- Master data: Creator ---
		{"creator:view", "View creators"},
		{"creator:create", "Create new creators"},
		{"creator:update", "Update creators"},
		{"creator:delete", "Delete creators"},

		// --- Master data: Genre ---
		{"genre:view", "View genres"},
		{"genre:create", "Create genres"},
		{"genre:update", "Update genres"},
		{"genre:delete", "Delete genres"},

		// --- Master data: Relations ---
		{"relation:view", "View content relations"},
		{"relation:create", "Create content relations"},
		{"relation:update", "Update content relations"},
		{"relation:delete", "Delete content relations"},

		// --- Moderation & System ---
		{"moderation:content_review", "Review reported content"},
		{"moderation:user_suspend", "Suspend users"},
		{"moderation:ban", "Ban users"},
		{"system:config_manage", "Manage system configuration"},
		{"system:metrics_view", "View platform metrics"},
		{"system:audit_view", "View audit logs"},
		{"support:ticket_manage", "Manage support tickets"},
	}

	// insert permissions
	permIDs := make(map[string]uuid.UUID, len(permissions))
	for _, p := range permissions {
		var id uuid.UUID
		err := pool.QueryRow(ctx, `
            INSERT INTO global_permissions (key, description)
            VALUES ($1, $2)
            ON CONFLICT (key) DO UPDATE SET description = EXCLUDED.description
            RETURNING id
        `, p.key, p.description).Scan(&id)
		if err != nil {
			return fmt.Errorf("seed global permission %s: %w", p.key, err)
		}
		permIDs[p.key] = id
	}

	// define roles
	type role struct {
		name        string
		description string
		perms       []string
	}

	// SUPER_ADMIN có tất cả permissions
	allKeys := make([]string, len(permissions))
	for i, p := range permissions {
		allKeys[i] = p.key
	}

	roles := []role{
		{
			name:        "SUPER_ADMIN",
			description: "Full access to all system capabilities",
			perms:       allKeys,
		},
		{
			name:        "ADMIN",
			description: "Manage users, tenants, and master data",
			perms: []string{
				"auth:login", "auth:logout", "auth:refresh_token",
				"user:view_self", "user:update_self", "user:delete_self", "user:change_password", "user:2fa_manage",
				"comment:report", "report:content",
				"content:view_public", "content:view_purchased", "content:stream_anime", "content:read_manga", "content:read_novel",
				// Master data management
				"character:view", "character:create", "character:approve", "character:reject", "character:update", "character:delete",
				"creator:view", "creator:create", "creator:update", "creator:delete",
				"genre:view", "genre:create", "genre:update", "genre:delete",
				"relation:view", "relation:create", "relation:update", "relation:delete",
				// System
				"system:metrics_view", "system:audit_view", "support:ticket_manage",
			},
		},
		{
			name:        "MODERATOR",
			description: "Moderate community content and manage master data",
			perms: []string{
				"auth:login", "auth:logout", "auth:refresh_token",
				"user:view_self", "user:update_self", "user:change_password", "user:2fa_manage",
				"comment:report", "translation:vote", "report:content",
				"content:view_public", "content:view_purchased", "content:stream_anime", "content:read_manga", "content:read_novel",
				// Moderation
				"moderation:content_review", "moderation:user_suspend", "moderation:ban",
				// Master data (approve/reject only)
				"character:view", "character:approve", "character:reject",
				"creator:view", "genre:view", "relation:view",
				// System
				"system:metrics_view", "support:ticket_manage",
			},
		},
		{
			name:        "USER",
			description: "Default user role with community & content access",
			perms: []string{
				"auth:login", "auth:logout", "auth:refresh_token",
				"user:view_self", "user:update_self", "user:change_password", "user:2fa_manage",
				"comment:create", "comment:update_self", "comment:delete_self", "comment:report",
				"reaction:add", "review:create", "review:update_self", "review:delete_self",
				"follow:content", "follow:user",
				"translation:submit", "translation:update_self", "translation:vote", "subtitle:contribute", "report:content",
				"content:view_public", "content:view_purchased", "content:stream_anime", "content:read_manga", "content:read_novel",
				// Master data (view + contribute only)
				"character:view", "character:contribute", "character:contribute_update_self",
				"creator:view", "genre:view", "relation:view",
			},
		},
		{
			name:        "GUEST",
			description: "Unauthenticated visitor with read-only access",
			perms: []string{
				"content:view_public",
			},
		},
	}

	// insert roles + mapping
	for _, r := range roles {
		var roleID uuid.UUID
		err := pool.QueryRow(ctx, `
            INSERT INTO global_roles (name, description)
            VALUES ($1, $2)
            ON CONFLICT (name) DO UPDATE SET description = EXCLUDED.description
            RETURNING id
        `, r.name, r.description).Scan(&roleID)
		if err != nil {
			return fmt.Errorf("seed global role %s: %w", r.name, err)
		}

		if _, err := pool.Exec(ctx, `DELETE FROM global_role_permissions WHERE role_id = $1`, roleID); err != nil {
			return fmt.Errorf("cleanup global role permissions for %s: %w", r.name, err)
		}

		for _, key := range r.perms {
			permID, ok := permIDs[key]
			if !ok {
				return fmt.Errorf("permission %s not found when assigning to role %s", key, r.name)
			}
			if _, err := pool.Exec(ctx, `
                INSERT INTO global_role_permissions (role_id, permission_id)
                VALUES ($1, $2)
                ON CONFLICT DO NOTHING
            `, roleID, permID); err != nil {
				return fmt.Errorf("assign permission %s to role %s: %w", key, r.name, err)
			}
		}
	}

	return nil
}

// SeedTenantPermissions inserts tenant-level permissions (không có role mặc định).
// Đây là các quyền để quản lý tenant, nội dung trong phạm vi tenant.
func SeedTenantPermissions(ctx context.Context, pool *pgxpool.Pool) error {
	type perm struct {
		key         string
		description string
	}

	permissions := []perm{
		{"tenant:manage_member", "Manage tenant members"},
		{"tenant:assign_permission", "Assign permissions within tenant"},
		{"tenant:update_info", "Update tenant profile information"},
		{"tenant:view_stats", "View tenant statistics"},
		{"tenant:billing_manage", "Manage tenant billing"},
		{"content:create_anime", "Create anime entries"},
		{"content:update_anime", "Update anime entries"},
		{"content:delete_anime", "Delete anime entries"},
		{"content:create_manga", "Create manga entries"},
		{"content:update_manga", "Update manga entries"},
		{"content:delete_manga", "Delete manga entries"},
		{"content:create_novel", "Create novel entries"},
		{"content:update_novel", "Update novel entries"},
		{"content:delete_novel", "Delete novel entries"},
		{"anime:episode_create", "Create anime episodes"},
		{"anime:episode_update", "Update anime episodes"},
		{"anime:episode_delete", "Delete anime episodes"},
		{"manga:chapter_create", "Create manga chapters"},
		{"manga:chapter_update", "Update manga chapters"},
		{"manga:chapter_delete", "Delete manga chapters"},
		{"novel:chapter_create", "Create novel chapters"},
		{"novel:chapter_update", "Update novel chapters"},
		{"novel:chapter_delete", "Delete novel chapters"},
		{"manga:volume_create", "Create manga volumes"},
		{"manga:volume_update", "Update manga volumes"},
		{"manga:volume_delete", "Delete manga volumes"},
		{"novel:volume_create", "Create novel volumes"},
		{"novel:volume_update", "Update novel volumes"},
		{"novel:volume_delete", "Delete novel volumes"},
		{"anime:season_create", "Create anime seasons"},
		{"anime:season_update", "Update anime seasons"},
		{"anime:season_delete", "Delete anime seasons"},
		{"character:manage", "Manage characters in tenant scope"},
		{"creator:manage", "Manage creators in tenant scope"},
		{"genre:manage", "Manage genres in tenant scope"},
		{"relation:manage", "Manage content relations in tenant scope"},
		{"content:publish", "Publish content"},
		{"content:unpublish", "Unpublish content"},
		{"analytics:view", "View content analytics"},
	}

	for _, p := range permissions {
		if _, err := pool.Exec(ctx, `
            INSERT INTO permissions (key, description)
            VALUES ($1, $2)
            ON CONFLICT (key) DO UPDATE SET description = EXCLUDED.description
        `, p.key, p.description); err != nil {
			return fmt.Errorf("seed tenant permission %s: %w", p.key, err)
		}
	}

	return nil
}

// SeedAllRolesAndPermissions runs both global and tenant permission seeds.
func SeedAllRolesAndPermissions(ctx context.Context, pool *pgxpool.Pool) error {
	if err := SeedGlobalRolesAndPermissions(ctx, pool); err != nil {
		return err
	}
	if err := SeedTenantPermissions(ctx, pool); err != nil {
		return err
	}
	return nil
}
