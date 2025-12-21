INSERT INTO identify.organization_reports (
    id, organization_id, reporter_id, reason, description, status
) VALUES ($1, $2, $3, $4, $5, $6)
