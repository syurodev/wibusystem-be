-- Enhanced Novel Tables Migration
-- Adds comprehensive fields for user tracking, content management, analytics, and monetization

-- =============================================================================
-- NOVEL TABLE ENHANCEMENTS
-- =============================================================================

-- User and Tenant tracking
ALTER TABLE novel ADD COLUMN created_by_user_id UUID; -- User tạo novel
ALTER TABLE novel ADD COLUMN updated_by_user_id UUID; -- User cập nhật cuối cùng
ALTER TABLE novel ADD COLUMN tenant_id UUID; -- Tenant sở hữu novel

-- Publishing information
ALTER TABLE novel ADD COLUMN published_at TIMESTAMP; -- Ngày xuất bản chính thức
ALTER TABLE novel ADD COLUMN original_language VARCHAR(5) DEFAULT 'en'; -- Ngôn ngữ gốc (ISO 639-1)
ALTER TABLE novel ADD COLUMN source_url TEXT; -- URL nguồn gốc (nếu có)
ALTER TABLE novel ADD COLUMN isbn VARCHAR(17); -- ISBN code cho xuất bản

-- Content Rating và Warnings
ALTER TABLE novel ADD COLUMN age_rating VARCHAR(10); -- G, PG, PG-13, R, NC-17
ALTER TABLE novel ADD COLUMN content_warnings JSONB; -- Cảnh báo nội dung (violence, sexual, etc)
ALTER TABLE novel ADD COLUMN mature_content BOOLEAN DEFAULT FALSE; -- Nội dung người lớn

-- Visibility and Status
ALTER TABLE novel ADD COLUMN is_public BOOLEAN DEFAULT FALSE; -- Công khai hay riêng tư
ALTER TABLE novel ADD COLUMN is_featured BOOLEAN DEFAULT FALSE; -- Được đề xuất trên trang chủ
ALTER TABLE novel ADD COLUMN is_completed BOOLEAN DEFAULT FALSE; -- Đã hoàn thành
ALTER TABLE novel ADD COLUMN is_deleted BOOLEAN DEFAULT FALSE; -- Đã xóa (soft delete)
ALTER TABLE novel ADD COLUMN deleted_at TIMESTAMP; -- Thời gian xóa
ALTER TABLE novel ADD COLUMN deleted_by_user_id UUID; -- User thực hiện xóa

-- SEO and Discovery
ALTER TABLE novel ADD COLUMN slug VARCHAR(255) UNIQUE; -- URL-friendly identifier
ALTER TABLE novel ADD COLUMN tags JSONB; -- Tags cho tìm kiếm và phân loại
ALTER TABLE novel ADD COLUMN keywords TEXT; -- Keywords cho SEO
ALTER TABLE novel ADD COLUMN meta_description TEXT; -- Meta description cho SEO

-- Analytics và Engagement
ALTER TABLE novel ADD COLUMN view_count BIGINT DEFAULT 0; -- Lượt xem novel
ALTER TABLE novel ADD COLUMN like_count BIGINT DEFAULT 0; -- Lượt thích
ALTER TABLE novel ADD COLUMN bookmark_count BIGINT DEFAULT 0; -- Lượt bookmark/favorite
ALTER TABLE novel ADD COLUMN comment_count BIGINT DEFAULT 0; -- Số bình luận
ALTER TABLE novel ADD COLUMN rating_average DECIMAL(3,2); -- Điểm đánh giá trung bình (0.00-5.00)
ALTER TABLE novel ADD COLUMN rating_count INTEGER DEFAULT 0; -- Số lượt đánh giá

-- Pricing và Monetization
ALTER TABLE novel ADD COLUMN price_coins INTEGER; -- Giá mua toàn bộ series (coins)
ALTER TABLE novel ADD COLUMN rental_price_coins INTEGER; -- Giá thuê series (coins)
ALTER TABLE novel ADD COLUMN rental_duration_days INTEGER; -- Thời hạn thuê series (ngày)
ALTER TABLE novel ADD COLUMN is_premium BOOLEAN DEFAULT FALSE; -- Nội dung premium

