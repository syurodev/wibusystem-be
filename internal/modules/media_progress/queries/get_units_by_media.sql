-- get_units_by_media.sql
-- ============================================================================
-- Get all unit progress for a specific media (for showing read status on chapter list)
--
-- FLOW:
-- 1. User vào trang novel detail, xem chapter list
-- 2. Frontend gọi GET /api/v1/progress/novel/{id}/units
-- 3. Handler → Service → Repository.GetUnitProgressByMedia()
-- 4. Return list các chapter đã đọc với status
-- 5. Frontend merge với chapter list để hiển thị icon "đã đọc"
--
-- PARAMETERS:
-- $1 = user_id
-- $2 = media_type
-- $3 = media_id
-- ============================================================================
SELECT 
    up.id,
    up.user_id,
    up.media_type,
    up.media_id,
    up.unit_id,
    up.status,
    up.position,
    up.started_at,
    up.completed_at,
    up.last_accessed_at
FROM catalog.unit_progress up
WHERE up.user_id = $1 
    AND up.media_type = $2 
    AND up.media_id = $3
ORDER BY up.started_at DESC;
