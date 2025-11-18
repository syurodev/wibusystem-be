# Novel System - Enhanced Database Design

## Tổng quan

Hệ thống novel được thiết kế với các tính năng nâng cao:
- ✅ Cấu trúc 3 tầng: **Novel → Volume → Chapter**
- ✅ **Genres** riêng biệt với hỗ trợ phân cấp
- ✅ **Contributors** đa dạng: Authors, Artists, Translators
- ✅ **Translation System** với community contributions
- ✅ **JSONB** cho rich content
- ✅ **Version Control** cho translations

## Kiến trúc Database

### Schema: `catalog`

```
catalog.genres (Thể loại)
    ├── Hỗ trợ phân cấp (parent_id)
    └── Many-to-many với novels qua novel_genres

catalog.authors (Tác giả)
    ├── Link với user account (user_id)
    ├── Biography, avatar, social links
    ├── Statistics: novel_count, views, followers
    └── Many-to-many với novels qua novel_authors

catalog.artists (Hoạ sĩ)
    ├── Link với user account
    ├── Specialization: cover_artist, illustrator, etc.
    ├── Statistics: novel_count, artwork_count
    └── Many-to-many với novels qua novel_artists

catalog.translators (Người dịch)
    ├── Link với user account
    ├── Supported languages
    ├── Rating system (0-5)
    ├── Statistics: chapter_count, word_count, contributions
    └── Many-to-many với novels qua novel_translators

catalog.novels (Tiểu thuyết)
    ├── Không còn author_id trực tiếp
    ├── Relationships qua junction tables
    ├── original_language, original_title
    └── Synopsis và metadata (JSONB)

catalog.volumes (Tập)
    └── Giữ nguyên như cũ

catalog.chapters (Chương)
    ├── Giữ nguyên như cũ
    └── Có thể có nhiều translations

catalog.chapter_translations (Bản dịch chính thức)
    ├── Một translation cho mỗi (chapter, language)
    ├── Version tracking
    ├── Status: draft, pending_review, approved, rejected, published
    ├── Translator và quality metrics
    └── History tracking

catalog.translation_contributions (Đóng góp từ community)
    ├── Contributions từ users
    ├── Types: new_translation, improvement, proofreading, correction
    ├── Review workflow
    ├── Credit points system
    └── Community voting (upvote/downvote)

catalog.translation_history (Lịch sử thay đổi)
    └── Track all versions of translations
```

## Chi tiết Tables

### 1. Genres System

#### `catalog.genres`
```sql
- id (UUID, PK)
- name, slug
- description
- parent_id (FK → genres, nullable) -- Phân cấp
- display_order, is_active
- timestamps
```

**Ví dụ phân cấp:**
```
Fantasy (parent)
├── Xuanhuan (child)
├── Xianxia (child)
└── Wuxia (child)

Romance (parent)
├── School Life (child)
└── Historical (child)
```

#### `catalog.novel_genres` (Junction Table)
```sql
- novel_id (FK → novels)
- genre_id (FK → genres)
- display_order
- PK: (novel_id, genre_id)
```

### 2. Contributors System

#### `catalog.authors`
```sql
- id (UUID, PK)
- user_id (FK → users, nullable)
- name, slug
- biography (JSONB)
- avatar_url
- social_links (JSONB) -- {"facebook": "...", "twitter": "..."}
- Statistics: novel_count, total_chapters, total_views, follower_count
- is_verified
- timestamps (created_at, updated_at, deleted_at)
```

#### `catalog.artists`
```sql
- id (UUID, PK)
- user_id (FK → users, nullable)
- name, slug
- biography (JSONB)
- avatar_url, social_links (JSONB)
- specialization VARCHAR(50) -- "cover_artist", "illustrator", "manga_artist"
- Statistics: novel_count, artwork_count, follower_count
- is_verified
- timestamps
```

#### `catalog.translators`
```sql
- id (UUID, PK)
- user_id (FK → users, nullable)
- name, slug
- biography (JSONB)
- avatar_url
- languages (JSONB) -- ["vi", "en", "zh", "ja", "ko"]
- Statistics: novel_count, chapter_count, word_count, contribution_count, follower_count
- rating_average DECIMAL(3,2), rating_count
- is_verified
- timestamps
```

#### Junction Tables

**`catalog.novel_authors`**
```sql
- novel_id, author_id
- role VARCHAR(50) -- "original_author", "co_author"
- display_order
- PK: (novel_id, author_id)
```

**`catalog.novel_artists`**
```sql
- novel_id, artist_id
- role VARCHAR(50) -- "cover_artist", "illustrator", "character_designer"
- display_order
- PK: (novel_id, artist_id, role)
```

