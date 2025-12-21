SELECT EXISTS(SELECT 1 FROM identify.organizations WHERE slug = $1 AND deleted_at IS NULL)
