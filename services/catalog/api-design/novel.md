# Thiết kế API - Dịch vụ Catalog (Module Tiểu Thuyết)

## Tổng quan

Tài liệu này mô tả chi tiết API cho module **Novel** trong Catalog Service, bao gồm quản lý tiểu thuyết, tập (volume), chương (chapter), bản dịch (translation) và các mối quan hệ liên quan.

## Base URL

```
/api/v1
```

## Chuẩn Response (Response Format)

**TẤT CẢ API responses phải tuân thủ `StandardResponse` format từ `/pkg/common/response/response.go`:**

```go
type StandardResponse struct {
    Success bool                   `json:"success"`
    Message string                 `json:"message"`
    Data    interface{}            `json:"data"`
    Error   *ErrorDetail           `json:"error"`
    Meta    map[string]interface{} `json:"meta"`
}
```

**Quy tắc:**

- `success`: true/false - Trạng thái request
- `message`: string - Thông báo cho user
- `data`: object/array/null - Dữ liệu chính
- `error`: object/null - Chi tiết lỗi (null khi success=true)
- `meta`: object - Metadata (pagination, filters, etc.)

## Xác thực (Authentication)

- **Public endpoints**: Không cần xác thực (chỉ đọc dữ liệu)
- **Protected endpoints**: Cần xác thực bằng token và kiểm tra quyền tương ứng

---

## 1. API Quản lý Tiểu thuyết (Core Novel Management)

### 1.1 Danh sách tiểu thuyết

```http
GET /api/v1/novels
```

**Tham số:**

- `page` (query, tuỳ chọn): Số trang (mặc định: 1)
- `limit` (query, tuỳ chọn): Số lượng mỗi trang (mặc định: 20, tối đa: 100)
- `search` (query, tuỳ chọn): Tìm kiếm theo tên
- `status` (query, tuỳ chọn): Lọc theo trạng thái (DRAFT, ONGOING, COMPLETED, HIATUS, CANCELLED)
- `genre` (query, tuỳ chọn): Lọc theo thể loại (UUID)
- `original_language` (query, tuỳ chọn): Lọc theo ngôn ngữ gốc
- `is_featured` (query, tuỳ chọn): Lọc theo featured (true/false)
- `is_completed` (query, tuỳ chọn): Lọc theo hoàn thành (true/false)
- `sort` (query, tuỳ chọn): Sắp xếp (created_at, updated_at, published_at, view_count, rating_average)
- `order` (query, tuỳ chọn): Thứ tự (asc, desc) - mặc định: desc

**Phản hồi:**

```json
{
  "success": true,
  "message": "Lấy danh sách tiểu thuyết thành công",
  "data": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "title": "Tên tiểu thuyết",
      "cover_image": "https://example.com/cover.jpg",
      "status": "ONGOING",
      "original_language": "vi",
      "is_featured": true,
      "is_completed": false,
      "view_count": 1500,
      "rating_average": 4.5,
      "rating_count": 150,
      "total_chapters": 25,
      "total_volumes": 3,
      "published_at": "2024-01-01T00:00:00Z",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "error": null,
  "meta": {
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 150,
      "total_pages": 8,
      "has_next": true,
      "has_prev": false
    }
  }
}
```

### 1.2 Tạo tiểu thuyết mới

```http
POST /api/v1/novels
```

**Headers:**

- `Authorization: Bearer {token}`

**Body:**

```json
{
  "title": "Tên tiểu thuyết mới",
  "cover_image": "https://example.com/cover.jpg",
  "summary": json,
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
}
```

**Phản hồi:**

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

### 1.3 Lấy chi tiết tiểu thuyết

```http
GET /api/v1/novels/{id}
```

**Tham số:**

- `id` (path, bắt buộc): UUID của tiểu thuyết
- `include_translations` (query, tuỳ chọn): Bao gồm dữ liệu dịch (mặc định: false)
- `include_stats` (query, tuỳ chọn): Bao gồm thống kê (mặc định: false)

**Headers:**

- `Accept-Language` (tuỳ chọn): Ngôn ngữ hiển thị (vi, en, ja, etc.) - mặc định: original_language
- `X-Language` (tuỳ chọn): Ngôn ngữ hiển thị override (ưu tiên hơn Accept-Language)

**Phản hồi:**