**`catalog.novel_translators`**
```sql
- novel_id, translator_id
- target_language VARCHAR(10)
- role VARCHAR(50) -- "lead_translator", "translator", "proofreader"
- display_order
- PK: (novel_id, translator_id, target_language)
```

### 3. Translation System

#### `catalog.chapter_translations`
```sql
- id (UUID, PK)
- chapter_id (FK → chapters)
- language VARCHAR(10) -- ISO 639-1: "vi", "en", "zh"
- title, content (JSONB)
- translator_notes (JSONB)
- translator_id (FK → translators)
- version INTEGER -- Auto-increment on update
- status ENUM -- draft, pending_review, approved, rejected, published
- word_count, character_count
- Statistics: view_count, like_count, rating_average, rating_count
- published_at
- timestamps
- UNIQUE: (chapter_id, language)
```

**Content JSONB Format:**
```json
{
  "version": "1.0",
  "blocks": [
    {"type": "paragraph", "content": "Nội dung bản dịch..."},
    {"type": "dialogue", "speaker": "Nhân vật", "content": "Lời thoại..."}
  ]
}
```

#### `catalog.translation_contributions`
```sql
- id (UUID, PK)
- chapter_id (FK → chapters)
- contributor_id (FK → users)
- language VARCHAR(10)
- contribution_type ENUM -- new_translation, improvement, proofreading, correction
- title, content (JSONB)
- contributor_notes TEXT
- status ENUM -- draft, pending_review, approved, rejected, published
- Review: reviewed_by, reviewed_at, review_notes
- official_translation_id (FK → chapter_translations, nullable)
- Credit: credit_points, is_credited
- Metrics: word_count, character_count
- Community: upvote_count, downvote_count
- timestamps
```

**Contribution Types:**
- `new_translation`: Bản dịch hoàn toàn mới
- `improvement`: Cải thiện bản dịch hiện tại
- `proofreading`: Đọc và chỉnh sửa lỗi
- `correction`: Sửa lỗi cụ thể

#### `catalog.translation_history`
```sql
- id (UUID, PK)
- translation_id (FK → chapter_translations)
- version INTEGER
- title, content (JSONB) -- Snapshot at this version
- changed_by (FK → users)
- change_description TEXT
- word_count
- created_at
```

## Enum Types

```sql
-- Novel status (existing)
CREATE TYPE catalog.novel_status AS ENUM (
    'draft', 'ongoing', 'completed', 'hiatus', 'dropped'
);

-- Chapter status (existing)
CREATE TYPE catalog.chapter_status AS ENUM (
    'draft', 'published', 'scheduled'
);

-- Translation status (NEW)
CREATE TYPE catalog.translation_status AS ENUM (
    'draft', 'pending_review', 'approved', 'rejected', 'published'
);

-- Contribution type (NEW)
CREATE TYPE catalog.contribution_type AS ENUM (
    'new_translation', 'improvement', 'proofreading', 'correction'
);
```

## Triggers và Auto-updates

### 1. Author Statistics
```sql
-- Tự động update author.novel_count khi novel_authors thay đổi
trg_novel_authors_update_stats → update_author_novel_count()
```

### 2. Translator Statistics
```sql
-- Tự động update translator stats khi translations thay đổi
trg_chapter_translations_update_stats → update_translator_stats()
```

### 3. Translation Versioning
```sql
-- Tự động tạo history khi translation content thay đổi
trg_chapter_translations_history → create_translation_history()
```

### 4. Novel Statistics (existing)
```sql
-- Auto-update total_volumes, total_chapters, total_words
trg_volumes_update_novel_stats → update_novel_volume_stats()
trg_chapters_update_stats → update_chapter_stats()
```

## Indexes

### Genres Indexes
```sql
idx_genres_parent_id, idx_genres_slug, idx_genres_active
idx_novel_genres_genre_id, idx_novel_genres_novel_id
```

### Contributors Indexes
```sql
idx_authors_user_id, idx_authors_slug, idx_authors_verified
idx_artists_user_id, idx_artists_slug
idx_translators_user_id, idx_translators_slug, idx_translators_languages (GIN)
```

### Translation Indexes
```sql
idx_chapter_translations_chapter_id, idx_chapter_translations_language
idx_chapter_translations_translator_id, idx_chapter_translations_status
idx_chapter_translations_content (GIN) -- For full-text search
idx_translation_contributions_chapter_id, idx_translation_contributions_contributor_id
idx_translation_contributions_status, idx_translation_contributions_reviewed_by
```

## Use Cases

### 1. Tạo Novel với Authors và Genres

