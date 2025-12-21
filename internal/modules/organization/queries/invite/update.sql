UPDATE identify.organization_pending_invites
SET status = $2, approved_by = $3, processed_at = $4
WHERE id = $1
