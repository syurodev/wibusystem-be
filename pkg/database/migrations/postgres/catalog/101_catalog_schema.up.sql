-- Catalog Service schema generated from catalog-db-design.md
-- Định nghĩa các enum dùng trong hệ thống
CREATE TYPE content_status AS ENUM ('ONGOING','COMPLETED','HIATUS');  -- trạng thái nội dung
CREATE TYPE season_name AS ENUM ('SPRING','SUMMER','FALL','WINTER');  -- mùa phát sóng anime
CREATE TYPE creator_role AS ENUM ('AUTHOR','ILLUSTRATOR','ARTIST','STUDIO','VOICE_ACTOR');  -- vai trò creator
CREATE TYPE content_type_enum AS ENUM ('ANIME','MANGA','NOVEL');  -- loại nội dung
CREATE TYPE content_relation_type AS ENUM ('ADAPTATION','SEQUEL','SPINOFF','RELATED');  -- loại liên kết nội dung
CREATE TYPE purchase_item_type AS ENUM
    ('ANIME_EPISODE','ANIME_SEASON',
     'MANGA_CHAPTER','MANGA_VOLUME','MANGA_SERIES',
     'NOVEL_CHAPTER','NOVEL_VOLUME','NOVEL_SERIES');  -- loại nội dung có thể mua
CREATE TYPE rental_item_type AS ENUM
    ('ANIME_SEASON','ANIME_SERIES',
     'MANGA_VOLUME','MANGA_SERIES',
     'NOVEL_VOLUME','NOVEL_SERIES');  -- loại nội dung cho thuê
CREATE TYPE subscription_tier AS ENUM ('FREE','PREMIUM','VIP');  -- hạng thuê bao

-- Bảng Anime (danh sách series anime)
CREATE TABLE anime (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    status content_status NOT NULL,         -- trạng thái: đang chiếu, hoàn thành, tạm ngưng
    cover_image TEXT,                       -- URL ảnh bìa
    broadcast_season season_name,           -- mùa phát sóng đầu tiên (Xuân/Hạ/Thu/Đông)
    broadcast_year INT                      -- năm phát sóng đầu tiên
);
COMMENT ON TABLE anime IS 'Anime series master table, one record per anime title.';
COMMENT ON COLUMN anime.status IS 'Current status of the anime (ongoing, completed, etc).';
COMMENT ON COLUMN anime.broadcast_season IS 'Broadcast season (quarter) of the anime''s premiere (Spring, Summer, Fall, Winter).';
COMMENT ON COLUMN anime.broadcast_year IS 'Broadcast year of the anime''s premiere.';

-- Bảng AnimeSeason (các mùa của anime)
CREATE TABLE anime_season (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    anime_id UUID NOT NULL REFERENCES anime(id) ON DELETE CASCADE,
    season_number INT NOT NULL,       -- số thứ tự mùa trong anime (1,2,...)
    season_title TEXT,               -- tiêu đề mùa (nếu có, ví dụ "Final Season")
    price_coins INT,                 -- giá mua trọn mùa (coins)
    rental_price_coins INT,          -- giá thuê mùa
    rental_duration_days INT,        -- thời gian thuê (ngày)
    UNIQUE(anime_id, season_number)
);
COMMENT ON TABLE anime_season IS 'Seasons of an anime series (e.g. Season 1, Season 2) grouping episodes.';
COMMENT ON COLUMN anime_season.season_number IS 'Sequential season number within the anime series.';
COMMENT ON COLUMN anime_season.price_coins IS 'Price (in virtual coins) to purchase this entire season permanently.';
COMMENT ON COLUMN anime_season.rental_price_coins IS 'Price (in coins) to rent this season for a limited time period.';
COMMENT ON COLUMN anime_season.rental_duration_days IS 'Duration of rental access for this season (in days).';

