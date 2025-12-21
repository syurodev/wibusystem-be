UPDATE identify.organizations
SET name = $2, description = $3, avatar_url = $4, settings = $5,
    is_recruiting = $6, updated_by = $7, version = version + 1
WHERE id = $1 AND deleted_at IS NULL
