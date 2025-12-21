-- Update: Cập nhật thông tin artist
-- Flow: Service.UpdateArtist() -> Repository.Update()
-- Params:
--   $1  = id (UUID) - WHERE condition
--   $2  = user_id (UUID, nullable)
--   $3  = name
--   $4  = slug
--   $5  = biography (JSONB)
--   $6  = avatar_url (nullable)
--   $7  = social_links (JSONB)
--   $8  = specialization (nullable)
--   $9  = is_verified (boolean)
--   $10 = updated_by (UUID)
UPDATE catalog.artists
SET user_id = $2,
    name = $3,
    slug = $4,
    biography = $5,
    avatar_url = $6,
    social_links = $7,
    specialization = $8,
    is_verified = $9,
    updated_by = $10
WHERE id = $1 AND deleted_at IS NULL
