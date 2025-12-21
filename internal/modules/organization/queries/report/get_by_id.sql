SELECT id, organization_id, reporter_id, reason, description,
       org_response, org_responded_by, org_responded_at,
       status, resolved_by, resolved_at, resolution_note,
       created_at, updated_at
FROM identify.organization_reports
WHERE id = $1