-- Metadata bổ sung
ALTER TABLE novel ADD COLUMN total_chapters INTEGER DEFAULT 0; -- Tổng số chương
ALTER TABLE novel ADD COLUMN total_volumes INTEGER DEFAULT 0; -- Tổng số tập
ALTER TABLE novel ADD COLUMN estimated_reading_time INTEGER; -- Thời gian đọc ước tính (phút)
ALTER TABLE novel ADD COLUMN word_count INTEGER; -- Tổng số từ trong toàn bộ novel

-- =============================================================================
-- NOVEL_VOLUME TABLE ENHANCEMENTS
-- =============================================================================

-- User tracking
ALTER TABLE novel_volume ADD COLUMN created_by_user_id UUID; -- User tạo volume
ALTER TABLE novel_volume ADD COLUMN updated_by_user_id UUID; -- User cập nhật volume

-- Publishing và Status
ALTER TABLE novel_volume ADD COLUMN published_at TIMESTAMP; -- Ngày xuất bản volume
ALTER TABLE novel_volume ADD COLUMN is_deleted BOOLEAN DEFAULT FALSE; -- Đã xóa (soft delete)
ALTER TABLE novel_volume ADD COLUMN deleted_at TIMESTAMP; -- Thời gian xóa
ALTER TABLE novel_volume ADD COLUMN is_available BOOLEAN DEFAULT TRUE; -- Có sẵn để đọc

-- Content metadata
ALTER TABLE novel_volume ADD COLUMN page_count INTEGER; -- Số trang (ước tính)
ALTER TABLE novel_volume ADD COLUMN word_count INTEGER; -- Số từ trong volume
ALTER TABLE novel_volume ADD COLUMN chapter_count INTEGER DEFAULT 0; -- Số chương trong volume
ALTER TABLE novel_volume ADD COLUMN estimated_reading_time INTEGER; -- Thời gian đọc ước tính (phút)

-- =============================================================================
-- NOVEL_CHAPTER TABLE ENHANCEMENTS
-- =============================================================================

-- User tracking
ALTER TABLE novel_chapter ADD COLUMN created_by_user_id UUID; -- User tạo chapter
ALTER TABLE novel_chapter ADD COLUMN updated_by_user_id UUID; -- User cập nhật chapter

-- Publishing workflow
ALTER TABLE novel_chapter ADD COLUMN scheduled_publish_at TIMESTAMP; -- Lên lịch xuất bản
ALTER TABLE novel_chapter ADD COLUMN is_draft BOOLEAN DEFAULT TRUE; -- Bản nháp
ALTER TABLE novel_chapter ADD COLUMN is_deleted BOOLEAN DEFAULT FALSE; -- Đã xóa (soft delete)
ALTER TABLE novel_chapter ADD COLUMN deleted_at TIMESTAMP; -- Thời gian xóa
ALTER TABLE novel_chapter ADD COLUMN version INTEGER DEFAULT 1; -- Version của chapter

-- Content metadata
ALTER TABLE novel_chapter ADD COLUMN word_count INTEGER; -- Số từ trong chapter
ALTER TABLE novel_chapter ADD COLUMN reading_time_minutes INTEGER; -- Thời gian đọc ước tính (phút)
ALTER TABLE novel_chapter ADD COLUMN character_count INTEGER; -- Số ký tự

-- Analytics
ALTER TABLE novel_chapter ADD COLUMN view_count BIGINT DEFAULT 0; -- Lượt xem chapter
ALTER TABLE novel_chapter ADD COLUMN like_count BIGINT DEFAULT 0; -- Lượt thích chapter
ALTER TABLE novel_chapter ADD COLUMN comment_count BIGINT DEFAULT 0; -- Số bình luận chapter

-- Content warnings cho chapter riêng lẻ
ALTER TABLE novel_chapter ADD COLUMN content_warnings JSONB; -- Cảnh báo nội dung riêng cho chapter
ALTER TABLE novel_chapter ADD COLUMN has_mature_content BOOLEAN DEFAULT FALSE; -- Chapter có nội dung nhạy cảm

