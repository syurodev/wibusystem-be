-- Delete: Soft delete artist
-- Flow: Service.DeleteArtist() -> Repository.Delete()
-- Params: $1 = artist UUID
UPDATE catalog.artists
SET deleted_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
