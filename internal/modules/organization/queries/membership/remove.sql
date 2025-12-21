UPDATE identify.user_organization_memberships
SET deleted_at = NOW(), left_at = NOW()
WHERE user_id = $1 AND organization_id = $2 AND deleted_at IS NULL
