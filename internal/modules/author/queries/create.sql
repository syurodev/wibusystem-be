-- Create: Tạo author mới
-- Flow: Service.CreateAuthor() -> Repository.Create()
-- Params:
--   $1 = id (UUID)
--   $2 = user_id (UUID, nullable)
--   $3 = name
--   $4 = slug
--   $5 = biography (JSONB)
--   $6 = avatar_url (nullable)
--   $7 = social_links (JSONB)
--   $8 = is_verified (boolean)
--   $9 = created_by (UUID)
INSERT INTO catalog.authors (
    id, user_id, name, slug, biography, avatar_url, social_links, is_verified, created_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
