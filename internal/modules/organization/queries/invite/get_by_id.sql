SELECT id, organization_id, user_id, invited_by, status, approved_by, processed_at, expires_at, created_at
FROM identify.organization_pending_invites
WHERE id = $1
