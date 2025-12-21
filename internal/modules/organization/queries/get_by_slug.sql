SELECT id, name, slug, status, description, avatar_url, settings,
       is_recruiting, can_translate, can_proofread, can_edit,
       member_count, active_projects, completed_translations, report_count,
       metadata, created_by, updated_by, deleted_by, version,
       created_at, updated_at, deleted_at
FROM identify.organizations
WHERE slug = $1 AND deleted_at IS NULL
