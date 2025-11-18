# Novel Database Design Documentation

## Tổng quan

Hệ thống database cho novel được thiết kế với cấu trúc 3 cấp độ phân cấp:
- **Novel** (Tiểu thuyết) - Cấp cao nhất
- **Volume** (Tập) - Cấp giữa
- **Chapter** (Chương) - Cấp thấp nhất

Tất cả các bảng được đặt trong schema `catalog` để tách biệt với các schema khác trong hệ thống.

## Cấu trúc Database

### 1. Bảng `catalog.novels` (Cấp cao nhất)

Lưu trữ thông tin cơ bản về tiểu thuyết.

**Các trường chính:**
- `id` (UUID) - Primary key
- `title` (VARCHAR) - Tên tiểu thuyết
- `slug` (VARCHAR) - SEO-friendly URL (unique)
- `author_id` (UUID) - Foreign key đến `identify.users`
- `synopsis` (JSONB) - Tóm tắt nội dung dạng JSON
- `metadata` (JSONB) - Thông tin bổ sung (tags, categories, language)
- `status` (ENUM) - Trạng thái: draft, ongoing, completed, hiatus, dropped
- Thống kê: `total_volumes`, `total_chapters`, `total_words`, `view_count`, `rating_average`, v.v.

**Ví dụ JSONB cho `synopsis`:**
```json
{
  "language": "vi",
  "blocks": [
    {
      "type": "paragraph",
      "content": "Trong thế giới tu tiên đầy nguy hiểm..."
    },
    {
      "type": "paragraph",
      "content": "Nhân vật chính phải đối mặt với..."
    }
  ]
}
```

**Ví dụ JSONB cho `metadata`:**
```json
{
  "tags": ["fantasy", "action", "cultivation"],
  "categories": ["xuanhuan", "donghuang"],
  "language": "vi",
  "original_language": "zh",
  "original_title": "斗破苍穹",
  "translator": "Translator Team",
  "status_note": "Đang dịch chapter mới nhất"
}
```

### 2. Bảng `catalog.volumes` (Cấp giữa)

Tổ chức các chapter thành các tập.

**Các trường chính:**
- `id` (UUID) - Primary key
- `novel_id` (UUID) - Foreign key đến `novels`
- `volume_number` (INTEGER) - Số thứ tự tập (unique với novel_id)
- `title` (VARCHAR) - Tên tập
- `slug` (VARCHAR) - SEO-friendly URL
- `display_order` (INTEGER) - Thứ tự hiển thị tùy chỉnh
- `is_published` (BOOLEAN) - Trạng thái công khai
- Thống kê: `chapter_count`, `word_count`

**Lưu ý:** Chapter có thể tồn tại mà không thuộc volume nào (volume_id = NULL).

### 3. Bảng `catalog.chapters` (Cấp thấp nhất)

Lưu trữ nội dung chapter.

**Các trường chính:**
- `id` (UUID) - Primary key
- `novel_id` (UUID) - Foreign key đến `novels`
- `volume_id` (UUID, nullable) - Foreign key đến `volumes`
- `chapter_number` (INTEGER) - Số thứ tự chapter (unique với novel_id)
- `title` (VARCHAR) - Tên chapter
- `content` (JSONB) - Nội dung chapter dạng JSON
- `author_notes` (JSONB) - Ghi chú của tác giả dạng JSON
- `status` (ENUM) - Trạng thái: draft, published, scheduled
- `is_free` (BOOLEAN) - Chapter miễn phí hay trả phí
- `price`, `currency` - Giá tiền nếu trả phí
- Thống kê: `view_count`, `like_count`, `comment_count`, `word_count`, `character_count`

**Ví dụ JSONB cho `content`:**
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
      "type": "paragraph",
      "content": "Tiêu Viêm ngồi trên tảng đá, ngước nhìn bầu trời..."
    },
    {
      "type": "dialogue",
      "speaker": "Tiêu Viêm",
      "content": "Ta sẽ trở thành người mạnh nhất!"
    },
    {
      "type": "image",
      "url": "https://example.com/image.jpg",
      "caption": "Minh họa cảnh núi Thanh Vân"
    }
  ]
}
```

**Ví dụ JSONB cho `author_notes`:**
```json
{
  "blocks": [
    {
      "type": "paragraph",
      "content": "Cảm ơn các bạn đã ủng hộ!"
    },
    {
      "type": "paragraph",
      "content": "Chapter tiếp theo sẽ ra vào thứ 7 tuần sau."
    }
  ]
}
```

## Enum Types

### `catalog.novel_status`
- `draft` - Bản nháp, chưa công khai
- `ongoing` - Đang tiến hành
- `completed` - Đã hoàn thành
- `hiatus` - Tạm ngừng
- `dropped` - Đã drop

### `catalog.chapter_status`
- `draft` - Bản nháp
- `published` - Đã xuất bản
- `scheduled` - Đã đặt lịch xuất bản

## Triggers và Auto-update

Hệ thống có các trigger tự động cập nhật thống kê:

1. **Khi volume thay đổi** → Cập nhật `total_volumes` trong `novels`
2. **Khi chapter thay đổi** → Cập nhật:
   - `total_chapters`, `total_words`, `last_chapter_at` trong `novels`
   - `chapter_count`, `word_count` trong `volumes` (nếu có)
3. **Khi update record** → Tự động cập nhật `updated_at`

## Indexes

Hệ thống có các index được tối ưu cho:
- Truy vấn theo author_id
- Truy vấn theo status
- Sắp xếp theo rating, views, created_at
- Full-text search trong synopsis và content (GIN index)
- Filter theo metadata (GIN index)

## Ví dụ sử dụng với Go

### 1. Tạo Novel mới

```go
package main

