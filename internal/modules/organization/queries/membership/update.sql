UPDATE identify.user_organization_memberships
SET status = $3, role = $4, is_active = $5, updated_by = $6, version = version + 1
WHERE user_id = $1 AND organization_id = $2 AND deleted_at IS NULL