-- =============================================================================
-- INDEXES FOR PERFORMANCE
-- =============================================================================

-- Novel indexes
CREATE INDEX idx_novel_tenant_id ON novel(tenant_id);
CREATE INDEX idx_novel_created_by_user_id ON novel(created_by_user_id);
CREATE INDEX idx_novel_published_at ON novel(published_at);
CREATE INDEX idx_novel_is_public ON novel(is_public);
CREATE INDEX idx_novel_is_featured ON novel(is_featured);
CREATE INDEX idx_novel_is_deleted ON novel(is_deleted);
CREATE INDEX idx_novel_age_rating ON novel(age_rating);
CREATE INDEX idx_novel_mature_content ON novel(mature_content);
CREATE INDEX idx_novel_tags ON novel USING GIN(tags);
CREATE INDEX idx_novel_rating_average ON novel(rating_average);

-- Volume indexes
CREATE INDEX idx_novel_volume_created_by_user_id ON novel_volume(created_by_user_id);
CREATE INDEX idx_novel_volume_published_at ON novel_volume(published_at);
CREATE INDEX idx_novel_volume_is_deleted ON novel_volume(is_deleted);

-- Chapter indexes
CREATE INDEX idx_novel_chapter_created_by_user_id ON novel_chapter(created_by_user_id);
CREATE INDEX idx_novel_chapter_scheduled_publish_at ON novel_chapter(scheduled_publish_at);
CREATE INDEX idx_novel_chapter_is_draft ON novel_chapter(is_draft);
CREATE INDEX idx_novel_chapter_is_deleted ON novel_chapter(is_deleted);
CREATE INDEX idx_novel_chapter_has_mature_content ON novel_chapter(has_mature_content);

-- =============================================================================
-- COMMENTS FOR DOCUMENTATION
-- =============================================================================

-- Novel table comments
COMMENT ON COLUMN novel.created_by_user_id IS 'User ID của người tạo novel';
COMMENT ON COLUMN novel.updated_by_user_id IS 'User ID của người cập nhật novel lần cuối';
COMMENT ON COLUMN novel.tenant_id IS 'Tenant ID sở hữu novel (multi-tenancy)';
COMMENT ON COLUMN novel.published_at IS 'Thời điểm xuất bản chính thức novel';
COMMENT ON COLUMN novel.original_language IS 'Mã ngôn ngữ gốc của novel (ISO 639-1)';
COMMENT ON COLUMN novel.source_url IS 'URL nguồn gốc novel (nếu được chuyển thể)';
COMMENT ON COLUMN novel.isbn IS 'Mã ISBN cho xuất bản (nếu có)';
COMMENT ON COLUMN novel.age_rating IS 'Phân loại độ tuổi (G, PG, PG-13, R, NC-17)';
COMMENT ON COLUMN novel.content_warnings IS 'Cảnh báo nội dung dạng JSON (violence, sexual, etc)';
COMMENT ON COLUMN novel.mature_content IS 'Đánh dấu nội dung dành cho người lớn';
COMMENT ON COLUMN novel.is_public IS 'Novel có công khai hay chỉ riêng tư';
COMMENT ON COLUMN novel.is_featured IS 'Novel được đề xuất trên trang chủ';
COMMENT ON COLUMN novel.is_completed IS 'Novel đã hoàn thành';
COMMENT ON COLUMN novel.is_deleted IS 'Novel đã bị xóa (soft delete)';
COMMENT ON COLUMN novel.deleted_at IS 'Thời điểm xóa novel';
COMMENT ON COLUMN novel.deleted_by_user_id IS 'User ID thực hiện xóa';
COMMENT ON COLUMN novel.slug IS 'URL-friendly identifier cho SEO';
COMMENT ON COLUMN novel.tags IS 'Tags tìm kiếm dạng JSON array';
COMMENT ON COLUMN novel.keywords IS 'Keywords cho SEO';
COMMENT ON COLUMN novel.meta_description IS 'Meta description cho SEO';
COMMENT ON COLUMN novel.view_count IS 'Tổng số lượt xem novel';
COMMENT ON COLUMN novel.like_count IS 'Tổng số lượt thích';
COMMENT ON COLUMN novel.bookmark_count IS 'Tổng số lượt bookmark';
COMMENT ON COLUMN novel.comment_count IS 'Tổng số bình luận';
COMMENT ON COLUMN novel.rating_average IS 'Điểm đánh giá trung bình (0.00-5.00)';
COMMENT ON COLUMN novel.rating_count IS 'Tổng số lượt đánh giá';
COMMENT ON COLUMN novel.price_coins IS 'Giá mua toàn bộ series (virtual coins)';
COMMENT ON COLUMN novel.rental_price_coins IS 'Giá thuê series (virtual coins)';
COMMENT ON COLUMN novel.rental_duration_days IS 'Số ngày thuê series';
COMMENT ON COLUMN novel.is_premium IS 'Nội dung premium yêu cầu subscription';
COMMENT ON COLUMN novel.total_chapters IS 'Tổng số chương trong novel';
COMMENT ON COLUMN novel.total_volumes IS 'Tổng số tập trong novel';
COMMENT ON COLUMN novel.estimated_reading_time IS 'Thời gian đọc ước tính (phút)';
COMMENT ON COLUMN novel.word_count IS 'Tổng số từ trong toàn bộ novel';

