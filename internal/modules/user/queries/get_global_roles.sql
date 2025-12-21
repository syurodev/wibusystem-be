SELECT r.name
FROM identify.user_global_roles ugr
JOIN identify.roles r ON ugr.role_id = r.id
WHERE ugr.user_id = $1
