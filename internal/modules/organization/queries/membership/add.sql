INSERT INTO identify.user_organization_memberships (
    user_id, organization_id, status, role, is_active, 
    invited_by, invited_at, joined_at, created_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
