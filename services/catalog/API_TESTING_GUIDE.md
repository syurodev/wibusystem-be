# Hướng dẫn Test Novel APIs

## APIs đã triển khai

### 1. Tạo tiểu thuyết mới (POST /api/v1/novels)

```bash
curl -X POST http://localhost:8080/api/v1/novels \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "title": "Tên tiểu thuyết mới",
    "cover_image": "https://example.com/cover.jpg",
    "summary": {"content": "Tóm tắt tiểu thuyết"},
    "status": "DRAFT",
    "original_language": "vi",
    "source_url": "https://example.com/source",
    "isbn": "978-3-16-148410-0",
    "age_rating": "PG-13",
    "content_warnings": ["bạo lực", "ngôn ngữ thô tục nhẹ"],
    "mature_content": false,
    "is_public": false,
    "is_featured": false,
    "keywords": "fantasy adventure novel",
    "price_coins": 100,
    "rental_price_coins": 20,
    "rental_duration_days": 30,
    "is_premium": false,
    "genres": ["genre-uuid-1", "genre-uuid-2"],
    "creators": [
      {
        "creator_id": "creator-uuid",
        "role": "AUTHOR"
      }
    ],
    "characters": ["character-uuid-1", "character-uuid-2"]
  }'
```

**Response mẫu:**
```json
{
  "success": true,
  "message": "Tạo tiểu thuyết thành công",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "title": "Tên tiểu thuyết mới",
    "status": "DRAFT",
    "slug": "ten-tieu-thuyet-moi",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "error": null,
  "meta": {}
}
```

### 2. Danh sách tiểu thuyết (GET /api/v1/novels)

```bash
# Basic list
curl "http://localhost:8080/api/v1/novels"

# With filters
curl "http://localhost:8080/api/v1/novels?page=1&limit=20&status=ONGOING&search=fantasy&sort=created_at&order=desc"
```

**Query Parameters:**
- `page`: Số trang (mặc định: 1)
- `limit`: Số lượng mỗi trang (mặc định: 20, tối đa: 100)
- `search`: Tìm kiếm theo tên
- `status`: Lọc theo trạng thái (DRAFT, ONGOING, COMPLETED, HIATUS, CANCELLED)
- `genre`: Lọc theo thể loại (UUID)
- `original_language`: Lọc theo ngôn ngữ gốc
- `is_featured`: Lọc theo featured (true/false)
- `is_completed`: Lọc theo hoàn thành (true/false)
- `sort`: Sắp xếp (created_at, updated_at, published_at, view_count, rating_average)
- `order`: Thứ tự (asc, desc) - mặc định: desc

**Response mẫu:**
```json
{
  "success": true,
  "message": "Lấy danh sách tiểu thuyết thành công",
  "data": {
    "novels": [
      {
        "id": "123e4567-e89b-12d3-a456-426614174000",
        "name": "Tên tiểu thuyết",
        "cover_image": "https://example.com/cover.jpg",
        "view_count": 1500,
        "created_at": "2024-01-01T00:00:00Z",
        "user": {
          "id": "user-uuid",
          "display_name": "Tên tác giả"
        },
        "tenant": {
          "id": "tenant-uuid",
          "name": "Tên tenant"
        },
        "latest_chapter_updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 150,
      "total_pages": 8,
      "has_next": true,
      "has_previous": false
    }
  },
  "error": null,
  "meta": {}
}
```

### 3. Chi tiết tiểu thuyết (GET /api/v1/novels/{id})

```bash
# Basic detail (chỉ thông tin cơ bản + genres, creators, characters)
curl "http://localhost:8080/api/v1/novels/123e4567-e89b-12d3-a456-426614174000"

# With translations (bao gồm bản dịch)
curl "http://localhost:8080/api/v1/novels/123e4567-e89b-12d3-a456-426614174000?include_translations=true" \
  -H "Accept-Language: vi" \
  -H "X-Language: vi"

# With stats (bao gồm thống kê chi tiết)
curl "http://localhost:8080/api/v1/novels/123e4567-e89b-12d3-a456-426614174000?include_stats=true"

# Full detail (bao gồm tất cả: translations + stats)
curl "http://localhost:8080/api/v1/novels/123e4567-e89b-12d3-a456-426614174000?include_translations=true&include_stats=true" \
  -H "Accept-Language: vi" \
  -H "X-Language: vi"
```

**Query Parameters:**
- `include_translations`: Bao gồm dữ liệu dịch (mặc định: false)
- `include_stats`: Bao gồm thống kê chi tiết (mặc định: false)

**Headers:**
- `Accept-Language`: Ngôn ngữ hiển thị (vi, en, ja, etc.)
- `X-Language`: Ngôn ngữ override (ưu tiên hơn Accept-Language)

**📊 Thông tin luôn được load:**
- **Genres**: Tất cả thể loại của novel
- **Creators**: Tất cả người tạo (tác giả, họa sĩ, etc.) với vai trò
- **Characters**: Tất cả nhân vật trong novel

**📊 Thông tin tùy chọn:**
- **Translations** (`include_translations=true`): Các bản dịch của novel
- **Stats** (`include_stats=true`): Thống kê chi tiết bao gồm:
  - **Content stats**: Số volume, chapter, tổng số từ, trung bình từ/chapter
  - **Engagement stats**: Lượt xem, thích, bookmark, comment, rating
  - **Purchase stats**: Số người mua/thuê series, volume, chapter