```go
// 1. Tạo author
author := &domain.Author{
    ID:   uuid.Must(uuid.NewV4()),
    Name: "Thiên Tằm Thổ Đậu",
    Slug: "thien-tam-tho-dau",
    Biography: synopsisJSON,
    IsVerified: true,
}
authorRepo.Create(ctx, author)

// 2. Tạo novel
novel := &domain.Novel{
    ID:               uuid.Must(uuid.NewV4()),
    Title:            "Đấu Phá Thương Khung",
    Slug:             "dau-pha-thuong-khung",
    OriginalLanguage: strPtr("zh"),
    OriginalTitle:    strPtr("斗破苍穹"),
    Status:           domain.NovelStatusOngoing,
}
novelRepo.Create(ctx, novel)

// 3. Link author với novel
authorRepo.AddNovelAuthor(ctx, novel.ID, author.ID, "original_author", 0)

// 4. Thêm genres
fantasyID := getGenreBySlug("fantasy")
xuanhuanID := getGenreBySlug("xuanhuan")
genreRepo.UpdateNovelGenres(ctx, novel.ID, []uuid.UUID{fantasyID, xuanhuanID})
```

### 2. Thêm Translator và Artist

```go
// Translator
translator := &domain.Translator{
    ID:        uuid.Must(uuid.NewV4()),
    Name:      "Hội Dịch Thuật",
    Slug:      "hoi-dich-thuat",
    Languages: json.RawMessage(`["vi", "en"]`),
}
translatorRepo.Create(ctx, translator)

// Link với novel
translatorRepo.AddNovelTranslator(ctx, novel.ID, translator.ID, "vi", "lead_translator", 0)

// Artist
artist := &domain.Artist{
    ID:             uuid.Must(uuid.NewV4()),
    Name:           "Nguyễn Văn A",
    Slug:           "nguyen-van-a",
    Specialization: strPtr("cover_artist"),
}
artistRepo.Create(ctx, artist)
artistRepo.AddNovelArtist(ctx, novel.ID, artist.ID, "cover_artist", 0)
```

### 3. Tạo Official Translation

```go
contentJSON := map[string]interface{}{
    "version": "1.0",
    "blocks": []map[string]interface{}{
        {
            "type":    "paragraph",
            "content": "Trên núi Thanh Vân, sương mù bao phủ...",
        },
    },
}
content, _ := json.Marshal(contentJSON)

translation := &domain.ChapterTranslation{
    ID:           uuid.Must(uuid.NewV4()),
    ChapterID:    chapterID,
    Language:     "vi",
    Title:        "Chương 1: Khởi đầu",
    Content:      content,
    TranslatorID: &translatorID,
    Status:       domain.TranslationStatusDraft,
    WordCount:    calculateWordCount(content),
}
translationRepo.Create(ctx, translation)

// Publish khi ready
translationRepo.Publish(ctx, translation.ID)
```

### 4. Community Contribution Workflow

```go
// User đóng góp bản dịch
contribution := &domain.TranslationContribution{
    ID:               uuid.Must(uuid.NewV4()),
    ChapterID:        chapterID,
    ContributorID:    userID,
    Language:         "vi",
    ContributionType: domain.ContributionTypeImprovement,
    Content:          improvedContentJSON,
    ContributorNotes: strPtr("Cải thiện ngữ pháp và dịch chính xác hơn"),
    Status:           domain.TranslationStatusPendingReview,
}
contributionRepo.Create(ctx, contribution)

// Moderator review
pendingContributions, _ := contributionRepo.GetPendingReview(ctx, strPtr("vi"), 50, 0)

// Approve contribution
contributionRepo.Approve(ctx, contribution.ID, moderatorID, "Bản dịch tốt!", 100)

// Update official translation with approved content (manually)
officialTranslation.Content = contribution.Content
officialTranslation.Version++
translationRepo.Update(ctx, officialTranslation)
```

### 5. Query Novels với Filters

```go
// Tìm novels theo genre và author
status := domain.NovelStatusOngoing
filter := domain.NovelFilter{
    Status:           &status,
    GenreIDs:         []uuid.UUID{xuanhuanGenreID, fantasyGenreID},
    AuthorID:         &authorID,
    OriginalLanguage: strPtr("zh"),
    SearchQuery:      strPtr("tu tiên"),
    SortBy:           "rating",
    SortOrder:        "desc",
    Limit:            20,
    Offset:           0,
}

novels, total, err := novelRepo.List(ctx, filter)
```

## Migration Steps

### 1. Run Migrations

```bash
# Check current version (should be 12)
make migrate-version

# Run new migrations
make migrate-up

# Verify version (should be 14)
make migrate-version
```

### 2. Verify Tables

