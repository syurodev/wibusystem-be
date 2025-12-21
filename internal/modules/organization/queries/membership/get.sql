SELECT user_id, organization_id, status, role, is_active,
       contribution_count, quality_score, metadata,
       invited_by, invited_at, joined_at, left_at,
       created_by, updated_by, deleted_by, version,
       created_at, updated_at, deleted_at
FROM identify.user_organization_memberships
WHERE user_id = $1 AND organization_id = $2 AND deleted_at IS NULL