-- Bảng AnimeEpisode (các tập phim)
CREATE TABLE anime_episode (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    season_id UUID NOT NULL REFERENCES anime_season(id) ON DELETE CASCADE,
    episode_number INT NOT NULL,     -- số tập trong mùa
    title TEXT,                      -- tiêu đề tập (nếu có)
    duration_seconds INT,            -- độ dài tập (giây)
    video_url TEXT,                  -- URL video stream
    is_public BOOLEAN DEFAULT FALSE, -- có miễn phí hay không
    price_coins INT,                 -- giá mua lẻ tập
    UNIQUE(season_id, episode_number)
);
COMMENT ON TABLE anime_episode IS 'Individual episodes within an anime season.';
COMMENT ON COLUMN anime_episode.episode_number IS 'Episode number within the season.';
COMMENT ON COLUMN anime_episode.duration_seconds IS 'Duration of the episode in seconds.';
COMMENT ON COLUMN anime_episode.is_public IS 'Whether this episode is freely accessible (public) or requires purchase.';
COMMENT ON COLUMN anime_episode.price_coins IS 'Price (in coins) to purchase this episode.';

-- Bảng EpisodeSubtitle (phụ đề cho tập phim)
CREATE TABLE episode_subtitle (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    episode_id UUID NOT NULL REFERENCES anime_episode(id) ON DELETE CASCADE,
    language_code VARCHAR(10) NOT NULL,  -- mã ngôn ngữ (ví dụ 'en', 'vi')
    subtitle_url TEXT,                   -- URL file phụ đề (.srt, .vtt)
    UNIQUE(episode_id, language_code)
);
COMMENT ON TABLE episode_subtitle IS 'Subtitle file reference for an episode in various languages.';
COMMENT ON COLUMN episode_subtitle.language_code IS 'Language code for the subtitle track (e.g., en, vi, jp).';

-- Bảng Manga (danh sách series manga)
CREATE TABLE manga (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    status content_status NOT NULL,   -- trạng thái xuất bản
    cover_image TEXT                  -- URL ảnh bìa manga
);
COMMENT ON TABLE manga IS 'Manga series master table (comic).';
COMMENT ON COLUMN manga.status IS 'Current publication status of the manga (ongoing, completed, hiatus).';

-- Bảng MangaVolume (tập truyện của manga)
CREATE TABLE manga_volume (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    manga_id UUID NOT NULL REFERENCES manga(id) ON DELETE CASCADE,
    volume_number INT NOT NULL,      -- số tập (nếu manga có phân tập)
    volume_title TEXT,               -- tên tập (nếu có)
    cover_image TEXT,                -- ảnh bìa tập
    description TEXT,                -- mô tả ngắn về tập
    price_coins INT,                 -- giá mua trọn tập
    rental_price_coins INT,          -- giá thuê tập
    rental_duration_days INT,        -- thời hạn thuê tập (ngày)
    UNIQUE(manga_id, volume_number)
);
COMMENT ON TABLE manga_volume IS 'Volume grouping chapters of a manga series.';
COMMENT ON COLUMN manga_volume.volume_number IS 'Volume number (if official; a default volume may be used for uncollected chapters).';
COMMENT ON COLUMN manga_volume.price_coins IS 'Price to purchase this volume permanently.';
COMMENT ON COLUMN manga_volume.rental_price_coins IS 'Price to rent this volume temporarily.';
COMMENT ON COLUMN manga_volume.rental_duration_days IS 'Rental duration for this volume (in days).';

-- Bảng MangaChapter (các chương của manga)
CREATE TABLE manga_chapter (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    volume_id UUID NOT NULL REFERENCES manga_volume(id) ON DELETE CASCADE,
    chapter_number INT NOT NULL,    -- số thứ tự chương trong tập
    title TEXT,                     -- tiêu đề chương (nếu có)
    released_at TIMESTAMP,          -- thời điểm phát hành chương
    is_public BOOLEAN DEFAULT FALSE, -- có miễn phí không
    price_coins INT,               -- giá mua lẻ chương
    UNIQUE(volume_id, chapter_number)
);
COMMENT ON TABLE manga_chapter IS 'Chapter of a manga, containing multiple pages.';
COMMENT ON COLUMN manga_chapter.chapter_number IS 'Chapter number within the volume (or within the series if no formal volumes).';
COMMENT ON COLUMN manga_chapter.released_at IS 'Release timestamp of this chapter.';
COMMENT ON COLUMN manga_chapter.is_public IS 'Whether this chapter is free to read (public) or requires purchase.';
COMMENT ON COLUMN manga_chapter.price_coins IS 'Price to purchase this chapter.';

