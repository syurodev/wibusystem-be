SELECT EXISTS(SELECT 1 FROM identify.user_organization_memberships 
WHERE user_id = $1 AND role::text = 'owner' AND deleted_at IS NULL)