```sql
-- Check all catalog tables
\dt catalog.*

-- Expected tables:
-- novels, volumes, chapters (existing)
-- genres, novel_genres (new)
-- authors, artists, translators (new)
-- novel_authors, novel_artists, novel_translators (new)
-- chapter_translations, translation_contributions, translation_history (new)
```

### 3. Seed Example Data

```sql
-- Create genres
INSERT INTO catalog.genres (id, name, slug, display_order, is_active) VALUES
(gen_random_uuid(), 'Fantasy', 'fantasy', 1, true),
(gen_random_uuid(), 'Xuanhuan', 'xuanhuan', 2, true),
(gen_random_uuid(), 'Romance', 'romance', 3, true);

-- Create author
INSERT INTO catalog.authors (id, name, slug, is_verified) VALUES
(gen_random_uuid(), 'Thiên Tằm Thổ Đậu', 'thien-tam-tho-dau', true);

-- Link novel với author (if you have existing novel)
INSERT INTO catalog.novel_authors (novel_id, author_id, role, display_order)
SELECT
    n.id,
    a.id,
    'original_author',
    0
FROM catalog.novels n, catalog.authors a
WHERE n.slug = 'your-novel-slug' AND a.slug = 'thien-tam-tho-dau';
```

## API Endpoints (Examples)

```
# Genres
GET    /api/v1/genres                    - List all genres
GET    /api/v1/genres/:id                - Get genre details
GET    /api/v1/genres/:id/novels         - Get novels in genre

# Authors
GET    /api/v1/authors                   - List authors
GET    /api/v1/authors/:id               - Get author profile
GET    /api/v1/authors/:id/novels        - Get author's novels

# Translators
GET    /api/v1/translators               - List translators
GET    /api/v1/translators/:id           - Get translator profile
GET    /api/v1/translators/:id/translations - Get translator's work

# Translations
GET    /api/v1/chapters/:id/translations - Get all translations of chapter
GET    /api/v1/translations/:id          - Get specific translation
POST   /api/v1/translations              - Create official translation (translator only)
PUT    /api/v1/translations/:id          - Update translation
POST   /api/v1/translations/:id/publish  - Publish translation

# Community Contributions
GET    /api/v1/contributions/pending     - Get pending contributions (moderator)
GET    /api/v1/chapters/:id/contributions - Get contributions for chapter
POST   /api/v1/contributions             - Submit contribution (any user)
POST   /api/v1/contributions/:id/approve - Approve contribution (moderator)
POST   /api/v1/contributions/:id/reject  - Reject contribution (moderator)
POST   /api/v1/contributions/:id/vote    - Vote on contribution
```

## Best Practices

### 1. Translation Workflow

1. Translator tạo draft translation
2. Translator hoàn thiện và submit for review
3. Moderator/Editor review và approve
4. Translator publish translation
5. Community có thể submit improvements
6. Moderator review improvements và merge vào official translation

### 2. Credit System

- New translation: 100-500 points (tùy độ dài)
- Improvement accepted: 50-200 points
- Proofreading accepted: 20-100 points
- Correction accepted: 10-50 points

### 3. Quality Control

- Translators được rate bởi community (0-5 stars)
- Contributions được vote bởi community
- Moderators review contributions trước khi approve
- Version history cho phép rollback nếu cần

## Troubleshooting

### Issue: Foreign key constraint fails khi migrate

```sql
-- Check if there are novels with author_id referencing users
SELECT id, title, author_id FROM catalog.novels WHERE author_id IS NOT NULL;

-- Option 1: Migrate existing data
-- Create authors từ user IDs và link qua novel_authors

-- Option 2: Allow NULL author_id temporarily
ALTER TABLE catalog.novels ALTER COLUMN author_id DROP NOT NULL;
```

### Issue: Duplicate translations

```sql
-- Check for duplicates
SELECT chapter_id, language, COUNT(*)
FROM catalog.chapter_translations
GROUP BY chapter_id, language
HAVING COUNT(*) > 1;

-- Fix: Keep only the latest version
DELETE FROM catalog.chapter_translations t1
WHERE id NOT IN (
    SELECT id FROM catalog.chapter_translations t2
    WHERE t1.chapter_id = t2.chapter_id AND t1.language = t2.language
    ORDER BY version DESC, created_at DESC LIMIT 1
);
```

## Resources

- [PostgreSQL JSONB](https://www.postgresql.org/docs/current/datatype-json.html)
- [ISO 639-1 Language Codes](https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes)
- [Version Control Best Practices](https://www.postgresql.org/docs/current/ddl-rowsecurity.html)

---

**Status:** ✅ Enhanced database schema with genres, contributors, and translations
**Migrations:** 000012, 000013, 000014
**Date:** 2025-11-18
