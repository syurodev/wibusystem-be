-- Delete: Soft delete author
-- Flow: Service.DeleteAuthor() -> Repository.Delete()
-- Params: $1 = author UUID
UPDATE catalog.authors
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