-- Bảng MangaPage (trang truyện của chapter)
CREATE TABLE manga_page (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    chapter_id UUID NOT NULL REFERENCES manga_chapter(id) ON DELETE CASCADE,
    page_number INT NOT NULL,      -- số thứ tự trang
    image_url TEXT,               -- URL ảnh của trang
    UNIQUE(chapter_id, page_number)
);
COMMENT ON TABLE manga_page IS 'Individual page (image) of a manga chapter.';
COMMENT ON COLUMN manga_page.page_number IS 'Page number within the chapter (for ordering).';
COMMENT ON COLUMN manga_page.image_url IS 'URL or file path of the image for this page.';

-- Bảng Novel (danh sách series novel)
CREATE TABLE novel (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    status content_status NOT NULL,   -- trạng thái (ongoing, completed, etc.)
    cover_image TEXT                  -- ảnh bìa novel
);
COMMENT ON TABLE novel IS 'Novel (text story) series master table.';
COMMENT ON COLUMN novel.status IS 'Current status of the novel (ongoing, completed, etc).';

-- Bảng NovelVolume (tập của novel)
CREATE TABLE novel_volume (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    novel_id UUID NOT NULL REFERENCES novel(id) ON DELETE CASCADE,
    volume_number INT NOT NULL,   -- số tập
    volume_title TEXT,            -- tên tập (nếu có)
    cover_image TEXT,             -- ảnh bìa tập
    description TEXT,             -- mô tả ngắn tập
    price_coins INT,              -- giá mua trọn tập
    rental_price_coins INT,       -- giá thuê tập
    rental_duration_days INT,     -- thời hạn thuê tập (ngày)
    UNIQUE(novel_id, volume_number)
);
COMMENT ON TABLE novel_volume IS 'Volume grouping chapters of a novel.';
COMMENT ON COLUMN novel_volume.volume_number IS 'Volume number in the novel series.';
COMMENT ON COLUMN novel_volume.price_coins IS 'Price to purchase this volume.';
COMMENT ON COLUMN novel_volume.rental_price_coins IS 'Price to rent this volume.';
COMMENT ON COLUMN novel_volume.rental_duration_days IS 'Rental duration for this volume (days).';

-- Bảng NovelChapter (các chương của novel)
CREATE TABLE novel_chapter (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    volume_id UUID NOT NULL REFERENCES novel_volume(id) ON DELETE CASCADE,
    chapter_number INT NOT NULL,   -- số chương trong tập
    title TEXT,                    -- tiêu đề chương
    content TEXT,                  -- nội dung văn bản của chương
    published_at TIMESTAMP,        -- thời gian xuất bản chương
    is_public BOOLEAN DEFAULT FALSE, -- có miễn phí không
    price_coins INT,               -- giá mua lẻ chương
    UNIQUE(volume_id, chapter_number)
);
COMMENT ON TABLE novel_chapter IS 'Chapter of a novel, containing text content.';
COMMENT ON COLUMN novel_chapter.chapter_number IS 'Chapter number within the volume.';
COMMENT ON COLUMN novel_chapter.content IS 'Full text content of the chapter.';
COMMENT ON COLUMN novel_chapter.published_at IS 'Publication timestamp of the chapter.';
COMMENT ON COLUMN novel_chapter.is_public IS 'Whether this chapter is free (public) or requires purchase.';
COMMENT ON COLUMN novel_chapter.price_coins IS 'Price to purchase this chapter.';

