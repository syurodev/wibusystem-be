-- update_completed_count.sql
-- ============================================================================
-- Update completed_units count in media_progress
--
-- FLOW:
-- 1. Sau khi mark unit complete
-- 2. Service gọi Repository.UpdateCompletedUnitsCount()
-- 3. Query này COUNT từ unit_progress và UPDATE vào media_progress
--
-- PARAMETERS:
-- $1 = user_id
-- $2 = media_type
-- $3 = media_id
-- ============================================================================
UPDATE catalog.media_progress mp
SET 
    completed_units = (
        SELECT COUNT(*)
        FROM catalog.unit_progress up
        WHERE up.user_id = mp.user_id 
            AND up.media_type = mp.media_type 
            AND up.media_id = mp.media_id
            AND up.status = 'completed'
    ),
    progress_percentage = CASE 
        WHEN mp.total_units > 0 
        THEN ROUND(
            (SELECT COUNT(*)::numeric FROM catalog.unit_progress up
             WHERE up.user_id = mp.user_id 
                 AND up.media_type = mp.media_type 
                 AND up.media_id = mp.media_id
                 AND up.status = 'completed') 
            / mp.total_units * 100, 2
        )
        ELSE 0
    END,
    updated_at = NOW()
WHERE mp.user_id = $1 
    AND mp.media_type = $2 
    AND mp.media_id = $3
RETURNING *;