```json
{
  "success": true,
  "message": "Lấy thông tin tiểu thuyết thành công",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "Tên tiểu thuyết (theo ngôn ngữ client)",
    "cover_image": "https://example.com/cover.jpg",
    "summary": {}, // JSON content from Plate editor (theo ngôn ngữ client hoặc bản gốc nếu không có bản dịch)
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
        "id": "genre-uuid",
        "name": "Fantasy"
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
        "id": "character-uuid",
        "name": "Nhân vật chính"
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

### 1.4 Cập nhật tiểu thuyết

```http
PUT /api/v1/novels/{id}
```

**Headers:**

- `Authorization: Bearer {token}`

**Body:**

```json
{
  "name": "Tên tiểu thuyết mới",
  "cover_image": "https://example.com/new-cover.jpg",
  "summary": json,
  "status": "COMPLETED",
  "genres": ["genre-uuid-1", "genre-uuid-2"],
  "is_featured": true,
  "is_completed": true
}
```

**Phản hồi:**

```json
{
  "success": true,
  "message": "Cập nhật tiểu thuyết thành công",
  "data": {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "Tên tiểu thuyết mới"
  },
  "error": null,
  "meta": {}
}
```

### 1.5 Xoá tiểu thuyết

```http
DELETE /api/v1/novels/{id}
```

**Headers:**

- `Authorization: Bearer {token}`

**Phản hồi:**

```json
{
  "success": true,
  "message": "Xoá tiểu thuyết thành công",
  "data": null,
  "error": null,
  "meta": {}
}
```

---

## 2. API Quản lý Volume (Volume Management)

### 2.1 Danh sách volumes của tiểu thuyết

```http
GET /api/v1/novels/{novel_id}/volumes
```

**Tham số:**

- `novel_id` (path, bắt buộc): UUID của tiểu thuyết
- `page` (query, tuỳ chọn): Số trang (mặc định: 1)
- `limit` (query, tuỳ chọn): Số lượng mỗi trang (mặc định: 20)
- `include_chapters` (query, tuỳ chọn): Bao gồm danh sách chapters (mặc định: false)

**Phản hồi:**

```json
{
  "success": true,
  "message": "Lấy danh sách volumes thành công",
  "data": [
    {
      "id": "volume-uuid",
      "novel_id": "novel-uuid",
      "volume_number": 1,
      "title": "Tập 1: Khởi đầu",
      "description": "Mô tả về tập 1",
      "cover_image": "https://example.com/volume1.jpg",
      "published_at": "2024-01-01T00:00:00Z",
      "is_public": true,
      "price_coins": 50,
      "chapter_count": 10,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "error": null,
  "meta": {
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 5,
      "total_pages": 1
    }
  }
}
```

### 2.2 Tạo volume mới

```http
POST /api/v1/novels/{novel_id}/volumes
```

**Headers:**

- `Authorization: Bearer {token}`

**Body:**

```json
{
  "volume_number": 2,
  "title": "Tập 2: Phát triển",
  "description": "Mô tả về tập 2",
  "cover_image": "https://example.com/volume2.jpg",
  "is_public": false,
  "price_coins": 60
}
```

**Phản hồi:**

```json
{
  "success": true,
  "message": "Tạo volume thành công",
  "data": {
    "id": "volume-uuid",
    "novel_id": "novel-uuid",
    "volume_number": 2,
    "title": "Tập 2: Phát triển",
    "created_at": "2024-01-01T00:00:00Z"
  },
  "error": null,
  "meta": {}
}
```

### 2.3 Lấy chi tiết volume

```http
GET /api/v1/volumes/{id}
```

**Tham số:**

- `include_chapters` (query, tuỳ chọn): Bao gồm danh sách chapters (mặc định: false)

**Phản hồi:**

```json
{
  "success": true,
  "message": "Lấy chi tiết volume thành công",
  "data": {
    "id": "volume-uuid",
    "novel_id": "novel-uuid",
    "volume_number": 1,
    "title": "Tập 1: Khởi đầu",
    "description": "Mô tả về tập 1",
    "cover_image": "https://example.com/volume1.jpg",
    "published_at": "2024-01-01T00:00:00Z",
    "is_public": true,
    "price_coins": 50,
    "chapter_count": 10,
    "chapters": [],
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "error": null,
  "meta": {}
}
```

### 2.4 Cập nhật volume

```http
PUT /api/v1/volumes/{id}
```

**Headers:**

- `Authorization: Bearer {token}`

**Body:**

```json
{
  "title": "Tập 1: Khởi đầu (Cập nhật)",
  "description": "Mô tả cập nhật",
  "cover_image": "https://example.com/volume1-updated.jpg",
  "is_public": true,
  "price_coins": 55
}
```

**Phản hồi:**

```json
{
  "success": true,
  "message": "Cập nhật volume thành công",
  "data": {
    "id": "volume-uuid",
    "title": "Tập 1: Khởi đầu (Cập nhật)",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "error": null,
  "meta": {}
}
```

### 2.5 Xóa volume

```http
DELETE /api/v1/volumes/{id}
```

**Headers:**

- `Authorization: Bearer {token}`

**Phản hồi:**

```json
{
  "success": true,
  "message": "Xóa volume thành công",
  "data": null,
  "error": null,
  "meta": {}
}
```

---

## 3. API Quản lý Chapter (Chapter Management)

### 3.1 Danh sách chapters của volume

```http
GET /api/v1/volumes/{volume_id}/chapters
```

**Tham số:**

- `volume_id` (path, bắt buộc): UUID của volume
- `page` (query, tuỳ chọn): Số trang (mặc định: 1)
- `limit` (query, tuỳ chọn): Số lượng mỗi trang (mặc định: 50)
- `include_content` (query, tuỳ chọn): Bao gồm nội dung chapter (mặc định: false)

**Phản hồi:**

```json
{
  "success": true,
  "message": "Lấy danh sách chapters thành công",
  "data": {
    "chapters": [
      {
        "id": "chapter-uuid",
        "volume_id": "volume-uuid",
        "chapter_number": 1,
        "title": "Chương 1: Bắt đầu cuộc hành trình",
        "published_at": "2024-01-01T00:00:00Z",
        "is_public": true,
        "is_draft": false,
        "price_coins": 10,
        "word_count": 2500,
        "reading_time_minutes": 10,
        "view_count": 500,
        "like_count": 45,
        "comment_count": 12,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 50,
      "total": 25,
      "total_pages": 1
    }
  },
  "error": null,
  "meta": {}
}
```

### 3.2 Tạo chapter mới

```http
POST /api/v1/volumes/{volume_id}/chapters
```

**Headers:**

- `Authorization: Bearer {token}`

**Body:**

```json
{
  "chapter_number": 2,
  "title": "Chương 2: Gặp gỡ đồng đội",
  "content": json,
  "is_public": false,
  "is_draft": true,
  "price_coins": 10,
  "content_warnings": ["bạo lực nhẹ"],
  "has_mature_content": false,
  "scheduled_publish_at": "2024-01-02T00:00:00Z"
}
```

**Phản hồi:**

```json
{
  "success": true,
  "message": "Tạo chapter thành công",
  "data": {
    "id": "chapter-uuid",
    "volume_id": "volume-uuid",
    "chapter_number": 2,
    "title": "Chương 2: Gặp gỡ đồng đội",
    "is_draft": true,
    "word_count": 2234,
    "character_count": 12456,
    "reading_time_minutes": 9,
    "created_at": "2024-01-01T00:00:00Z"
  },
  "error": null,
  "meta": {}
}
```

### 3.3 Lấy chi tiết chapter

```http
GET /api/v1/chapters/{id}
```

**Tham số:**

- `include_content` (query, tuỳ chọn): Bao gồm nội dung chapter (mặc định: true)

**Phản hồi:**

```json
{
  "success": true,
  "message": "Lấy chi tiết chapter thành công",
  "data": {
    "id": "chapter-uuid",
    "volume_id": "volume-uuid",
    "chapter_number": 1,
    "title": "Chương 1: Bắt đầu cuộc hành trình",
    "content": json,
    "published_at": "2024-01-01T00:00:00Z",
    "is_public": true,
    "is_draft": false,
    "price_coins": 10,
    "word_count": 2500,
    "character_count": 14500,
    "reading_time_minutes": 10,
    "view_count": 500,
    "like_count": 45,
    "comment_count": 12,
    "content_warnings": [],
    "has_mature_content": false,
    "version": 1,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "error": null,
  "meta": {}
}
```

### 3.4 Cập nhật chapter

```http
PUT /api/v1/chapters/{id}
```

**Headers:**

- `Authorization: Bearer {token}`

**Body:**

```json
{
  "title": "Chương 1: Bắt đầu cuộc hành trình (Cập nhật)",
  "content": json,
  "is_public": true,
  "is_draft": false,
  "price_coins": 12,
  "content_warnings": ["bạo lực nhẹ"],
  "has_mature_content": false
}
```

**Phản hồi:**

```json
{
  "success": true,
  "message": "Cập nhật chapter thành công",
  "data": {
    "id": "chapter-uuid",
    "title": "Chương 1: Bắt đầu cuộc hành trình (Cập nhật)",
    "word_count": 2634,
    "character_count": 15234,
    "reading_time_minutes": 11,
    "version": 2,
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "error": null,
  "meta": {}
}
```

### 3.5 Xóa chapter

```http
DELETE /api/v1/chapters/{id}
```

**Headers:**

- `Authorization: Bearer {token}`

**Phản hồi:**

```json
{
  "success": true,
  "message": "Xóa chapter thành công",
  "data": null,
  "error": null,
  "meta": {}
}
```

### 3.6 Publish chapter

```http
POST /api/v1/chapters/{id}/publish
```

**Headers:**

- `Authorization: Bearer {token}`

**Body:**

```json
{
  "publish_at": "2024-01-01T00:00:00Z"
}
```

**Phản hồi:**

```json
{
  "success": true,
  "message": "Publish chapter thành công",
  "data": {
    "id": "chapter-uuid",
    "is_public": true,
    "is_draft": false,
    "published_at": "2024-01-01T00:00:00Z"
  },
  "error": null,
  "meta": {}
}
```

### 3.7 Unpublish chapter

```http
POST /api/v1/chapters/{id}/unpublish
```

**Headers:**

- `Authorization: Bearer {token}`

**Phản hồi:**

```json
{
  "success": true,
  "message": "Unpublish chapter thành công",
  "data": {
    "id": "chapter-uuid",
    "is_public": false,
    "published_at": null
  },
  "error": null,
  "meta": {}
}
```

---

## 4. API Đóng góp Bản dịch (Translation Contributions)

### 4.1 Gửi đóng góp bản dịch

```http
POST /api/v1/translations/contribute
```

**Headers:**

- `Authorization: Bearer {token}`

**Body:**

```json
{
  "reference_type": "novel_chapter",
  "reference_id": "123e4567-e89b-12d3-a456-426614174000",
  "title": "Chương 1: Khởi đầu mới",
  "content": json,
  "source_language": "en",
  "target_language": "vi",
  "is_machine_translation": false
}
```

**Ví dụ cho novel translation:**

```json
{
  "reference_type": "novel",
  "reference_id": "123e4567-e89b-12d3-a456-426614174000",
  "title": "Tên Tiểu Thuyết Dịch",
  "content": json, // Summary content only
  "source_language": "en",
  "target_language": "vi",
  "is_machine_translation": false
}
```

**Phản hồi:**

```json
{
  "success": true,
  "message": "Đóng góp bản dịch đã được gửi thành công",
  "data": {
    "id": "translation-contribution-uuid",
    "reference_type": "novel_chapter",
    "reference_id": "123e4567-e89b-12d3-a456-426614174000",
    "title": "Chương 1: Khởi đầu mới",
    "source_language": "en",
    "target_language": "vi",
    "status": "pending",
    "user_id": "user-uuid",
    "created_at": "2024-01-01T00:00:00Z"
  },
  "error": null,
  "meta": {}
}
```

### 4.2 Danh sách đóng góp của tôi

```http
GET /api/v1/translations/my-contributions
```

**Headers:**

- `Authorization: Bearer {token}`

**Tham số:**

- `status` (query, tuỳ chọn): Lọc theo trạng thái (pending, approved, rejected)
- `page` (query, tuỳ chọn): Số trang (mặc định: 1)
- `limit` (query, tuỳ chọn): Số lượng mỗi trang (mặc định: 20)

**Phản hồi:**

```json
{
  "success": true,
  "message": "Lấy danh sách đóng góp thành công",
  "data": {
    "contributions": [
      {
        "id": "contribution-uuid",
        "reference_type": "novel_chapter",
        "reference_id": "chapter-uuid",
        "title": "Chương 1: Khởi đầu mới",
        "source_language": "en",
        "target_language": "vi",
        "status": "pending",
        "upvotes": 5,
        "downvotes": 1,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 50,
      "total_pages": 3
    }
  },
  "error": null,
  "meta": {}
}
```

### 4.3 Cập nhật đóng góp (chỉ khi status = pending)

```http
PUT /api/v1/translations/contributions/{id}
```

**Headers:**

- `Authorization: Bearer {token}`

**Body:**

```json
{
  "title": "Chương 1: Khởi đầu mới (Cập nhật)",
  "content": json
}
```

**Phản hồi:**

```json
{
  "success": true,
  "message": "Đóng góp đã được cập nhật thành công",
  "data": {
    "id": "contribution-uuid",
    "title": "Chương 1: Khởi đầu mới (Cập nhật)",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "error": null,
  "meta": {}
}
```

### 4.4 Vote cho đóng góp bản dịch

```http
POST /api/v1/translations/contributions/{id}/vote
```

**Headers:**

- `Authorization: Bearer {token}`

**Body:**

```json
{
  "vote_type": "upvote"
}
```

**Phản hồi:**

```json
{
  "success": true,
  "message": "Vote đã được ghi nhận",
  "data": {
    "contribution_id": "contribution-uuid",
    "vote_type": "upvote",
    "total_upvotes": 6,
    "total_downvotes": 1
  },
  "error": null,
  "meta": {}
}
```

### 4.5 Lấy chi tiết đóng góp

```http
GET /api/v1/translations/contributions/{id}
```

**Phản hồi:**

```json
{
  "success": true,
  "message": "Lấy chi tiết đóng góp thành công",
  "data": {
    "id": "contribution-uuid",
    "reference_type": "novel_chapter",
    "reference_id": "chapter-uuid",
    "title": "Chương 1: Khởi đầu mới",
    "content": json,
    "source_language": "en",
    "target_language": "vi",
    "is_machine_translation": false,
    "status": "pending",
    "user_id": "user-uuid",
    "tenant_id": "tenant-uuid",
    "upvotes": 6,
    "downvotes": 1,
    "user_vote": "upvote",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "error": null,
  "meta": {}
}
```

### 4.6 Danh sách đóng góp chờ duyệt (Moderator)

```http
GET /api/v1/translations/pending
```

**Headers:**

- `Authorization: Bearer {token}`

**Tham số:**

- `language` (query, tuỳ chọn): Lọc theo ngôn ngữ đích
- `reference_type` (query, tuỳ chọn): Lọc theo loại nội dung
- `page` (query, tuỳ chọn): Số trang (mặc định: 1)
- `limit` (query, tuỳ chọn): Số lượng mỗi trang (mặc định: 20)

**Phản hồi:**

```json
{
  "success": true,
  "message": "Lấy danh sách đóng góp chờ duyệt thành công",
  "data": {
    "contributions": [
      {
        "id": "contribution-uuid",
        "reference_type": "novel_chapter",
        "reference_id": "chapter-uuid",
        "title": "Chương 1: Khởi đầu mới",
        "source_language": "en",
        "target_language": "vi",
        "user_id": "user-uuid",
        "upvotes": 6,
        "downvotes": 1,
        "is_machine_translation": false,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 15,
      "total_pages": 1
    }
  },
  "error": null,
  "meta": {}
}
```

### 4.7 Duyệt đóng góp bản dịch (Moderator)

```http
POST /api/v1/translations/contributions/{id}/review
```

**Headers:**

- `Authorization: Bearer {token}`

**Body:**

```json
{
  "action": "approve",
  "rejection_reason": null
}
```

**Phản hồi:**

```json
{
  "success": true,
  "message": "Đóng góp đã được duyệt thành công",
  "data": {
    "id": "contribution-uuid",
    "status": "approved",
    "reviewer_id": "moderator-uuid",
    "reviewed_at": "2024-01-01T00:00:00Z"
  },
  "error": null,
  "meta": {}
}
```

### 4.8 Từ chối đóng góp bản dịch (Moderator)

```http
POST /api/v1/translations/contributions/{id}/review
```

**Headers:**

- `Authorization: Bearer {token}`

**Body:**

```json
{
  "action": "reject",
  "rejection_reason": "Chất lượng bản dịch không đạt yêu cầu. Vui lòng kiểm tra lại ngữ pháp và ý nghĩa."
}
```

**Phản hồi:**

```json
{
  "success": true,
  "message": "Đóng góp đã được từ chối",
  "data": {
    "id": "contribution-uuid",
    "status": "rejected",
    "rejection_reason": "Chất lượng bản dịch không đạt yêu cầu. Vui lòng kiểm tra lại ngữ pháp và ý nghĩa.",
    "reviewer_id": "moderator-uuid",
    "reviewed_at": "2024-01-01T00:00:00Z"
  },
  "error": null,
  "meta": {}
}
```

### 4.9 Danh sách đóng góp cho một nội dung

```http
GET /api/v1/novels/{novel_id}/chapters/{chapter_id}/translations/contributions
```

**Tham số:**

- `language` (query, tuỳ chọn): Lọc theo ngôn ngữ đích
- `status` (query, tuỳ chọn): Lọc theo trạng thái (pending, approved, rejected)

**Phản hồi:**

```json
{
  "success": true,
  "message": "Lấy danh sách đóng góp cho chương thành công",
  "data": {
    "chapter_id": "chapter-uuid",
    "contributions": [
      {
        "id": "contribution-uuid",
        "title": "Chương 1: Khởi đầu mới",
        "target_language": "vi",
        "status": "approved",
        "user_id": "user-uuid",
        "upvotes": 10,
        "downvotes": 2,
        "is_machine_translation": false,
        "created_at": "2024-01-01T00:00:00Z"
      }
    ]
  },
  "error": null,
  "meta": {}
}
```

---

## Workflow Đóng góp Bản dịch

### 1. Submission Flow

1. User gửi đóng góp bản dịch thông qua API
2. System validate dữ liệu và tạo record với status = 'pending'
3. Community có thể vote upvote/downvote
4. Moderator review và approve/reject

### 2. Approval Flow

1. **Approved**:
   - Nếu `reference_type = 'novel'`: Tạo/cập nhật record trong `novel_translation` với:
     - `title` từ contribution.title
     - `summary` từ contribution.content
   - Nếu `reference_type = 'novel_chapter'`: Tạo bản sao chapter mới hoặc cập nhật chapter hiện tại
2. **Rejected**: Contributor nhận được feedback và có thể resubmit

### 3. Quality Control

- Community voting system
- Moderator review trước khi publish
- Soft delete cho các contribution không phù hợp

---

## Quyền hạn API (API Permissions)

Các API được bảo vệ bằng system quyền dựa trên enum đã định nghĩa:

### Novel Management

- **Tạo novel**: `PermContentCreateNovel` (tenant permission)
- **Cập nhật novel**: `PermContentUpdateNovel` (tenant permission)
- **Xóa novel**: `PermContentDeleteNovel` (tenant permission)

### Volume Management

- **Tạo volume**: `PermNovelVolumeCreate` (tenant permission)
- **Cập nhật volume**: `PermNovelVolumeUpdate` (tenant permission)
- **Xóa volume**: `PermNovelVolumeDelete` (tenant permission)

### Chapter Management

- **Tạo chapter**: `PermNovelChapterCreate` (tenant permission)
- **Cập nhật chapter**: `PermNovelChapterUpdate` (tenant permission)
- **Xóa chapter**: `PermNovelChapterDelete` (tenant permission)
- **Publish/Unpublish**: `PermContentPublish`, `PermContentUnpublish` (tenant permission)

### Translation Contributions

- **Submit translation**: `PermTranslationSubmit` (global permission)
- **Update own translation**: `PermTranslationUpdateSelf` (global permission)
- **Vote on translation**: `PermTranslationVote` (global permission)
- **Review translations**: Moderator role (`RoleModerator`, `RoleAdmin`, `RoleSuperAdmin`)
- **Approve/Reject**: Moderator role (`RoleModerator`, `RoleAdmin`, `RoleSuperAdmin`)

### Public Access

- **Xem novels**: `PermContentViewPublic` (global permission)
- **Đọc content**: `PermContentReadNovel` (global permission)

**Lưu ý**: Tenant permissions chỉ áp dụng trong phạm vi tenant của user, global permissions áp dụng toàn hệ thống.
