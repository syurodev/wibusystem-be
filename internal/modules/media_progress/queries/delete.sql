-- delete.sql
-- ============================================================================
-- Delete media progress and all related unit progress
--
-- FLOW:
-- 1. User chọn xóa lịch sử của 1 media
-- 2. Frontend gọi DELETE /api/v1/history/{id}
-- 3. Handler → Service → Repository.DeleteMediaProgress()
-- 4. Xóa cả media_progress và các unit_progress liên quan
--
-- PART 1: Delete unit progress
-- PARAMETERS: $1 = user_id, $2 = media_type, $3 = media_id
-- ============================================================================
DELETE FROM catalog.unit_progress
WHERE user_id = $1 AND media_type = $2 AND media_id = $3;

-- ============================================================================
-- PART 2: Delete media progress
-- PARAMETERS: $1 = media_progress_id
-- ============================================================================
-- DELETE FROM catalog.media_progress WHERE id = $1;
