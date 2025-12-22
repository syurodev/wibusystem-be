-- mark_unit_complete.sql
-- ============================================================================
-- Mark a unit (chapter/episode) as completed
--
-- FLOW:
-- 1. User scroll tới cuối chapter (hoặc xem xong episode)
-- 2. Frontend gọi POST /api/v1/progress/{media_type}/{media_id}/units/{unit_id}/complete
-- 3. Handler → Service.MarkUnitComplete() → Repository
-- 4. UPDATE unit_progress.status = 'completed'
-- 5. Sau đó Service sẽ gọi UpdateCompletedUnitsCount() để sync
--
-- PARAMETERS:
-- $1 = user_id
-- $2 = unit_id
-- ============================================================================
UPDATE catalog.unit_progress
SET 
    status = 'completed',
    completed_at = NOW(),
    last_accessed_at = NOW()
WHERE user_id = $1 AND unit_id = $2
RETURNING *;