-- Bảng Character (nhân vật)
CREATE TABLE character (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name TEXT NOT NULL,        -- tên nhân vật
    description TEXT,          -- mô tả
    image_url TEXT             -- ảnh đại diện nhân vật
);
COMMENT ON TABLE character IS 'Fictional character that can appear in multiple content (anime/manga/novel).';
COMMENT ON COLUMN character.name IS 'Name of the character.';

-- Bảng Creator (nhà sáng tạo: tác giả, họa sĩ, studio, seiyuu, v.v.)
CREATE TABLE creator (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name TEXT NOT NULL,       -- tên creator (tên người hoặc tên nhóm/studio)
    description TEXT
);
COMMENT ON TABLE creator IS 'Content creator (author, artist, studio, voice actor, etc.) entity.';
COMMENT ON COLUMN creator.name IS 'Name of the creator (person or organization).';

-- Bảng Genre (thể loại nội dung)
CREATE TABLE genre (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name TEXT NOT NULL        -- tên thể loại
);
COMMENT ON TABLE genre IS 'Genre/category for content (shared across anime, manga, novel).';
COMMENT ON COLUMN genre.name IS 'Name of the genre.';

-- Bảng anime_character (liên kết anime và nhân vật, kèm diễn viên lồng tiếng)
CREATE TABLE anime_character (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    anime_id UUID NOT NULL REFERENCES anime(id) ON DELETE CASCADE,
    character_id UUID NOT NULL REFERENCES character(id) ON DELETE CASCADE,
    voice_actor_id UUID REFERENCES creator(id) ON DELETE SET NULL,  -- diễn viên lồng tiếng (có thể null nếu chưa biết)
    UNIQUE(anime_id, character_id)
);
COMMENT ON TABLE anime_character IS 'Maps characters to anime appearances (with optional voice actor casting).';
COMMENT ON COLUMN anime_character.voice_actor_id IS 'Voice actor (creator) who voices the character in this anime.';

-- Bảng manga_character (liên kết manga và nhân vật)
CREATE TABLE manga_character (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    manga_id UUID NOT NULL REFERENCES manga(id) ON DELETE CASCADE,
    character_id UUID NOT NULL REFERENCES character(id) ON DELETE CASCADE,
    UNIQUE(manga_id, character_id)
);
COMMENT ON TABLE manga_character IS 'Maps characters to manga appearances.';

-- Bảng novel_character (liên kết novel và nhân vật)
CREATE TABLE novel_character (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    novel_id UUID NOT NULL REFERENCES novel(id) ON DELETE CASCADE,
    character_id UUID NOT NULL REFERENCES character(id) ON DELETE CASCADE,
    UNIQUE(novel_id, character_id)
);
COMMENT ON TABLE novel_character IS 'Maps characters to novel appearances.';

-- Bảng anime_genre (liên kết anime và thể loại)
CREATE TABLE anime_genre (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    anime_id UUID NOT NULL REFERENCES anime(id) ON DELETE CASCADE,
    genre_id UUID NOT NULL REFERENCES genre(id),
    UNIQUE(anime_id, genre_id)
);
COMMENT ON TABLE anime_genre IS 'Associates anime with genres (many-to-many).';

-- Bảng manga_genre (liên kết manga và thể loại)
CREATE TABLE manga_genre (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    manga_id UUID NOT NULL REFERENCES manga(id) ON DELETE CASCADE,
    genre_id UUID NOT NULL REFERENCES genre(id),
    UNIQUE(manga_id, genre_id)
);
COMMENT ON TABLE manga_genre IS 'Associates manga with genres.';

-- Bảng novel_genre (liên kết novel và thể loại)
CREATE TABLE novel_genre (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    novel_id UUID NOT NULL REFERENCES novel(id) ON DELETE CASCADE,
    genre_id UUID NOT NULL REFERENCES genre(id),
    UNIQUE(novel_id, genre_id)
);
COMMENT ON TABLE novel_genre IS 'Associates novels with genres.';

-- Bảng anime_creator (liên kết anime và creator với vai trò)
CREATE TABLE anime_creator (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    anime_id UUID NOT NULL REFERENCES anime(id) ON DELETE CASCADE,
    creator_id UUID NOT NULL REFERENCES creator(id),
    role creator_role NOT NULL,
    UNIQUE(anime_id, creator_id, role)
);
COMMENT ON TABLE anime_creator IS 'Associates anime with creators (e.g., studio production company).';

