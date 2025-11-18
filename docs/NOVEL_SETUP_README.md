# Novel Database Schema - Setup Guide

## Tổng quan thay đổi

Đã thiết kế và implement hoàn chỉnh hệ thống database cho novel với cấu trúc 3 cấp: **Novel → Volume → Chapter**.

## Các file đã tạo

### 1. Migration Files
- `migrations/000012_create_novel_tables.up.sql` - Migration script để tạo tables
- `migrations/000012_create_novel_tables.down.sql` - Rollback script

### 2. Domain Models (Entity Definitions)
- `internal/domain/novel.go` - Novel entity và repository interface
- `internal/domain/volume.go` - Volume entity và repository interface
- `internal/domain/chapter.go` - Chapter entity và repository interface

### 3. Repository Implementations (Data Access Layer)
- `internal/pkg/repository/novel_repo.go` - Novel repository implementation
- `internal/pkg/repository/volume_repo.go` - Volume repository implementation
- `internal/pkg/repository/chapter_repo.go` - Chapter repository implementation

### 4. Documentation
- `docs/novel_database_design.md` - Chi tiết thiết kế database và usage examples
- `docs/NOVEL_SETUP_README.md` - File này

## Cấu trúc Database

### Schema: `catalog`
Tất cả các bảng novel được đặt trong schema `catalog` để tách biệt với các schema khác.

### Tables

#### 1. `catalog.novels` (Top Level)
```sql
- id (UUID, PK)
- title, slug, author_id
- synopsis (JSONB) ⭐ Rich content format
- metadata (JSONB) ⭐ Tags, categories, language
- status: draft, ongoing, completed, hiatus, dropped
- Statistics: total_volumes, total_chapters, view_count, rating_average
- Timestamps: created_at, updated_at, deleted_at (soft delete)
```

#### 2. `catalog.volumes` (Middle Level)
```sql
- id (UUID, PK)
- novel_id (FK → novels)
- volume_number (unique per novel)
- title, slug, description
- display_order, is_published
- Statistics: chapter_count, word_count
- Timestamps: created_at, updated_at, deleted_at
```

#### 3. `catalog.chapters` (Bottom Level)
```sql
- id (UUID, PK)
- novel_id (FK → novels)
- volume_id (FK → volumes, nullable)
- chapter_number (unique per novel)
- title, slug
- content (JSONB) ⭐ Rich content format
- author_notes (JSONB) ⭐ Rich format
- status: draft, published, scheduled
- Access control: is_free, price, currency
- Statistics: view_count, like_count, comment_count, word_count
- Timestamps: published_at, scheduled_at, created_at, updated_at, deleted_at
```

### Đặc điểm nổi bật

✅ **JSONB Fields** - Synopsis và Content lưu dạng JSON linh hoạt
✅ **Auto-update Statistics** - Triggers tự động cập nhật thống kê
✅ **Soft Delete** - Tất cả tables support soft delete
✅ **Optimized Indexes** - GIN indexes cho JSONB, indexes cho search và sort
✅ **Multi-language Support** - Metadata field hỗ trợ đa ngôn ngữ
✅ **Scheduled Publishing** - Chapter có thể đặt lịch xuất bản
✅ **Flexible Structure** - Chapter có thể tồn tại độc lập không cần volume

## Cách áp dụng

### Bước 1: Kiểm tra môi trường

```bash
# Check Docker đang chạy
docker ps

# Check database connection
make db-shell
```

### Bước 2: Run migration

```bash
# Check migration status hiện tại
make migrate-status

# Run migration để tạo tables
make migrate-up

# Verify migration thành công
make migrate-version
# Output: 12 (version mới nhất)
```

### Bước 3: Verify tables đã tạo

```bash
# Kết nối database
make db-shell

# Trong psql shell:
\dt catalog.*

# Expected output:
#              List of relations
#  Schema  |    Name    | Type  |   Owner
# ---------+------------+-------+------------
#  catalog | chapters   | table | system_dev
#  catalog | novels     | table | system_dev
#  catalog | volumes    | table | system_dev
```

### Bước 4: Test JSONB fields

```sql
-- Test insert novel với JSONB synopsis
INSERT INTO catalog.novels (
    id, title, slug, author_id, synopsis, metadata, status
) VALUES (
    gen_random_uuid(),
    'Test Novel',
    'test-novel',
    (SELECT id FROM identify.users LIMIT 1),
    '{"language": "vi", "blocks": [{"type": "paragraph", "content": "Test content"}]}'::jsonb,
    '{"tags": ["fantasy"], "language": "vi"}'::jsonb,
    'draft'
);

-- Test query với JSONB
SELECT title, synopsis->>'language' as language
FROM catalog.novels
WHERE metadata @> '{"tags": ["fantasy"]}'::jsonb;
```

## Example Usage trong Code

### Initialize Repositories

```go
package main

import (
    "system/internal/pkg/repository"
    "system/internal/platform/database"
)

func main() {
    // Get database pool
    pool, err := database.GetPostgresPool(ctx, config.Database)
    if err != nil {
        log.Fatal(err)
    }

    // Initialize repositories
    novelRepo := repository.NewNovelRepository(pool)
    volumeRepo := repository.NewVolumeRepository(pool)
    chapterRepo := repository.NewChapterRepository(pool)

    // Use repositories...
}
```

