UPDATE identify.organization_reports
SET org_response = $2, org_responded_by = $3, org_responded_at = $4, 
    status = $5, updated_at = $6
WHERE id = $1
