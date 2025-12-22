-- upsert_media_progress.sql
-- ============================================================================
-- UPSERT (INSERT or UPDATE) user's media progress
--
-- FLOW:
-- 1. User mở chapter để đọc
-- 2. Frontend gọi POST /api/v1/history
-- 3. Handler parse request → Service.UpdateProgress()
-- 4. Service gọi Repository.UpsertMediaProgress() với query này
-- 5. INSERT nếu chưa có, UPDATE nếu đã có
--
-- PARAMETERS:
-- $1 = user_id
-- $2 = media_type ('novel', 'manga', 'anime')
-- $3 = media_id
-- $4 = current_unit_id (chapter/episode đang đọc)
-- $5 = position (JSONB - vị trí trong chapter)
-- $6 = total_units (tổng số chapters)
-- ============================================================================
INSERT INTO catalog.media_progress (
    user_id,
    media_type,
    media_id,
    current_unit_id,
    position,
    total_units,
    last_accessed_at
)
VALUES ($1, $2, $3, $4, $5, $6, NOW())
ON CONFLICT (user_id, media_type, media_id)
DO UPDATE SET
    current_unit_id = EXCLUDED.current_unit_id,
    position = EXCLUDED.position,
    total_units = EXCLUDED.total_units,
    last_accessed_at = NOW(),
    updated_at = NOW()
RETURNING *;