-- Bảng manga_creator (liên kết manga và creator)
CREATE TABLE manga_creator (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    manga_id UUID NOT NULL REFERENCES manga(id) ON DELETE CASCADE,
    creator_id UUID NOT NULL REFERENCES creator(id),
    role creator_role NOT NULL,
    UNIQUE(manga_id, creator_id, role)
);
COMMENT ON TABLE manga_creator IS 'Associates manga with creators (e.g., author, artist).';

-- Bảng novel_creator (liên kết novel và creator)
CREATE TABLE novel_creator (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    novel_id UUID NOT NULL REFERENCES novel(id) ON DELETE CASCADE,
    creator_id UUID NOT NULL REFERENCES creator(id),
    role creator_role NOT NULL,
    UNIQUE(novel_id, creator_id, role)
);
COMMENT ON TABLE novel_creator IS 'Associates novel with creators (e.g., author, illustrator).';

-- Bảng content_relation (liên kết các nội dung với nhau: adaptation, sequel, etc.)
CREATE TABLE content_relation (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    source_id UUID NOT NULL,
    source_type content_type_enum NOT NULL,
    target_id UUID NOT NULL,
    target_type content_type_enum NOT NULL,
    relation_type content_relation_type NOT NULL
    -- Không đặt FOREIGN KEY trực tiếp do source/target có thể là nhiều bảng khác nhau
);
COMMENT ON TABLE content_relation IS 'Inter-content relationships between works (adaptation, sequel, spinoff, etc).';
COMMENT ON COLUMN content_relation.source_type IS 'Type of source content (ANIME, MANGA, NOVEL).';
COMMENT ON COLUMN content_relation.relation_type IS 'Relationship type (adaptation, sequel, spinoff, related).';

-- Index hỗ trợ tra cứu content_relation
CREATE INDEX idx_content_relation_source ON content_relation(source_type, source_id);
CREATE INDEX idx_content_relation_target ON content_relation(target_type, target_id);

-- Bảng anime_translation (tiêu đề/mô tả đa ngôn ngữ cho anime)
CREATE TABLE anime_translation (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    anime_id UUID NOT NULL REFERENCES anime(id) ON DELETE CASCADE,
    language_code VARCHAR(5) NOT NULL,  -- mã ngôn ngữ (ví dụ 'en', 'vi')
    title TEXT NOT NULL,                -- tiêu đề anime bằng ngôn ngữ này
    description TEXT,                   -- mô tả bằng ngôn ngữ này
    is_primary BOOLEAN DEFAULT FALSE,   -- đánh dấu tên chính
    UNIQUE(anime_id, language_code, title)  -- tránh trùng exact tiêu đề (có thể bỏ nếu cho phép)
);
-- Đảm bảo mỗi anime mỗi ngôn ngữ chỉ có một title chính thức
CREATE UNIQUE INDEX ux_anime_translation_primary ON anime_translation(anime_id, language_code) WHERE is_primary;
COMMENT ON TABLE anime_translation IS 'Localized titles and descriptions for anime in multiple languages.';
COMMENT ON COLUMN anime_translation.language_code IS 'Locale/language code of this translation.';
COMMENT ON COLUMN anime_translation.is_primary IS 'Indicates the primary title for the anime in this language.';

-- Bảng manga_translation (đa ngôn ngữ cho manga)
CREATE TABLE manga_translation (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    manga_id UUID NOT NULL REFERENCES manga(id) ON DELETE CASCADE,
    language_code VARCHAR(5) NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    is_primary BOOLEAN DEFAULT FALSE,
    UNIQUE(manga_id, language_code, title)
);
CREATE UNIQUE INDEX ux_manga_translation_primary ON manga_translation(manga_id, language_code) WHERE is_primary;
COMMENT ON TABLE manga_translation IS 'Localized titles and descriptions for manga in multiple languages.';

