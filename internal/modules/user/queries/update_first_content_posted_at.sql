UPDATE identify.user_statistics
SET first_content_posted_at = $2
WHERE user_id = $1 AND first_content_posted_at IS NULL