import (
	"context"
	"encoding/json"
	"system/internal/domain"
	"system/internal/pkg/repository"

	"github.com/gofrs/uuid/v5"
)

func createNovel(novelRepo domain.NovelRepository, authorID uuid.UUID) error {
	ctx := context.Background()

	// Tạo synopsis JSON
	synopsis := map[string]interface{}{
		"language": "vi",
		"blocks": []map[string]interface{}{
			{
				"type":    "paragraph",
				"content": "Trong thế giới tu tiên...",
			},
		},
	}
	synopsisJSON, _ := json.Marshal(synopsis)

	// Tạo metadata JSON
	metadata := map[string]interface{}{
		"tags":       []string{"fantasy", "action"},
		"categories": []string{"xuanhuan"},
		"language":   "vi",
	}
	metadataJSON, _ := json.Marshal(metadata)

	novel := &domain.Novel{
		ID:       uuid.Must(uuid.NewV4()),
		Title:    "Đấu Phá Thương Khung",
		Slug:     "dau-pha-thuong-khung",
		AuthorID: authorID,
		Synopsis: synopsisJSON,
		Metadata: metadataJSON,
		Status:   domain.NovelStatusDraft,
	}

	return novelRepo.Create(ctx, novel)
}
```

### 2. Tạo Volume và Chapter

```go
func createVolumeAndChapter(
	volumeRepo domain.VolumeRepository,
	chapterRepo domain.ChapterRepository,
	novelID uuid.UUID,
) error {
	ctx := context.Background()

	// Tạo volume
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

	// Tạo chapter content JSON
	content := map[string]interface{}{
		"version": "1.0",
		"blocks": []map[string]interface{}{
			{
				"type":    "paragraph",
				"content": "Nội dung chapter...",
			},
		},
	}
	contentJSON, _ := json.Marshal(content)

	// Tạo chapter
	chapter := &domain.Chapter{
		ID:            uuid.Must(uuid.NewV4()),
		NovelID:       novelID,
		VolumeID:      &volume.ID,
		ChapterNumber: 1,
		Title:         "Chương 1: Bắt đầu",
		Slug:          "chuong-1-bat-dau",
		Content:       contentJSON,
		WordCount:     1500,
		IsFree:        true,
		Status:        domain.ChapterStatusPublished,
		DisplayOrder:  1,
	}

	return chapterRepo.Create(ctx, chapter)
}
```

### 3. Query Novel với Filter

```go
func listNovels(novelRepo domain.NovelRepository) ([]*domain.Novel, int64, error) {
	ctx := context.Background()

	status := domain.NovelStatusOngoing
	searchQuery := "tu tiên"

	filter := domain.NovelFilter{
		Status:      &status,
		Tags:        []string{"fantasy"},
		Categories:  []string{"xuanhuan"},
		SearchQuery: &searchQuery,
		SortBy:      "rating",
		SortOrder:   "desc",
		Limit:       20,
		Offset:      0,
	}

	return novelRepo.List(ctx, filter)
}
```

### 4. Đọc Chapter với Parsing JSON Content

```go
func readChapter(chapterRepo domain.ChapterRepository, chapterID uuid.UUID) error {
	ctx := context.Background()

	chapter, err := chapterRepo.GetByID(ctx, chapterID)
	if err != nil {
		return err
	}

	// Parse content JSON
	var content struct {
		Version string `json:"version"`
		Blocks  []struct {
			Type    string `json:"type"`
			Content string `json:"content"`
		} `json:"blocks"`
	}

	if err := json.Unmarshal(chapter.Content, &content); err != nil {
		return err
	}

	// Process content blocks
	for _, block := range content.Blocks {
		switch block.Type {
		case "paragraph":
			fmt.Printf("Paragraph: %s\n", block.Content)
		case "dialogue":
			fmt.Printf("Dialogue: %s\n", block.Content)
		}
	}

	// Increment view count
	return chapterRepo.IncrementViewCount(ctx, chapterID)
}
```

## Migration

Để apply migration:

```bash
# Check migration status
make migrate-status

# Run migration up
make migrate-up

# Rollback migration (nếu cần)
make migrate-down
```

## Lưu ý quan trọng

1. **JSONB Performance**: Các trường JSONB có GIN index để tối ưu truy vấn. Sử dụng `@>` operator cho JSONB contains queries.

2. **Soft Delete**: Tất cả các bảng sử dụng soft delete với `deleted_at` field. Luôn check `deleted_at IS NULL` trong query.

3. **Unique Constraints**:
   - Novel: slug phải unique
   - Volume: (novel_id, volume_number) phải unique
   - Chapter: (novel_id, chapter_number) phải unique

4. **Auto-update Statistics**: Không cần manually update statistics (total_volumes, total_chapters, etc.) vì được handle bởi triggers.

5. **Transaction**: Khi tạo/update nhiều records liên quan, nên sử dụng transaction để đảm bảo data consistency.

## Performance Considerations

1. **Pagination**: Luôn sử dụng LIMIT và OFFSET cho danh sách lớn
2. **Select Fields**: Khi list chapters, có thể dùng ChapterSummary thay vì full Chapter để giảm data size
3. **Caching**: Nên cache các novel metadata và chapter list
4. **Full-text Search**: Sử dụng GIN index cho search performance tốt hơn

## Future Enhancements

Có thể mở rộng:
1. Table `novel_ratings` - Lưu rating của từng user
2. Table `novel_comments` - Comment và discussion
3. Table `reading_progress` - Track vị trí đọc của user
4. Table `novel_bookmarks` - Bookmark và favorites
5. Table `novel_tags` và `novel_categories` - Normalize tags và categories