-- Bảng novel_translation (đa ngôn ngữ cho novel)
CREATE TABLE novel_translation (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    novel_id UUID NOT NULL REFERENCES novel(id) ON DELETE CASCADE,
    language_code VARCHAR(5) NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    is_primary BOOLEAN DEFAULT FALSE,
    UNIQUE(novel_id, language_code, title)
);
CREATE UNIQUE INDEX ux_novel_translation_primary ON novel_translation(novel_id, language_code) WHERE is_primary;
COMMENT ON TABLE novel_translation IS 'Localized titles and descriptions for novels in multiple languages.';

-- Bảng user_content_purchases (lịch sử mua nội dung của user)
CREATE TABLE user_content_purchases (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL,         -- tham chiếu user (từ service user)
    item_type purchase_item_type NOT NULL,  -- loại nội dung đã mua
    item_id UUID NOT NULL,         -- id của nội dung (tập/chương/volume...)
    purchase_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    -- (Không dùng FK cứng cho item_id vì có nhiều loại nội dung, xác thực qua ứng dụng)
);
COMMENT ON TABLE user_content_purchases IS 'Records of user purchases of content (episodes, chapters, volumes, etc).';
COMMENT ON COLUMN user_content_purchases.user_id IS 'ID of the user who made the purchase (from user service).';
COMMENT ON COLUMN user_content_purchases.item_type IS 'Type of content item purchased (e.g. NOVEL_CHAPTER, MANGA_VOLUME).';
COMMENT ON COLUMN user_content_purchases.item_id IS 'ID of the purchased content item.';
CREATE INDEX idx_user_purchase_user ON user_content_purchases(user_id);
CREATE INDEX idx_user_purchase_item ON user_content_purchases(item_type, item_id, user_id);

-- Bảng user_content_rentals (lịch sử thuê nội dung của user)
CREATE TABLE user_content_rentals (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL,
    item_type rental_item_type NOT NULL,  -- loại nội dung thuê
    item_id UUID NOT NULL,
    rent_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expiry_date TIMESTAMP NOT NULL       -- thời điểm hết hạn thuê
);
COMMENT ON TABLE user_content_rentals IS 'Records of user rentals of content with limited-time access.';
COMMENT ON COLUMN user_content_rentals.expiry_date IS 'Datetime when the rental access expires for the user.';
CREATE INDEX idx_user_rental_user ON user_content_rentals(user_id);
CREATE INDEX idx_user_rental_item ON user_content_rentals(item_type, item_id, user_id);

-- Bảng user_subscriptions (thuê bao của user)
CREATE TABLE user_subscriptions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL,
    tier subscription_tier NOT NULL,   -- hạng thuê bao (FREE/PREMIUM/VIP)
    start_date DATE NOT NULL,
    end_date DATE                     -- ngày hết hạn (null nếu đang active không kỳ hạn cố định)
);
COMMENT ON TABLE user_subscriptions IS 'User subscriptions (membership plans) for premium access/perks.';
COMMENT ON COLUMN user_subscriptions.tier IS 'Subscription plan tier (e.g., VIP for special perks).';
COMMENT ON COLUMN user_subscriptions.end_date IS 'End date of the subscription (if applicable).';
CREATE INDEX idx_user_subscription_user ON user_subscriptions(user_id);

-- Bảng user_donations (lịch sử donate của user)
CREATE TABLE user_donations (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL,
    content_type content_type_enum NOT NULL,  -- loại nội dung được donate (ANIME/MANGA/NOVEL)
    content_id UUID NOT NULL,
    amount DECIMAL(10,2) NOT NULL,           -- số tiền donate (giả sử đơn vị USD hoặc coin)
    donation_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
COMMENT ON TABLE user_donations IS 'Records of users donating to content creators (by content).';
COMMENT ON COLUMN user_donations.amount IS 'Donation amount (could be in platform currency).';
CREATE INDEX idx_user_donation_user ON user_donations(user_id);
CREATE INDEX idx_user_donation_content ON user_donations(content_type, content_id);