**Response mẫu (Basic - không include translations/stats):**
```json
{
  "success": true,
  "message": "Lấy thông tin tiểu thuyết thành công",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "Tên tiểu thuyết (theo ngôn ngữ client)",
    "cover_image": "https://example.com/cover.jpg",
    "summary": {"content": "Tóm tắt tiểu thuyết"},
    "status": "ONGOING",
    "published_at": "2024-01-01T00:00:00Z",
    "original_language": "vi",
    "current_language": "vi",
    "source_url": "https://example.com/source",
    "isbn": "978-3-16-148410-0",
    "age_rating": "PG-13",
    "content_warnings": ["bạo lực", "ngôn ngữ thô tục nhẹ"],
    "mature_content": false,
    "is_public": true,
    "is_featured": true,
    "is_completed": false,
    "slug": "ten-tieu-thuyet",
    "keywords": "fantasy adventure novel",
    "price_coins": 100,
    "rental_price_coins": 20,
    "rental_duration_days": 30,
    "is_premium": false,
    "view_count": 1500,
    "rating_average": 4.5,
    "rating_count": 150,
    "chapter_count": 25,
    "volume_count": 3,
    "genres": [
      {
        "id": "genre-uuid-1",
        "name": "Fantasy"
      },
      {
        "id": "genre-uuid-2",
        "name": "Adventure"
      }
    ],
    "creators": [
      {
        "id": "creator-uuid",
        "name": "Tên tác giả",
        "role": "AUTHOR"
      }
    ],
    "characters": [
      {
        "id": "character-uuid-1",
        "name": "Nhân vật chính"
      },
      {
        "id": "character-uuid-2",
        "name": "Nhân vật phụ"
      }
    ],
    "translations": [],
    "stats": {},
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "error": null,
  "meta": {}
}
```

**Response mẫu (Với include_translations=true):**
```json
{
  "success": true,
  "message": "Lấy thông tin tiểu thuyết thành công",
  "data": {
    // ... các field khác giống như trên ...
    "translations": [
      {
        "id": "translation-uuid-1",
        "language_code": "en",
        "title": "English Novel Title",
        "description": "English description",
        "is_primary": true
      },
      {
        "id": "translation-uuid-2",
        "language_code": "ja",
        "title": "Japanese Novel Title",
        "description": null,
        "is_primary": false
      }
    ]
    // ... các field khác ...
  }
}
```

**Response mẫu (Với include_stats=true):**
```json
{
  "success": true,
  "message": "Lấy thông tin tiểu thuyết thành công",
  "data": {
    // ... các field khác giống như trên ...
    "stats": {
      "content": {
        "volume_count": 3,
        "chapter_count": 25,
        "total_word_count": 125000,
        "average_chapter_word_count": 5000
      },
      "engagement": {
        "view_count": 1500,
        "like_count": 89,
        "bookmark_count": 45,
        "comment_count": 23,
        "rating_average": 4.5,
        "rating_count": 150
      },
      "purchases": {
        "series_buyers": 12,
        "volume_buyers": 34,
        "chapter_buyers": 156,
        "series_renters": 5,
        "volume_renters": 28,
        "total_buyers": 202,
        "total_renters": 33
      }
    }
    // ... các field khác ...
  }
}
```

### 4. Cập nhật tiểu thuyết (PUT /api/v1/novels/{id})

```bash
curl -X PUT "http://localhost:8080/api/v1/novels/123e4567-e89b-12d3-a456-426614174000" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "title": "Tên tiểu thuyết mới (cập nhật)",
    "cover_image": "https://example.com/new-cover.jpg",
    "summary": {"content": "Tóm tắt cập nhật"},
    "genres": ["genre-uuid-1", "genre-uuid-2"],
    "is_featured": true,
    "is_completed": true
  }'
```

**Response mẫu:**
```json
{
  "success": true,
  "message": "Cập nhật tiểu thuyết thành công",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "title": "Tên tiểu thuyết mới (cập nhật)"
  },
  "error": null,
  "meta": {}
}
```

### 5. Xóa tiểu thuyết (DELETE /api/v1/novels/{id})

```bash
curl -X DELETE "http://localhost:8080/api/v1/novels/123e4567-e89b-12d3-a456-426614174000" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Response mẫu (thành công):**
```json
{
  "success": true,
  "message": "Xóa tiểu thuyết thành công",
  "data": null,
  "error": null,
  "meta": {}
}
```

**Response mẫu (có người mua):**
```json
{
  "success": false,
  "message": "Cannot delete: users have purchased content",
  "data": null,
  "error": {
    "code": "cannot_delete",
    "description": "cannot delete novel: users have purchased content from this novel"
  },
  "meta": {}
}
```

**⚠️ Đặc điểm quan trọng của API xóa:**

1. **Kiểm tra mua hàng**: API sẽ kiểm tra xem có user nào đã mua:
   - Novel series
   - Bất kỳ volume nào của novel
   - Bất kỳ chapter nào của novel
   - Thuê novel series hoặc volume

2. **Soft delete**: Nếu không có ai mua, sẽ thực hiện soft delete:
   - Set `is_deleted = TRUE`
   - Set `deleted_at` và `deleted_by_user_id`
   - Cũng soft delete tất cả volumes và chapters con

3. **Transaction safety**: Toàn bộ quá trình trong transaction để đảm bảo consistency

## Lưu ý quan trọng

1. **Migration Database**: Cần chạy migration `110_add_name_field_to_content_tables.up.sql` trước khi test
2. **Authentication**: API POST, PUT, DELETE cần token admin
3. **Validation**: Các trường bắt buộc phải có đầy đủ
4. **Error Handling**: Tất cả lỗi đều trả về StandardResponse format

## Status Codes

- `200`: Success
- `201`: Created successfully
- `400`: Bad Request (validation errors)
- `401`: Unauthorized
- `404`: Not Found
- `409`: Conflict (duplicated data)
- `500`: Internal Server Error