### Create Novel with JSONB

```go
import (
    "encoding/json"
    "system/internal/domain"
    "github.com/gofrs/uuid/v5"
)

func createNovel() error {
    // Create synopsis JSON
    synopsis := map[string]interface{}{
        "language": "vi",
        "blocks": []map[string]interface{}{
            {
                "type": "paragraph",
                "content": "Một câu chuyện về tu tiên...",
            },
        },
    }
    synopsisJSON, _ := json.Marshal(synopsis)

    // Create metadata JSON
    metadata := map[string]interface{}{
        "tags": []string{"fantasy", "action", "cultivation"},
        "categories": []string{"xuanhuan"},
        "language": "vi",
    }
    metadataJSON, _ := json.Marshal(metadata)

    novel := &domain.Novel{
        ID:       uuid.Must(uuid.NewV4()),
        Title:    "Đấu Phá Thương Khung",
        Slug:     "dau-pha-thuong-khung",
        AuthorID: authorID,
        Synopsis: synopsisJSON,
        Metadata: metadataJSON,
        Status:   domain.NovelStatusOngoing,
    }

    return novelRepo.Create(ctx, novel)
}
```

### Query with Filters

```go
func listPopularNovels() ([]*domain.Novel, error) {
    status := domain.NovelStatusOngoing

    filter := domain.NovelFilter{
        Status:    &status,
        Tags:      []string{"fantasy"},
        SortBy:    "rating",
        SortOrder: "desc",
        Limit:     20,
        Offset:    0,
    }

    novels, total, err := novelRepo.List(ctx, filter)
    return novels, err
}
```

## Testing

### Unit Tests (TODO - Có thể implement sau)

```bash
# Create test file
touch internal/pkg/repository/novel_repo_test.go

# Run tests
make test
```

### Manual Testing

```bash
# 1. Start services
make docker-up

# 2. Run migrations
make migrate-up

# 3. Test với curl hoặc Postman
# (Sau khi implement HTTP handlers)
```

## Rollback

Nếu cần rollback migration:

```bash
# Rollback về version trước
make migrate-down

# Check version
make migrate-version
# Output: 11
```

## Next Steps (Tuỳ chọn)

### 1. Service Layer (Business Logic)
Tạo các file:
- `internal/pkg/service/novel_service.go`
- `internal/pkg/service/volume_service.go`
- `internal/pkg/service/chapter_service.go`

### 2. HTTP Handlers (API Endpoints)
Tạo các file:
- `internal/app/handler/v1/novel/handler.go`
- `internal/app/handler/v1/novel/router.go`
- `internal/app/handler/v1/novel/request.go`
- `internal/app/handler/v1/novel/response.go`

### 3. API Endpoints Examples

```
GET    /api/v1/novels              - List novels với filter
GET    /api/v1/novels/:id          - Get novel by ID
POST   /api/v1/novels              - Create novel
PUT    /api/v1/novels/:id          - Update novel
DELETE /api/v1/novels/:id          - Delete novel

GET    /api/v1/novels/:id/volumes  - List volumes của novel
POST   /api/v1/novels/:id/volumes  - Create volume

GET    /api/v1/novels/:id/chapters - List chapters của novel
GET    /api/v1/chapters/:id        - Get chapter by ID với content
POST   /api/v1/chapters            - Create chapter
PUT    /api/v1/chapters/:id        - Update chapter
POST   /api/v1/chapters/:id/publish - Publish chapter
```

### 4. Additional Features

- Reading progress tracking
- Bookmarks và favorites
- Comments và ratings system
- Full-text search với PostgreSQL tsvector
- Caching layer với Redis
- CDN cho images

## Troubleshooting

### Issue: Migration failed

```bash
# Check database connection
make db-shell

# Check migration version
make migrate-version

# Force to specific version (nếu cần)
make migrate-force VERSION=11
```

### Issue: Triggers không hoạt động

```sql
-- Check triggers
SELECT * FROM pg_trigger
WHERE tgname LIKE '%novel%' OR tgname LIKE '%chapter%';

-- Re-run migration
make migrate-down
make migrate-up
```

### Issue: JSONB query không hoạt động

```sql
-- Check GIN indexes
SELECT * FROM pg_indexes
WHERE schemaname = 'catalog';

-- Test JSONB contains
SELECT * FROM catalog.novels
WHERE metadata @> '{"tags": ["fantasy"]}'::jsonb;
```

## Resources

- [PostgreSQL JSONB Documentation](https://www.postgresql.org/docs/current/datatype-json.html)
- [golang-migrate Documentation](https://github.com/golang-migrate/migrate)
- [pgx Documentation](https://pkg.go.dev/github.com/jackc/pgx/v5)

## Support

Nếu có vấn đề hoặc câu hỏi:
1. Check file `docs/novel_database_design.md` để xem chi tiết design
2. Check migration logs: `make migrate-version`
3. Check database logs: `make docker-logs`

---

**Status:** ✅ Database schema design hoàn tất
**Version:** 000012
**Date:** 2025-11-18
**Author:** Claude AI
