-- get_recent.sql
-- ============================================================================
-- Get recent media progress for "Continue Reading" section
--
-- FLOW:
-- 1. Homepage hoặc Continue Section gọi GET /api/v1/history/recent
-- 2. Handler → Service.GetRecent() → Repository.GetRecentMediaProgress()
-- 3. Query JOIN với novels và chapters để lấy đầy đủ thông tin hiển thị
--
-- RETURNS: List of MediaProgress với Media và CurrentUnit đã được populate
--
-- PARAMETERS:
-- $1 = user_id
-- $2 = limit (số lượng items)
-- ============================================================================
SELECT 
    -- Media Progress fields
    mp.id,
    mp.user_id,
    mp.media_type,
    mp.media_id,
    mp.current_unit_id,
    mp.position,
    mp.total_units,
    mp.completed_units,
    mp.progress_percentage,
    mp.last_accessed_at,
    mp.created_at,
    mp.updated_at,
    
    -- Media (Novel) fields
    n.id AS novel_id,
    n.title AS novel_title,
    n.slug AS novel_slug,
    n.cover_image_url AS novel_cover_url,
    n.status AS novel_status,
    n.owner_id,
    u.full_name AS owner_display_name,
    u.username AS owner_username,
    u.avatar_url AS owner_avatar_url,
    
    -- Current Unit (Chapter) fields
    c.id AS chapter_id,
    c.chapter_number,
    c.title AS chapter_title,
    c.slug AS chapter_slug
    
FROM catalog.media_progress mp

-- JOIN với novels table (chỉ cho media_type = 'novel')
LEFT JOIN catalog.novels n 
    ON mp.media_type = 'novel' AND mp.media_id = n.id AND n.deleted_at IS NULL

-- JOIN với users để lấy owner info
LEFT JOIN identify.users u 
    ON n.owner_type = 'user' AND n.owner_id = u.id

-- JOIN với novel_chapters để lấy current unit info
LEFT JOIN catalog.novel_chapters c 
    ON mp.current_unit_id = c.id AND c.deleted_at IS NULL

WHERE mp.user_id = $1

ORDER BY mp.last_accessed_at DESC

LIMIT $2;