-- Volume table comments
COMMENT ON COLUMN novel_volume.created_by_user_id IS 'User ID tạo volume';
COMMENT ON COLUMN novel_volume.updated_by_user_id IS 'User ID cập nhật volume';
COMMENT ON COLUMN novel_volume.published_at IS 'Thời điểm xuất bản volume';
COMMENT ON COLUMN novel_volume.is_deleted IS 'Volume đã bị xóa (soft delete)';
COMMENT ON COLUMN novel_volume.deleted_at IS 'Thời điểm xóa volume';
COMMENT ON COLUMN novel_volume.is_available IS 'Volume có sẵn để đọc';
COMMENT ON COLUMN novel_volume.page_count IS 'Số trang ước tính của volume';
COMMENT ON COLUMN novel_volume.word_count IS 'Số từ trong volume';
COMMENT ON COLUMN novel_volume.chapter_count IS 'Số chương trong volume';
COMMENT ON COLUMN novel_volume.estimated_reading_time IS 'Thời gian đọc ước tính volume (phút)';

-- Chapter table comments
COMMENT ON COLUMN novel_chapter.created_by_user_id IS 'User ID tạo chapter';
COMMENT ON COLUMN novel_chapter.updated_by_user_id IS 'User ID cập nhật chapter';
COMMENT ON COLUMN novel_chapter.scheduled_publish_at IS 'Thời gian lên lịch xuất bản';
COMMENT ON COLUMN novel_chapter.is_draft IS 'Chapter đang ở trạng thái draft';
COMMENT ON COLUMN novel_chapter.is_deleted IS 'Chapter đã bị xóa (soft delete)';
COMMENT ON COLUMN novel_chapter.deleted_at IS 'Thời điểm xóa chapter';
COMMENT ON COLUMN novel_chapter.version IS 'Version number của chapter';
COMMENT ON COLUMN novel_chapter.word_count IS 'Số từ trong chapter';
COMMENT ON COLUMN novel_chapter.reading_time_minutes IS 'Thời gian đọc ước tính (phút)';
COMMENT ON COLUMN novel_chapter.character_count IS 'Số ký tự trong chapter';
COMMENT ON COLUMN novel_chapter.view_count IS 'Số lượt xem chapter';
COMMENT ON COLUMN novel_chapter.like_count IS 'Số lượt thích chapter';
COMMENT ON COLUMN novel_chapter.comment_count IS 'Số bình luận chapter';
COMMENT ON COLUMN novel_chapter.content_warnings IS 'Cảnh báo nội dung riêng cho chapter';
COMMENT ON COLUMN novel_chapter.has_mature_content IS 'Chapter có nội dung nhạy cảm';