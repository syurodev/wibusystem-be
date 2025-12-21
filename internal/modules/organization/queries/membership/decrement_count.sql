UPDATE identify.organizations SET member_count = member_count - 1 WHERE id = $1 AND member_count > 0
