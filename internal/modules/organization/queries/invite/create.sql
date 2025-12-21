INSERT INTO identify.organization_pending_invites (
    id, organization_id, user_id, invited_by, status, expires_at
) VALUES ($1, $2, $3, $4, $5, $6)
