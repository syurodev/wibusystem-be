SELECT r.name
FROM identify.user_organization_roles uor
JOIN identify.roles r ON uor.role_id = r.id
WHERE uor.user_id = $1 AND uor.organization_id = $2
