-- Merge step 5: Soft delete source genres
UPDATE catalog.genres
SET deleted_at = NOW(),
    deleted_by = $2,
    is_active = false
WHERE id = ANY($1::uuid[])
