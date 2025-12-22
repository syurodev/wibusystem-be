-- upsert_unit_progress.sql
-- ============================================================================
-- UPSERT (INSERT or UPDATE) user's unit (chapter/episode) progress
--
-- FLOW:
-- 1. User mở chapter để đọc
-- 2. Service gọi Repository.UpsertUnitProgress()
-- 3. INSERT với status='in_progress' nếu chưa có
-- 4. UPDATE position và last_accessed_at nếu đã có
--
-- PARAMETERS:
-- $1 = user_id
-- $2 = media_type
-- $3 = media_id
-- $4 = unit_id (chapter_id hoặc episode_id)
-- $5 = status ('in_progress' hoặc 'completed')
-- $6 = position (JSONB)
-- ============================================================================
INSERT INTO catalog.unit_progress (
    user_id,
    media_type,
    media_id,
    unit_id,
    status,
    position,
    started_at,
    last_accessed_at
)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
ON CONFLICT (user_id, unit_id)
DO UPDATE SET
    status = EXCLUDED.status,
    position = EXCLUDED.position,
    last_accessed_at = NOW(),
    -- Chỉ set completed_at khi status chuyển sang completed
    completed_at = CASE 
        WHEN EXCLUDED.status = 'completed' AND catalog.unit_progress.status != 'completed' 
        THEN NOW() 
        ELSE catalog.unit_progress.completed_at 
    END
RETURNING *;
