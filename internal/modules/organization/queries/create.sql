INSERT INTO identify.organizations (
    id, name, slug, status, description, avatar_url, settings, 
    is_recruiting, created_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
