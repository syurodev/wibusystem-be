# Novel Database System - Complete Documentation

## 📋 Mục lục

1. [Tổng quan](#tổng-quan)
2. [Kiến trúc Database](#kiến-trúc-database)
3. [Chi tiết Tables](#chi-tiết-tables)
4. [JSONB Structure Examples](#jsonb-structure-examples)
5. [Migrations](#migrations)
6. [Code Examples](#code-examples)
7. [Setup Guide](#setup-guide)
8. [API Endpoints](#api-endpoints)
9. [Best Practices](#best-practices)
10. [Troubleshooting](#troubleshooting)

---

## Tổng quan

Hệ thống database cho novel được thiết kế với các tính năng:

### ✅ Core Features
- **3-tier Structure**: Novel → Volume → Chapter
- **JSONB Content**: Rich content format cho synopsis, content, notes
- **Soft Delete**: Tất cả tables hỗ trợ soft delete
- **Auto Statistics**: Triggers tự động cập nhật thống kê
- **Multi-language**: Hỗ trợ đa ngôn ngữ

### ✅ Advanced Features
- **Genres System**: Thể loại với phân cấp (parent-child)
- **Contributors**: Authors, Artists, Translators riêng biệt
- **Translation System**: Official translations + Community contributions
- **Version Control**: Translation history cho rollback
- **Review Workflow**: Approval process cho contributions
- **Credit System**: Points và rewards cho contributors

### 📦 Schema
Tất cả tables nằm trong schema `catalog`

---

## Kiến trúc Database

```
catalog Schema
│
├── Core Structure (3-tier)
│   ├── novels (top level)
│   ├── volumes (middle level)
│   └── chapters (bottom level)
│
├── Genres System
│   ├── genres (hierarchical)
│   └── novel_genres (junction)
│
├── Contributors System
│   ├── authors
│   ├── artists
│   ├── translators
│   ├── novel_authors (junction)
│   ├── novel_artists (junction)
│   └── novel_translators (junction)
│
└── Translation System
    ├── chapter_translations (official)
    ├── translation_contributions (community)
    └── translation_history (version control)
```

---

## Chi tiết Tables

### 1. Core Tables (3-tier)

#### `catalog.novels`

Lưu trữ thông tin novel (cấp cao nhất).

**Schema:**
```sql
CREATE TABLE catalog.novels (
    id UUID PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(500) NOT NULL UNIQUE,

    -- Synopsis dạng JSONB
    synopsis JSONB,

    cover_image_url VARCHAR(1000),
    thumbnail_url VARCHAR(1000),

    status catalog.novel_status NOT NULL DEFAULT 'draft',

    -- Original info
    original_language VARCHAR(10),
    original_title VARCHAR(500),

    -- Statistics (auto-updated by triggers)
    total_volumes INTEGER DEFAULT 0,
    total_chapters INTEGER DEFAULT 0,
    total_words BIGINT DEFAULT 0,
    view_count BIGINT DEFAULT 0,
    favorite_count INTEGER DEFAULT 0,
    rating_average DECIMAL(3,2) DEFAULT 0.00,
    rating_count INTEGER DEFAULT 0,

    -- Metadata cho additional info
    metadata JSONB DEFAULT '{}',

    -- Publishing dates
    first_published_at TIMESTAMP WITH TIME ZONE,
    last_chapter_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,

    -- Audit
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);
```

**Relationships:**
- Authors via `novel_authors` (many-to-many)
- Artists via `novel_artists` (many-to-many)
- Translators via `novel_translators` (many-to-many)
- Genres via `novel_genres` (many-to-many)

#### `catalog.volumes`

Tổ chức chapters thành volumes (cấp giữa).

**Schema:**
```sql
CREATE TABLE catalog.volumes (
    id UUID PRIMARY KEY,
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,

    volume_number INTEGER NOT NULL,
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(500) NOT NULL,
    description TEXT,

    cover_image_url VARCHAR(1000),

    -- Statistics (auto-updated)
    chapter_count INTEGER DEFAULT 0,
    word_count BIGINT DEFAULT 0,

    display_order INTEGER NOT NULL,
    is_published BOOLEAN DEFAULT FALSE,

    published_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    UNIQUE(novel_id, volume_number),
    UNIQUE(novel_id, slug)
);
```

**Note:** Chapters có thể tồn tại độc lập (không cần volume).

#### `catalog.chapters`

Lưu trữ nội dung chapters (cấp thấp nhất).

**Schema:**
```sql
CREATE TABLE catalog.chapters (
    id UUID PRIMARY KEY,
    novel_id UUID NOT NULL REFERENCES catalog.novels(id) ON DELETE CASCADE,
    volume_id UUID REFERENCES catalog.volumes(id) ON DELETE SET NULL,

    chapter_number INTEGER NOT NULL,
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(500) NOT NULL,

    -- Content dạng JSONB
    content JSONB NOT NULL,

    word_count INTEGER DEFAULT 0,
    character_count INTEGER DEFAULT 0,

    -- Access control
    is_free BOOLEAN DEFAULT TRUE,
    price DECIMAL(10,2) DEFAULT 0.00,
    currency VARCHAR(3) DEFAULT 'VND',

    status catalog.chapter_status DEFAULT 'draft',

    -- Statistics
    view_count BIGINT DEFAULT 0,
    like_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,

    display_order INTEGER NOT NULL,

    -- Author notes dạng JSONB
    author_notes JSONB,

    published_at TIMESTAMP WITH TIME ZONE,
    scheduled_at TIMESTAMP WITH TIME ZONE,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    UNIQUE(novel_id, chapter_number),
    UNIQUE(novel_id, slug)
);
```

---

### 2. Genres System

#### `catalog.genres`

Thể loại với hỗ trợ phân cấp.

**Schema:**
```sql
CREATE TABLE catalog.genres (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,

    -- Hierarchical support
    parent_id UUID REFERENCES catalog.genres(id) ON DELETE SET NULL,

    display_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

**Ví dụ phân cấp:**
```
Fantasy (root)
├── Xuanhuan (child of Fantasy)
├── Xianxia (child of Fantasy)
└── Wuxia (child of Fantasy)

Romance (root)
├── School Life (child of Romance)
└── Historical (child of Romance)
```

#### `catalog.novel_genres`

Junction table: novels ↔ genres (many-to-many).

```sql
CREATE TABLE catalog.novel_genres (
    novel_id UUID REFERENCES catalog.novels(id) ON DELETE CASCADE,
    genre_id UUID REFERENCES catalog.genres(id) ON DELETE CASCADE,
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (novel_id, genre_id)
);
```

---

### 3. Contributors System

#### `catalog.authors`

Tác giả (separate from users).

**Schema:**
```sql
CREATE TABLE catalog.authors (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) NOT NULL UNIQUE,

    biography JSONB,
    avatar_url VARCHAR(1000),
    social_links JSONB DEFAULT '{}',

    -- Statistics
    novel_count INTEGER DEFAULT 0,
    total_chapters INTEGER DEFAULT 0,
    total_views BIGINT DEFAULT 0,
    follower_count INTEGER DEFAULT 0,

    is_verified BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);
```

**Social Links Example:**
```json
{
  "facebook": "https://facebook.com/author",
  "twitter": "@authorname",
  "website": "https://author.com",
  "weibo": "authorname"
}
```

#### `catalog.artists`

Hoạ sĩ/minh họa.

**Schema:**
```sql
CREATE TABLE catalog.artists (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) NOT NULL UNIQUE,

    biography JSONB,
    avatar_url VARCHAR(1000),
    social_links JSONB DEFAULT '{}',

    specialization VARCHAR(50), -- cover_artist, illustrator, manga_artist

    -- Statistics
    novel_count INTEGER DEFAULT 0,
    artwork_count INTEGER DEFAULT 0,
    follower_count INTEGER DEFAULT 0,

    is_verified BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);
```

#### `catalog.translators`

Người dịch.

**Schema:**
```sql
CREATE TABLE catalog.translators (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES identify.users(id) ON DELETE SET NULL,

    name VARCHAR(200) NOT NULL,
    slug VARCHAR(200) NOT NULL UNIQUE,

    biography JSONB,
    avatar_url VARCHAR(1000),

    languages JSONB NOT NULL DEFAULT '[]', -- ["vi", "en", "zh"]

    -- Statistics
    novel_count INTEGER DEFAULT 0,
    chapter_count INTEGER DEFAULT 0,
    word_count BIGINT DEFAULT 0,
    contribution_count INTEGER DEFAULT 0,
    follower_count INTEGER DEFAULT 0,

    -- Quality rating
    rating_average DECIMAL(3,2) DEFAULT 0.00,
    rating_count INTEGER DEFAULT 0,

    is_verified BOOLEAN DEFAULT FALSE,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);
```

#### Junction Tables

**`catalog.novel_authors`**
```sql
CREATE TABLE catalog.novel_authors (
    novel_id UUID REFERENCES catalog.novels(id) ON DELETE CASCADE,
    author_id UUID REFERENCES catalog.authors(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'author', -- original_author, co_author
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (novel_id, author_id)
);
```

**`catalog.novel_artists`**
```sql
CREATE TABLE catalog.novel_artists (
    novel_id UUID REFERENCES catalog.novels(id) ON DELETE CASCADE,
    artist_id UUID REFERENCES catalog.artists(id) ON DELETE CASCADE,
    role VARCHAR(50) DEFAULT 'illustrator', -- cover_artist, illustrator, character_designer
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (novel_id, artist_id, role)
);
```

**`catalog.novel_translators`**
```sql
CREATE TABLE catalog.novel_translators (
    novel_id UUID REFERENCES catalog.novels(id) ON DELETE CASCADE,
    translator_id UUID REFERENCES catalog.translators(id) ON DELETE CASCADE,
    target_language VARCHAR(10) NOT NULL, -- vi, en, zh
    role VARCHAR(50) DEFAULT 'translator', -- lead_translator, translator, proofreader
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (novel_id, translator_id, target_language)
);
```

---

### 4. Translation System

#### `catalog.chapter_translations`

Bản dịch chính thức của chapters.

**Schema:**
```sql
CREATE TABLE catalog.chapter_translations (
    id UUID PRIMARY KEY,
    chapter_id UUID NOT NULL REFERENCES catalog.chapters(id) ON DELETE CASCADE,

    language VARCHAR(10) NOT NULL, -- ISO 639-1

    title VARCHAR(500) NOT NULL,
    content JSONB NOT NULL,
    translator_notes JSONB,

    translator_id UUID REFERENCES catalog.translators(id) ON DELETE SET NULL,

    version INTEGER DEFAULT 1, -- Auto-increment on update
    status catalog.translation_status DEFAULT 'draft',

    word_count INTEGER DEFAULT 0,
    character_count INTEGER DEFAULT 0,

    -- Statistics
    view_count BIGINT DEFAULT 0,
    like_count INTEGER DEFAULT 0,
    rating_average DECIMAL(3,2) DEFAULT 0.00,
    rating_count INTEGER DEFAULT 0,

    published_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    UNIQUE(chapter_id, language)
);
```

**Status Flow:**
```
draft → pending_review → approved → published
                      ↘ rejected
```

#### `catalog.translation_contributions`

Đóng góp bản dịch từ community.

**Schema:**
```sql
CREATE TABLE catalog.translation_contributions (
    id UUID PRIMARY KEY,
    chapter_id UUID NOT NULL REFERENCES catalog.chapters(id) ON DELETE CASCADE,
    contributor_id UUID NOT NULL REFERENCES identify.users(id) ON DELETE CASCADE,

    language VARCHAR(10) NOT NULL,
    contribution_type catalog.contribution_type NOT NULL,

    title VARCHAR(500),
    content JSONB NOT NULL,
    contributor_notes TEXT,

    status catalog.translation_status DEFAULT 'pending_review',

    -- Review
    reviewed_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMP WITH TIME ZONE,
    review_notes TEXT,

    -- Link to official translation if approved
    official_translation_id UUID REFERENCES catalog.chapter_translations(id),

    -- Credits
    credit_points INTEGER DEFAULT 0,
    is_credited BOOLEAN DEFAULT FALSE,

    word_count INTEGER DEFAULT 0,
    character_count INTEGER DEFAULT 0,

    -- Community feedback
    upvote_count INTEGER DEFAULT 0,
    downvote_count INTEGER DEFAULT 0,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);
```

**Contribution Types:**
- `new_translation`: Bản dịch hoàn toàn mới
- `improvement`: Cải thiện bản dịch hiện tại
- `proofreading`: Đọc và sửa lỗi
- `correction`: Sửa lỗi cụ thể

#### `catalog.translation_history`

Lịch sử thay đổi (version control).

**Schema:**
```sql
CREATE TABLE catalog.translation_history (
    id UUID PRIMARY KEY,
    translation_id UUID NOT NULL REFERENCES catalog.chapter_translations(id) ON DELETE CASCADE,

    version INTEGER NOT NULL,

    -- Snapshot
    title VARCHAR(500) NOT NULL,
    content JSONB NOT NULL,

    changed_by UUID REFERENCES identify.users(id) ON DELETE SET NULL,
    change_description TEXT,
    word_count INTEGER DEFAULT 0,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

---

## JSONB Structure Examples

### Synopsis Format

```json
{
  "language": "vi",
  "blocks": [
    {
      "type": "paragraph",
      "content": "Trong thế giới tu tiên đầy nguy hiểm, Tiêu Viêm..."
    },
    {
      "type": "paragraph",
      "content": "Với ý chí kiên cường và tài năng thiên phú..."
    }
  ]
}
```

### Chapter Content Format

```json
{
  "version": "1.0",
  "blocks": [
    {
      "type": "heading",
      "level": 2,
      "content": "Chương 1: Khởi đầu"
    },
    {
      "type": "paragraph",
      "content": "Trên núi Thanh Vân, sương mù bao phủ..."
    },
    {
      "type": "dialogue",
      "speaker": "Tiêu Viêm",
      "content": "Ta sẽ trở thành người mạnh nhất!"
    },
    {
      "type": "image",
      "url": "https://example.com/image.jpg",
      "caption": "Minh họa núi Thanh Vân"
    }
  ]
}
```

### Author Biography Format

```json
{
  "language": "vi",
  "blocks": [
    {
      "type": "paragraph",
      "content": "Thiên Tằm Thổ Đậu là tác giả nổi tiếng với các tác phẩm..."
    }
  ],
  "awards": [
    "Best Fantasy Novel 2023",
    "Most Popular Author 2022"
  ]
}
```

### Metadata Format

```json
{
  "external_ids": {
    "myanimelist": "12345",
    "novelupdates": "dau-pha-thuong-khung",
    "qidian": "123456789"
  },
  "custom_fields": {
    "alternate_names": ["Battle Through the Heavens", "BTTH"],
    "adaptations": ["anime", "manhua", "donghua"]
  }
}
```

---

## Migrations

### Migration Files

**000012**: Core 3-tier structure (novels, volumes, chapters)
**000013**: Genres, contributors, and translation system
**000014**: Update novels table (remove author_id)

### Apply Migrations

```bash
# Check current version
make migrate-version

# Run all pending migrations
make migrate-up

# Verify
make migrate-version  # Should show: 14
```

### Verify Tables

```bash
make db-shell
```

```sql
-- Check all catalog tables
\dt catalog.*

-- Expected output:
-- novels, volumes, chapters
-- genres, novel_genres
-- authors, artists, translators
-- novel_authors, novel_artists, novel_translators
-- chapter_translations, translation_contributions, translation_history
```

---

## Code Examples

### 1. Create Novel with Authors and Genres

```go
import (
    "context"
    "encoding/json"
    "system/internal/domain"
    "system/internal/pkg/repository"
    "github.com/gofrs/uuid/v5"
)

func createNovelExample(
    novelRepo domain.NovelRepository,
    authorRepo domain.AuthorRepository,
    genreRepo domain.GenreRepository,
) error {
    ctx := context.Background()

    // 1. Create author
    author := &domain.Author{
        ID:   uuid.Must(uuid.NewV4()),
        Name: "Thiên Tằm Thổ Đậu",
        Slug: "thien-tam-tho-dau",
        Biography: json.RawMessage(`{
            "language": "vi",
            "blocks": [
                {"type": "paragraph", "content": "Tác giả nổi tiếng..."}
            ]
        }`),
        IsVerified: true,
    }
    if err := authorRepo.Create(ctx, author); err != nil {
        return err
    }

    // 2. Create novel
    novel := &domain.Novel{
        ID:    uuid.Must(uuid.NewV4()),
        Title: "Đấu Phá Thương Khung",
        Slug:  "dau-pha-thuong-khung",
        Synopsis: json.RawMessage(`{
            "language": "vi",
            "blocks": [
                {"type": "paragraph", "content": "Trong thế giới tu tiên..."}
            ]
        }`),
        OriginalLanguage: strPtr("zh"),
        OriginalTitle:    strPtr("斗破苍穹"),
        Status:           domain.NovelStatusOngoing,
    }
    if err := novelRepo.Create(ctx, novel); err != nil {
        return err
    }

    // 3. Link author to novel
    if err := authorRepo.AddNovelAuthor(ctx, novel.ID, author.ID, "original_author", 0); err != nil {
        return err
    }

    // 4. Add genres
    fantasyID := getGenreBySlug("fantasy")
    xuanhuanID := getGenreBySlug("xuanhuan")
    if err := genreRepo.UpdateNovelGenres(ctx, novel.ID, []uuid.UUID{fantasyID, xuanhuanID}); err != nil {
        return err
    }

    return nil
}

func strPtr(s string) *string {
    return &s
}
```

### 2. Create Volume and Chapters

```go
func createVolumeWithChapters(
    volumeRepo domain.VolumeRepository,
    chapterRepo domain.ChapterRepository,
    novelID uuid.UUID,
) error {
    ctx := context.Background()

    // Create volume
    volume := &domain.Volume{
        ID:           uuid.Must(uuid.NewV4()),
        NovelID:      novelID,
        VolumeNumber: 1,
        Title:        "Tập 1: Khởi đầu",
        Slug:         "tap-1-khoi-dau",
        DisplayOrder: 1,
        IsPublished:  true,
    }
    if err := volumeRepo.Create(ctx, volume); err != nil {
        return err
    }

    // Create chapters
    for i := 1; i <= 10; i++ {
        content := map[string]interface{}{
            "version": "1.0",
            "blocks": []map[string]interface{}{
                {
                    "type":    "paragraph",
                    "content": fmt.Sprintf("Nội dung chương %d...", i),
                },
            },
        }
        contentJSON, _ := json.Marshal(content)

        chapter := &domain.Chapter{
            ID:            uuid.Must(uuid.NewV4()),
            NovelID:       novelID,
            VolumeID:      &volume.ID,
            ChapterNumber: i,
            Title:         fmt.Sprintf("Chương %d", i),
            Slug:          fmt.Sprintf("chuong-%d", i),
            Content:       contentJSON,
            WordCount:     1500,
            IsFree:        i <= 3, // First 3 chapters free
            Status:        domain.ChapterStatusPublished,
            DisplayOrder:  i,
        }

        if err := chapterRepo.Create(ctx, chapter); err != nil {
            return err
        }
    }

    return nil
}
```

### 3. Translation Workflow

```go
func translationWorkflow(
    translationRepo domain.ChapterTranslationRepository,
    contributionRepo domain.TranslationContributionRepository,
    chapterID, translatorID, userID, moderatorID uuid.UUID,
) error {
    ctx := context.Background()

    // Step 1: Translator creates official translation
    contentJSON := json.RawMessage(`{
        "version": "1.0",
        "blocks": [
            {"type": "paragraph", "content": "Bản dịch tiếng Việt..."}
        ]
    }`)

    translation := &domain.ChapterTranslation{
        ID:           uuid.Must(uuid.NewV4()),
        ChapterID:    chapterID,
        Language:     "vi",
        Title:        "Chương 1: Khởi đầu",
        Content:      contentJSON,
        TranslatorID: &translatorID,
        Status:       domain.TranslationStatusDraft,
        WordCount:    1500,
    }
    if err := translationRepo.Create(ctx, translation); err != nil {
        return err
    }

    // Step 2: Publish translation
    if err := translationRepo.Publish(ctx, translation.ID); err != nil {
        return err
    }

    // Step 3: User submits improvement contribution
    improvedJSON := json.RawMessage(`{
        "version": "1.0",
        "blocks": [
            {"type": "paragraph", "content": "Bản dịch cải thiện..."}
        ]
    }`)

    contribution := &domain.TranslationContribution{
        ID:               uuid.Must(uuid.NewV4()),
        ChapterID:        chapterID,
        ContributorID:    userID,
        Language:         "vi",
        ContributionType: domain.ContributionTypeImprovement,
        Content:          improvedJSON,
        ContributorNotes: strPtr("Cải thiện ngữ pháp và độ chính xác"),
        Status:           domain.TranslationStatusPendingReview,
        WordCount:        1500,
    }
    if err := contributionRepo.Create(ctx, contribution); err != nil {
        return err
    }

    // Step 4: Moderator reviews and approves
    if err := contributionRepo.Approve(ctx, contribution.ID, moderatorID, "Bản dịch tốt!", 100); err != nil {
        return err
    }

    // Step 5: Update official translation with approved content
    translation.Content = contribution.Content
    translation.Version++ // Auto-incremented, history auto-created by trigger
    if err := translationRepo.Update(ctx, translation); err != nil {
        return err
    }

    return nil
}
```

### 4. Query with Filters

```go
func queryNovelsExample(novelRepo domain.NovelRepository) ([]*domain.Novel, error) {
    ctx := context.Background()

    status := domain.NovelStatusOngoing
    searchQuery := "tu tiên"

    filter := domain.NovelFilter{
        Status:           &status,
        GenreIDs:         []uuid.UUID{fantasyGenreID, xuanhuanGenreID},
        OriginalLanguage: strPtr("zh"),
        SearchQuery:      &searchQuery,
        SortBy:           "rating",
        SortOrder:        "desc",
        Limit:            20,
        Offset:           0,
    }

    novels, total, err := novelRepo.List(ctx, filter)
    if err != nil {
        return nil, err
    }

    log.Printf("Found %d novels matching filter", total)
    return novels, nil
}
```

---

## Setup Guide

### Prerequisites

- Docker và Docker Compose đang chạy
- PostgreSQL 13+ (via Docker)
- Go 1.21+
- golang-migrate tool

### Step-by-Step Setup

#### 1. Check Environment

```bash
# Verify Docker is running
docker ps

# Check database connection
make db-shell
```

#### 2. Run Migrations

```bash
# Check current migration version
make migrate-version

# Run all pending migrations
make migrate-up

# Verify migrations completed
make migrate-version  # Should show: 14
```

#### 3. Verify Database

```sql
-- Connect to database
make db-shell

-- List all catalog tables
\dt catalog.*

-- Check table structure
\d catalog.novels
\d catalog.genres
\d catalog.authors
\d catalog.chapter_translations

-- Test JSONB functionality
SELECT title, synopsis->>'language' as lang
FROM catalog.novels
LIMIT 5;
```

#### 4. Seed Sample Data (Optional)

```sql
-- Create sample genres
INSERT INTO catalog.genres (id, name, slug, display_order, is_active) VALUES
(gen_random_uuid(), 'Fantasy', 'fantasy', 1, true),
(gen_random_uuid(), 'Xuanhuan', 'xuanhuan', 2, true),
(gen_random_uuid(), 'Romance', 'romance', 3, true),
(gen_random_uuid(), 'Action', 'action', 4, true);

-- Create sample author
INSERT INTO catalog.authors (id, name, slug, is_verified) VALUES
(gen_random_uuid(), 'Thiên Tằm Thổ Đậu', 'thien-tam-tho-dau', true);

-- Create sample translator
INSERT INTO catalog.translators (id, name, slug, languages, is_verified) VALUES
(gen_random_uuid(), 'Hội Dịch Thuật', 'hoi-dich-thuat', '["vi", "en"]'::jsonb, true);
```

#### 5. Initialize Repositories in Code

```go
package main

import (
    "system/internal/pkg/repository"
    "system/internal/platform/database"
)

func setupRepositories() {
    // Get database pool
    pool, err := database.GetPostgresPool(ctx, config.Database)
    if err != nil {
        log.Fatal(err)
    }

    // Initialize all repositories
    novelRepo := repository.NewNovelRepository(pool)
    volumeRepo := repository.NewVolumeRepository(pool)
    chapterRepo := repository.NewChapterRepository(pool)
    genreRepo := repository.NewGenreRepository(pool)
    authorRepo := repository.NewAuthorRepository(pool)
    translatorRepo := repository.NewTranslatorRepository(pool)
    translationRepo := repository.NewChapterTranslationRepository(pool)
    contributionRepo := repository.NewTranslationContributionRepository(pool)

    // Use repositories...
}
```

---

## API Endpoints

### Novels

```
GET    /api/v1/novels              - List novels with filters
GET    /api/v1/novels/:id          - Get novel details
POST   /api/v1/novels              - Create novel
PUT    /api/v1/novels/:id          - Update novel
DELETE /api/v1/novels/:id          - Delete novel (soft delete)
GET    /api/v1/novels/:id/volumes  - Get novel's volumes
GET    /api/v1/novels/:id/chapters - Get novel's chapters
```

### Genres

```
GET    /api/v1/genres              - List all genres
GET    /api/v1/genres/:id          - Get genre details
GET    /api/v1/genres/:id/novels   - Get novels in genre
POST   /api/v1/genres              - Create genre (admin)
PUT    /api/v1/genres/:id          - Update genre (admin)
```

### Authors

```
GET    /api/v1/authors             - List authors
GET    /api/v1/authors/:id         - Get author profile
GET    /api/v1/authors/:id/novels  - Get author's novels
POST   /api/v1/authors             - Create author profile
PUT    /api/v1/authors/:id         - Update author profile
```

### Translators

```
GET    /api/v1/translators                    - List translators
GET    /api/v1/translators/:id                - Get translator profile
GET    /api/v1/translators/:id/translations   - Get translator's work
POST   /api/v1/translators                    - Create translator profile
```

### Translations

```
GET    /api/v1/chapters/:id/translations      - Get all translations
GET    /api/v1/translations/:id               - Get specific translation
POST   /api/v1/translations                   - Create translation (translator)
PUT    /api/v1/translations/:id               - Update translation
POST   /api/v1/translations/:id/publish       - Publish translation
GET    /api/v1/translations/:id/history       - Get version history
```

### Contributions

```
GET    /api/v1/contributions/pending          - Get pending (moderator)
GET    /api/v1/chapters/:id/contributions     - Get chapter contributions
POST   /api/v1/contributions                  - Submit contribution
PUT    /api/v1/contributions/:id              - Update contribution
POST   /api/v1/contributions/:id/approve      - Approve (moderator)
POST   /api/v1/contributions/:id/reject       - Reject (moderator)
POST   /api/v1/contributions/:id/vote         - Vote (upvote/downvote)
```

---

## Best Practices

### 1. Transaction Usage

```go
// Always use transactions for multi-table operations
func createNovelWithRelationships(pool *pgxpool.Pool, novel *Novel, authorIDs, genreIDs []uuid.UUID) error {
    tx, err := pool.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    // Insert novel
    _, err = tx.Exec(ctx, "INSERT INTO catalog.novels (...) VALUES (...)")
    if err != nil {
        return err
    }

    // Insert authors
    for _, authorID := range authorIDs {
        _, err = tx.Exec(ctx, "INSERT INTO catalog.novel_authors (...) VALUES (...)")
        if err != nil {
            return err
        }
    }

    // Insert genres
    for _, genreID := range genreIDs {
        _, err = tx.Exec(ctx, "INSERT INTO catalog.novel_genres (...) VALUES (...)")
        if err != nil {
            return err
        }
    }

    return tx.Commit(ctx)
}
```

### 2. JSONB Querying

```sql
-- Search in JSONB fields
SELECT * FROM catalog.novels
WHERE synopsis @> '{"language": "vi"}'::jsonb;

-- Extract JSONB values
SELECT title, synopsis->>'language' as language
FROM catalog.novels;

-- Filter by array in JSONB
SELECT * FROM catalog.translators
WHERE languages @> '["vi"]'::jsonb;
```

### 3. Pagination

```go
// Always use LIMIT and OFFSET
filter := domain.NovelFilter{
    Limit:  20,  // Items per page
    Offset: 0,   // (page - 1) * limit
}

novels, total, err := novelRepo.List(ctx, filter)

// Calculate total pages
totalPages := (total + int64(filter.Limit) - 1) / int64(filter.Limit)
```

### 4. Translation Workflow

**Best Flow:**
```
1. Translator creates draft
2. Translator completes and submits for review
3. Editor/Moderator reviews
4. Translator publishes
5. Community can submit improvements
6. Moderator reviews improvements
7. Merge approved improvements to official
```

### 5. Credit Points System

**Suggested Points:**
- New translation: 100-500 (based on word count)
- Improvement accepted: 50-200
- Proofreading accepted: 20-100
- Correction accepted: 10-50

### 6. Caching Strategy

```go
// Cache frequently accessed data
- Novel details: 5 minutes
- Chapter list: 10 minutes
- Genre list: 1 hour
- Author profiles: 30 minutes
- Translation content: 15 minutes
```

---

## Troubleshooting

### Issue: Migration fails with foreign key error

**Problem:** Existing novels have `author_id` that doesn't exist in authors table.

**Solution:**
```sql
-- Option 1: Create authors from existing user IDs
INSERT INTO catalog.authors (id, user_id, name, slug)
SELECT DISTINCT
    gen_random_uuid(),
    n.author_id,
    u.full_name,
    lower(regexp_replace(u.full_name, '[^a-zA-Z0-9]+', '-', 'g'))
FROM catalog.novels n
JOIN identify.users u ON n.author_id = u.id
WHERE n.author_id IS NOT NULL;

-- Then link to novels
INSERT INTO catalog.novel_authors (novel_id, author_id, role, display_order)
SELECT n.id, a.id, 'original_author', 0
FROM catalog.novels n
JOIN catalog.authors a ON n.author_id = a.user_id;

-- Option 2: Set author_id to NULL temporarily
UPDATE catalog.novels SET author_id = NULL WHERE author_id IS NOT NULL;
```

### Issue: JSONB query not working

**Problem:** Can't find records with JSONB contains query.

**Solution:**
```sql
-- Check if GIN index exists
SELECT * FROM pg_indexes WHERE tablename = 'novels' AND indexname LIKE '%jsonb%';

-- Create GIN index if missing
CREATE INDEX idx_novels_synopsis ON catalog.novels USING GIN(synopsis);
CREATE INDEX idx_novels_metadata ON catalog.novels USING GIN(metadata);

-- Test query
SELECT * FROM catalog.novels
WHERE metadata @> '{"tags": ["fantasy"]}'::jsonb;
```

### Issue: Duplicate translations

**Problem:** Multiple translations for same (chapter, language).

**Solution:**
```sql
-- Find duplicates
SELECT chapter_id, language, COUNT(*)
FROM catalog.chapter_translations
WHERE deleted_at IS NULL
GROUP BY chapter_id, language
HAVING COUNT(*) > 1;

-- Keep only latest version
DELETE FROM catalog.chapter_translations t1
WHERE id NOT IN (
    SELECT id
    FROM catalog.chapter_translations t2
    WHERE t1.chapter_id = t2.chapter_id
      AND t1.language = t2.language
      AND t2.deleted_at IS NULL
    ORDER BY version DESC, created_at DESC
    LIMIT 1
);
```

### Issue: Statistics not auto-updating

**Problem:** Triggers not firing.

**Solution:**
```sql
-- Check if triggers exist
SELECT * FROM pg_trigger WHERE tgname LIKE '%novel%';

-- Re-create triggers
DROP TRIGGER IF EXISTS trg_volumes_update_novel_stats ON catalog.volumes;
CREATE TRIGGER trg_volumes_update_novel_stats
    AFTER INSERT OR UPDATE OR DELETE ON catalog.volumes
    FOR EACH ROW
    EXECUTE FUNCTION catalog.update_novel_volume_stats();

-- Manually recalculate statistics
UPDATE catalog.novels n
SET
    total_volumes = (
        SELECT COUNT(*) FROM catalog.volumes v
        WHERE v.novel_id = n.id AND v.deleted_at IS NULL
    ),
    total_chapters = (
        SELECT COUNT(*) FROM catalog.chapters c
        WHERE c.novel_id = n.id AND c.deleted_at IS NULL
    );
```

### Issue: Translation version not incrementing

**Problem:** Version stays at 1 after updates.

**Solution:**
```sql
-- Check if trigger exists
SELECT * FROM pg_trigger
WHERE tgname = 'trg_chapter_translations_history';

-- The trigger auto-increments version
-- If not working, check trigger function
\df catalog.create_translation_history

-- Manual increment if needed
UPDATE catalog.chapter_translations
SET version = version + 1
WHERE id = 'translation-id';
```

### Issue: Performance issues with large datasets

**Solution:**
```sql
-- Add missing indexes
CREATE INDEX IF NOT EXISTS idx_chapters_novel_id ON catalog.chapters(novel_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_chapters_published ON catalog.chapters(published_at DESC) WHERE status = 'published';

-- Analyze tables
ANALYZE catalog.novels;
ANALYZE catalog.chapters;
ANALYZE catalog.chapter_translations;

-- Check query performance
EXPLAIN ANALYZE
SELECT * FROM catalog.novels
WHERE status = 'ongoing'
LIMIT 20;
```

---

## Enum Types Reference

```sql
-- Novel status
CREATE TYPE catalog.novel_status AS ENUM (
    'draft',      -- Bản nháp
    'ongoing',    -- Đang tiến hành
    'completed',  -- Đã hoàn thành
    'hiatus',     -- Tạm ngừng
    'dropped'     -- Đã drop
);

-- Chapter status
CREATE TYPE catalog.chapter_status AS ENUM (
    'draft',      -- Bản nháp
    'published',  -- Đã xuất bản
    'scheduled'   -- Đã đặt lịch
);

-- Translation status
CREATE TYPE catalog.translation_status AS ENUM (
    'draft',          -- Bản nháp
    'pending_review', -- Chờ review
    'approved',       -- Đã duyệt
    'rejected',       -- Bị từ chối
    'published'       -- Đã xuất bản
);

-- Contribution type
CREATE TYPE catalog.contribution_type AS ENUM (
    'new_translation', -- Bản dịch mới
    'improvement',     -- Cải thiện
    'proofreading',    -- Đọc và sửa lỗi
    'correction'       -- Sửa lỗi cụ thể
);
```

---

## Resources

- [PostgreSQL JSONB Documentation](https://www.postgresql.org/docs/current/datatype-json.html)
- [PostgreSQL GIN Indexes](https://www.postgresql.org/docs/current/gin.html)
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [pgx - PostgreSQL Driver](https://pkg.go.dev/github.com/jackc/pgx/v5)
- [ISO 639-1 Language Codes](https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes)

---

**Status:** ✅ Complete database system for novel platform
**Migrations:** 000012, 000013, 000014
**Schema:** catalog
**Date:** 2025-11-18
**Version:** 2.0
