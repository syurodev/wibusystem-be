# API Client Guide - Novel Management System

## Tổng quan

Tài liệu này hướng dẫn cách sử dụng các API cho hệ thống quản lý tiểu thuyết, bao gồm 6 domain chính:
- **Genre** (Thể loại)
- **Author** (Tác giả)
- **Artist** (Hoạ sĩ)
- **Novel** (Tiểu thuyết)
- **Volume** (Tập)
- **Chapter** (Chương)

**Base URL**: `http://localhost:8080/api/v1`

**Response Format**: Tất cả các API đều trả về chuẩn `StandardResponse`:

```json
{
  "success": true,
  "message": "operation.success",
  "data": { ... },
  "meta": { ... }
}
```

**Error Response**:
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "error.message",
    "details": "Chi tiết lỗi (nếu có)"
  }
}
```

---

## 1. Genre API (Thể loại)

### 1.1. Lấy danh sách Genres

**Endpoint**: `GET /api/v1/genres`

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| page | int | No | 1 | Số trang |
| limit | int | No | 20 | Số items mỗi trang (max: 100) |
| search | string | No | - | Tìm kiếm theo tên |
| sort_by | string | No | - | Sắp xếp theo: `name`, `views`, `series`, `created`, `updated` |
| sort_order | string | No | desc | Thứ tự: `asc`, `desc` |
| active_only | bool | No | false | Chỉ lấy genres đang active |

**Response**:
```json
{
  "success": true,
  "message": "genre.list_success",
  "data": [
    {
      "id": "uuid-string",
      "name": "Fantasy",
      "series_count": 150,
      "active_readers": 5000,
      "total_views": 1000000,
      "trend": "rising",
      "description": "Mô tả thể loại Fantasy",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-02T00:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total_items": 100,
    "total_pages": 5
  }
}
```

**Curl Example**:
```bash
# Lấy danh sách genres với pagination
curl -X GET "http://localhost:8080/api/v1/genres?page=1&limit=20"

# Tìm kiếm genres có tên chứa "fantasy"
curl -X GET "http://localhost:8080/api/v1/genres?search=fantasy"

# Sắp xếp theo số lượt xem giảm dần
curl -X GET "http://localhost:8080/api/v1/genres?sort_by=views&sort_order=desc"

# Chỉ lấy genres đang active
curl -X GET "http://localhost:8080/api/v1/genres?active_only=true"
```

---

### 1.2. Lấy chi tiết Genre

**Endpoint**: `GET /api/v1/genres/:id`

**Path Parameters**:
- `id` (string, required): UUID của genre

**Response**:
```json
{
  "success": true,
  "message": "genre.get_success",
  "data": {
    "id": "uuid-string",
    "name": "Fantasy",
    "slug": "fantasy",
    "description": "Mô tả chi tiết về Fantasy",
    "parent_id": "parent-uuid-string",
    "display_order": 1,
    "is_active": true,
    "series_count": 150,
    "active_readers": 5000,
    "total_views": 1000000,
    "trend": "rising",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-02T00:00:00Z"
  }
}
```

**Curl Example**:
```bash
curl -X GET "http://localhost:8080/api/v1/genres/550e8400-e29b-41d4-a716-446655440000"
```

---

### 1.3. Tạo Genre mới

**Endpoint**: `POST /api/v1/genres`

**Headers**:
- `Content-Type: application/json`
- `Authorization: Bearer <token>` (nếu có)

**Request Body**:
```json
{
  "name": "Xuanhuan",
  "description": "Thể loại tiên hiệp phương Đông",
  "parent_id": "parent-uuid-string"
}
```

**Validation**:
- `name`: bắt buộc, độ dài 1-100 ký tự
- `description`: tùy chọn, max 1000 ký tự
- `parent_id`: tùy chọn, UUID string

**Response**:
```json
{
  "success": true,
  "message": "genre.created_success",
  "data": {
    "id": "new-uuid",
    "name": "Xuanhuan",
    "slug": "xuanhuan",
    "description": "Thể loại tiên hiệp phương Đông",
    "parent_id": "parent-uuid-string",
    "display_order": 0,
    "is_active": false,
    "series_count": 0,
    "active_readers": 0,
    "total_views": 0,
    "trend": "stable",
    "created_at": "2024-01-03T00:00:00Z",
    "updated_at": "2024-01-03T00:00:00Z"
  }
}
```

**Curl Example**:
```bash
curl -X POST "http://localhost:8080/api/v1/genres" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Xuanhuan",
    "description": "Thể loại tiên hiệp phương Đông"
  }'
```

**Error Responses**:
- `400 INVALID_INPUT`: Dữ liệu đầu vào không hợp lệ
- `409 SLUG_EXISTS`: Slug đã tồn tại
- `400 PARENT_NOT_FOUND`: Parent genre không tồn tại
- `401 UNAUTHORIZED`: Chưa đăng nhập

---

### 1.4. Cập nhật Genre

**Endpoint**: `PUT /api/v1/genres/:id`

**Request Body**:
```json
{
  "name": "Xuanhuan Updated",
  "description": "Mô tả mới",
  "parent_id": "new-parent-uuid",
  "display_order": 5,
  "is_active": true
}
```

**Response**: Giống như response của GET genre detail

**Curl Example**:
```bash
curl -X PUT "http://localhost:8080/api/v1/genres/550e8400-e29b-41d4-a716-446655440000" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Xuanhuan Updated",
    "description": "Mô tả mới",
    "display_order": 5,
    "is_active": true
  }'
```

**Error Responses**:
- `404 GENRE_NOT_FOUND`: Genre không tồn tại
- `409 SLUG_EXISTS`: Slug mới đã tồn tại
- `400 CIRCULAR_REFERENCE`: Tham chiếu vòng (parent reference)

---

### 1.5. Xóa Genre

**Endpoint**: `DELETE /api/v1/genres/:id`

**Response**:
```json
{
  "success": true,
  "message": "genre.deleted_success"
}
```

**Curl Example**:
```bash
curl -X DELETE "http://localhost:8080/api/v1/genres/550e8400-e29b-41d4-a716-446655440000"
```

**Error Responses**:
- `404 GENRE_NOT_FOUND`: Genre không tồn tại
- `409 GENRE_IN_USE`: Genre đang được sử dụng bởi novels
- `409 GENRE_HAS_CHILDREN`: Genre có các genre con

---

## 2. Author API (Tác giả)

### 2.1. Lấy danh sách Authors

**Endpoint**: `GET /api/v1/authors`

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| page | int | No | 1 | Số trang |
| limit | int | No | 20 | Số items mỗi trang (max: 100) |
| search | string | No | - | Tìm kiếm theo tên |
| sort_by | string | No | - | Sắp xếp theo: `name`, `views`, `novels`, `created` |
| sort_order | string | No | desc | Thứ tự: `asc`, `desc` |
| is_verified | bool | No | - | Lọc theo trạng thái verified |

**Response**:
```json
{
  "success": true,
  "message": "author.list_success",
  "data": [
    {
      "id": "uuid-string",
      "name": "Nhĩ Căn",
      "description": "Tiểu sử tác giả",
      "novel_count": 25,
      "total_views": 5000000,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total_items": 50,
    "total_pages": 3
  }
}
```

**Curl Example**:
```bash
# Lấy danh sách authors
curl -X GET "http://localhost:8080/api/v1/authors?page=1&limit=20"

# Tìm kiếm author theo tên
curl -X GET "http://localhost:8080/api/v1/authors?search=nhĩ+căn"

# Sắp xếp theo số lượng novel giảm dần
curl -X GET "http://localhost:8080/api/v1/authors?sort_by=novels&sort_order=desc"

# Chỉ lấy authors đã verified
curl -X GET "http://localhost:8080/api/v1/authors?is_verified=true"
```

---

### 2.2. Lấy chi tiết Author

**Endpoint**: `GET /api/v1/authors/:id`

**Response**:
```json
{
  "success": true,
  "message": "author.get_success",
  "data": {
    "id": "uuid-string",
    "name": "Nhĩ Căn",
    "slug": "nhi-can",
    "description": "Tiểu sử chi tiết của tác giả",
    "avatar_url": "https://example.com/avatar.jpg",
    "social_links": "{\"facebook\": \"fb.com/author\", \"twitter\": \"@author\"}",
    "novel_count": 25,
    "total_chapters": 5000,
    "total_views": 5000000,
    "follower_count": 10000,
    "is_verified": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-02T00:00:00Z"
  }
}
```

**Curl Example**:
```bash
curl -X GET "http://localhost:8080/api/v1/authors/550e8400-e29b-41d4-a716-446655440000"
```

---

### 2.3. Tạo Author mới

**Endpoint**: `POST /api/v1/authors`

**Request Body**:
```json
{
  "name": "Tân Tác Giả",
  "biography": "Tiểu sử tác giả mới",
  "avatar_url": "https://example.com/avatar.jpg",
  "social_links": "{\"facebook\": \"fb.com/newauthor\"}"
}
```

**Validation**:
- `name`: bắt buộc, độ dài 1-200 ký tự
- `biography`: tùy chọn, max 5000 ký tự
- `avatar_url`: tùy chọn, phải là URL hợp lệ
- `social_links`: tùy chọn, phải là JSON string hợp lệ

**Response**:
```json
{
  "success": true,
  "message": "author.created_success",
  "data": {
    "id": "new-uuid",
    "name": "Tân Tác Giả",
    "slug": "tan-tac-gia",
    "description": "Tiểu sử tác giả mới",
    "avatar_url": "https://example.com/avatar.jpg",
    "social_links": "{\"facebook\": \"fb.com/newauthor\"}",
    "novel_count": 0,
    "total_chapters": 0,
    "total_views": 0,
    "follower_count": 0,
    "is_verified": false,
    "created_at": "2024-01-03T00:00:00Z",
    "updated_at": "2024-01-03T00:00:00Z"
  }
}
```

**Curl Example**:
```bash
curl -X POST "http://localhost:8080/api/v1/authors" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Tân Tác Giả",
    "biography": "Tiểu sử tác giả mới",
    "avatar_url": "https://example.com/avatar.jpg"
  }'
```

---

### 2.4. Cập nhật Author

**Endpoint**: `PUT /api/v1/authors/:id`

**Request Body**:
```json
{
  "name": "Tên Mới",
  "biography": "Tiểu sử cập nhật",
  "avatar_url": "https://example.com/new-avatar.jpg",
  "social_links": "{\"facebook\": \"fb.com/updated\"}"
}
```

**Curl Example**:
```bash
curl -X PUT "http://localhost:8080/api/v1/authors/550e8400-e29b-41d4-a716-446655440000" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Tên Mới",
    "biography": "Tiểu sử cập nhật"
  }'
```

---

### 2.5. Xóa Author

**Endpoint**: `DELETE /api/v1/authors/:id`

**Response**:
```json
{
  "success": true,
  "message": "author.deleted_success"
}
```

**Curl Example**:
```bash
curl -X DELETE "http://localhost:8080/api/v1/authors/550e8400-e29b-41d4-a716-446655440000"
```

**Error Responses**:
- `404 AUTHOR_NOT_FOUND`: Author không tồn tại
- `409 AUTHOR_IN_USE`: Author đang được sử dụng bởi novels

---

## 3. Artist API (Hoạ sĩ)

### 3.1. Lấy danh sách Artists

**Endpoint**: `GET /api/v1/artists`

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| page | int | No | 1 | Số trang |
| limit | int | No | 20 | Số items mỗi trang (max: 100) |
| search | string | No | - | Tìm kiếm theo tên |
| sort_by | string | No | - | Sắp xếp theo: `name`, `novels`, `created` |
| sort_order | string | No | desc | Thứ tự: `asc`, `desc` |
| specialization | string | No | - | Lọc theo chuyên môn |
| is_verified | bool | No | - | Lọc theo trạng thái verified |

**Response**:
```json
{
  "success": true,
  "message": "artist.list_success",
  "data": [
    {
      "id": "uuid-string",
      "name": "Hoạ Sĩ ABC",
      "description": "Tiểu sử hoạ sĩ",
      "novel_count": 15,
      "specialization": "cover_artist",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total_items": 30,
    "total_pages": 2
  }
}
```

**Specialization Values**:
- `cover_artist`: Hoạ sĩ vẽ bìa
- `illustrator`: Hoạ sĩ minh họa
- `character_designer`: Nhà thiết kế nhân vật
- `manga_artist`: Hoạ sĩ manga

**Curl Example**:
```bash
# Lấy danh sách artists
curl -X GET "http://localhost:8080/api/v1/artists?page=1&limit=20"

# Tìm kiếm artist theo tên
curl -X GET "http://localhost:8080/api/v1/artists?search=hoạ+sĩ"

# Lọc theo chuyên môn cover_artist
curl -X GET "http://localhost:8080/api/v1/artists?specialization=cover_artist"

# Sắp xếp theo số lượng novel
curl -X GET "http://localhost:8080/api/v1/artists?sort_by=novels&sort_order=desc"
```

---

### 3.2. Lấy chi tiết Artist

**Endpoint**: `GET /api/v1/artists/:id`

**Response**:
```json
{
  "success": true,
  "message": "artist.get_success",
  "data": {
    "id": "uuid-string",
    "name": "Hoạ Sĩ ABC",
    "slug": "hoa-si-abc",
    "description": "Tiểu sử chi tiết của hoạ sĩ",
    "avatar_url": "https://example.com/avatar.jpg",
    "social_links": "{\"instagram\": \"@artist\"}",
    "specialization": "cover_artist",
    "novel_count": 15,
    "artwork_count": 200,
    "follower_count": 5000,
    "is_verified": true,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-02T00:00:00Z"
  }
}
```

**Curl Example**:
```bash
curl -X GET "http://localhost:8080/api/v1/artists/550e8400-e29b-41d4-a716-446655440000"
```

---

### 3.3. Tạo Artist mới

**Endpoint**: `POST /api/v1/artists`

**Request Body**:
```json
{
  "name": "Hoạ Sĩ Mới",
  "biography": "Tiểu sử hoạ sĩ",
  "avatar_url": "https://example.com/avatar.jpg",
  "social_links": "{\"instagram\": \"@newartist\"}",
  "specialization": "illustrator"
}
```

**Validation**:
- `name`: bắt buộc, độ dài 1-200 ký tự
- `biography`: tùy chọn, max 5000 ký tự
- `avatar_url`: tùy chọn, phải là URL hợp lệ
- `social_links`: tùy chọn, phải là JSON string hợp lệ
- `specialization`: tùy chọn, max 100 ký tự

**Response**:
```json
{
  "success": true,
  "message": "artist.created_success",
  "data": {
    "id": "new-uuid",
    "name": "Hoạ Sĩ Mới",
    "slug": "hoa-si-moi",
    "description": "Tiểu sử hoạ sĩ",
    "avatar_url": "https://example.com/avatar.jpg",
    "social_links": "{\"instagram\": \"@newartist\"}",
    "specialization": "illustrator",
    "novel_count": 0,
    "artwork_count": 0,
    "follower_count": 0,
    "is_verified": false,
    "created_at": "2024-01-03T00:00:00Z",
    "updated_at": "2024-01-03T00:00:00Z"
  }
}
```

**Curl Example**:
```bash
curl -X POST "http://localhost:8080/api/v1/artists" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Hoạ Sĩ Mới",
    "biography": "Tiểu sử hoạ sĩ",
    "specialization": "illustrator"
  }'
```

---

### 3.4. Cập nhật Artist

**Endpoint**: `PUT /api/v1/artists/:id`

**Request Body**:
```json
{
  "name": "Tên Mới",
  "biography": "Tiểu sử cập nhật",
  "avatar_url": "https://example.com/new-avatar.jpg",
  "social_links": "{\"instagram\": \"@updated\"}",
  "specialization": "character_designer"
}
```

**Curl Example**:
```bash
curl -X PUT "http://localhost:8080/api/v1/artists/550e8400-e29b-41d4-a716-446655440000" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Tên Mới",
    "biography": "Tiểu sử cập nhật",
    "specialization": "character_designer"
  }'
```

---

### 3.5. Xóa Artist

**Endpoint**: `DELETE /api/v1/artists/:id`

**Response**:
```json
{
  "success": true,
  "message": "artist.deleted_success"
}
```

**Curl Example**:
```bash
curl -X DELETE "http://localhost:8080/api/v1/artists/550e8400-e29b-41d4-a716-446655440000"
```

**Error Responses**:
- `404 ARTIST_NOT_FOUND`: Artist không tồn tại
- `409 ARTIST_IN_USE`: Artist đang được sử dụng bởi novels

---

## 4. Novel API (Tiểu thuyết)

### 4.1. Lấy danh sách Novels

**Endpoint**: `GET /api/v1/novels`

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| page | int | No | 1 | Số trang |
| limit | int | No | 20 | Số items mỗi trang (max: 100) |
| search | string | No | - | Tìm kiếm trong title và synopsis |
| status | string | No | - | Filter theo status: `draft`, `ongoing`, `completed`, `hiatus`, `dropped` |
| original_language | string | No | - | Filter theo ngôn ngữ gốc (ISO 639-1: vi, en, zh, ja, ko...) |
| sort_by | string | No | created_at | Sắp xếp theo: `created_at`, `rating`, `views`, `last_chapter` |
| sort_order | string | No | desc | Thứ tự: `asc`, `desc` |

**Response**:
```json
{
  "success": true,
  "message": "novel.list_success",
  "data": [
    {
      "id": "uuid-string",
      "title": "Tiên Nghịch",
      "slug": "tien-nghich",
      "cover_image_url": "https://example.com/cover.jpg",
      "thumbnail_url": "https://example.com/thumb.jpg",
      "status": "ongoing",
      "total_chapters": 2500,
      "view_count": 10000000,
      "favorite_count": 50000,
      "rating_average": 4.8,
      "rating_count": 5000,
      "last_chapter_at": "2024-01-15T10:00:00Z",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-15T10:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 20,
    "total_items": 500,
    "total_pages": 25
  }
}
```

**Curl Examples**:
```bash
# Lấy danh sách novels
curl -X GET "http://localhost:8080/api/v1/novels?page=1&limit=20"

# Tìm kiếm novels
curl -X GET "http://localhost:8080/api/v1/novels?search=tiên+nghịch"

# Filter theo status ongoing
curl -X GET "http://localhost:8080/api/v1/novels?status=ongoing"

# Filter theo ngôn ngữ gốc tiếng Trung
curl -X GET "http://localhost:8080/api/v1/novels?original_language=zh"

# Sắp xếp theo rating giảm dần
curl -X GET "http://localhost:8080/api/v1/novels?sort_by=rating&sort_order=desc"

# Sắp xếp theo lượt xem
curl -X GET "http://localhost:8080/api/v1/novels?sort_by=views&sort_order=desc"

# Combine: Tìm novels ongoing, sort theo views
curl -X GET "http://localhost:8080/api/v1/novels?status=ongoing&sort_by=views&sort_order=desc&page=1"
```

---

### 4.2. Lấy chi tiết Novel

**Endpoint**: `GET /api/v1/novels/:id`

**Path Parameters**:
- `id` (string, required): UUID hoặc slug của novel

**Response**:
```json
{
  "success": true,
  "message": "novel.get_success",
  "data": {
    "id": "uuid-string",
    "title": "Tiên Nghịch",
    "slug": "tien-nghich",
    "synopsis": "Thuận thiên giả, sống. Nghịch thiên giả, chết. Đây là câu nói vạn cổ bất biến...",
    "cover_image_url": "https://example.com/cover.jpg",
    "thumbnail_url": "https://example.com/thumb.jpg",
    "status": "ongoing",
    "original_language": "zh",
    "original_title": "仙逆",
    "total_volumes": 6,
    "total_chapters": 2500,
    "total_words": 5000000,
    "view_count": 10000000,
    "favorite_count": 50000,
    "rating_average": 4.8,
    "rating_count": 5000,
    "metadata": "{\"tags\": [\"cultivation\", \"reincarnation\"]}",
    "first_published_at": "2024-01-01T00:00:00Z",
    "last_chapter_at": "2024-01-15T10:00:00Z",
    "completed_at": null,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-15T10:00:00Z"
  }
}
```

**Curl Examples**:
```bash
# Lấy theo UUID
curl -X GET "http://localhost:8080/api/v1/novels/550e8400-e29b-41d4-a716-446655440000"

# Lấy theo slug
curl -X GET "http://localhost:8080/api/v1/novels/tien-nghich"
```

---

### 4.3. Tạo Novel mới

**Endpoint**: `POST /api/v1/novels`

**Headers**:
- `Content-Type: application/json`
- `Authorization: Bearer <token>` (nếu có)

**Request Body**:
```json
{
  "title": "Tiên Nghịch",
  "synopsis": "Thuận thiên giả, sống. Nghịch thiên giả, chết...",
  "cover_image_url": "https://example.com/cover.jpg",
  "thumbnail_url": "https://example.com/thumb.jpg",
  "status": "draft",
  "original_language": "zh",
  "original_title": "仙逆",
  "metadata": "{\"tags\": [\"cultivation\", \"reincarnation\"]}"
}
```

**Validation**:
- `title`: bắt buộc, độ dài 1-500 ký tự
- `synopsis`: tùy chọn, max 10000 ký tự
- `cover_image_url`: tùy chọn, phải là URL hợp lệ
- `thumbnail_url`: tùy chọn, phải là URL hợp lệ
- `status`: bắt buộc, phải là: `draft`, `ongoing`, `completed`, `hiatus`, `dropped`
- `original_language`: tùy chọn, phải là mã ISO 639-1 (2 ký tự: vi, en, zh, ja, ko...)
- `original_title`: tùy chọn, max 500 ký tự
- `metadata`: tùy chọn, phải là JSON string hợp lệ

**Novel Status**:
- `draft`: Bản nháp (chưa publish)
- `ongoing`: Đang cập nhật
- `completed`: Đã hoàn thành
- `hiatus`: Tạm ngưng
- `dropped`: Đã drop (ngưng dịch/viết)

**Response**:
```json
{
  "success": true,
  "message": "novel.created_success",
  "data": {
    "id": "new-uuid",
    "title": "Tiên Nghịch",
    "slug": "tien-nghich",
    "synopsis": "Thuận thiên giả, sống. Nghịch thiên giả, chết...",
    "cover_image_url": "https://example.com/cover.jpg",
    "thumbnail_url": "https://example.com/thumb.jpg",
    "status": "draft",
    "original_language": "zh",
    "original_title": "仙逆",
    "total_volumes": 0,
    "total_chapters": 0,
    "total_words": 0,
    "view_count": 0,
    "favorite_count": 0,
    "rating_average": 0,
    "rating_count": 0,
    "metadata": "{\"tags\": [\"cultivation\", \"reincarnation\"]}",
    "first_published_at": null,
    "last_chapter_at": null,
    "completed_at": null,
    "created_at": "2024-01-16T00:00:00Z",
    "updated_at": "2024-01-16T00:00:00Z"
  }
}
```

**Curl Example**:
```bash
curl -X POST "http://localhost:8080/api/v1/novels" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Tiên Nghịch",
    "synopsis": "Thuận thiên giả, sống. Nghịch thiên giả, chết...",
    "status": "draft",
    "original_language": "zh",
    "original_title": "仙逆"
  }'
```

**Error Responses**:
- `400 INVALID_INPUT`: Dữ liệu đầu vào không hợp lệ (status sai, JSON không hợp lệ...)
- `409 SLUG_EXISTS`: Slug đã tồn tại (trùng title)
- `401 UNAUTHORIZED`: Chưa đăng nhập

---

### 4.4. Cập nhật Novel

**Endpoint**: `PUT /api/v1/novels/:id`

**Request Body**:
```json
{
  "title": "Tiên Nghịch (Updated)",
  "synopsis": "Synopsis mới...",
  "cover_image_url": "https://example.com/new-cover.jpg",
  "thumbnail_url": "https://example.com/new-thumb.jpg",
  "status": "ongoing",
  "original_language": "zh",
  "original_title": "仙逆",
  "metadata": "{\"tags\": [\"cultivation\", \"reincarnation\", \"revenge\"]}"
}
```

**Note**:
- Khi update status từ bất kỳ status nào sang `completed`, hệ thống sẽ tự động set `completed_at` timestamp
- Slug sẽ được tự động update nếu title thay đổi

**Response**: Giống như response của GET novel detail

**Curl Example**:
```bash
curl -X PUT "http://localhost:8080/api/v1/novels/550e8400-e29b-41d4-a716-446655440000" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Tiên Nghịch (Updated)",
    "synopsis": "Synopsis mới...",
    "status": "ongoing"
  }'
```

**Error Responses**:
- `404 NOVEL_NOT_FOUND`: Novel không tồn tại
- `400 INVALID_INPUT`: Dữ liệu không hợp lệ
- `409 SLUG_EXISTS`: Slug mới đã tồn tại

---

### 4.5. Xóa Novel

**Endpoint**: `DELETE /api/v1/novels/:id`

**Response**:
```json
{
  "success": true,
  "message": "novel.deleted_success"
}
```

**Curl Example**:
```bash
curl -X DELETE "http://localhost:8080/api/v1/novels/550e8400-e29b-41d4-a716-446655440000"
```

**Error Responses**:
- `404 NOVEL_NOT_FOUND`: Novel không tồn tại

**Note**: Đây là soft delete, dữ liệu vẫn được giữ trong database với `deleted_at` timestamp

---

### 4.6. Tăng View Count

**Endpoint**: `POST /api/v1/novels/:id/view`

**Description**: Endpoint này dùng để track lượt xem novel. Gọi endpoint này mỗi khi user xem novel.

**Response**:
```json
{
  "success": true,
  "message": "novel.view_incremented"
}
```

**Curl Example**:
```bash
curl -X POST "http://localhost:8080/api/v1/novels/550e8400-e29b-41d4-a716-446655440000/view"
```

**Error Responses**:
- `400 INVALID_ID`: UUID không hợp lệ
- `500 INCREMENT_FAILED`: Lỗi khi increment view count

**Use Case**:
```javascript
// Gọi khi user mở trang đọc novel
async function trackNovelView(novelId) {
  await fetch(`http://localhost:8080/api/v1/novels/${novelId}/view`, {
    method: 'POST'
  });
}
```

---

## 5. Volume API (Tập)

### Tổng quan

Volume (Tập) là một nhóm các chapter thuộc về một novel. Một novel có thể có nhiều volumes, và mỗi volume chứa nhiều chapters. Volume API cung cấp:
- Full CRUD operations (Create, Read, Update, Delete)
- Publish/Unpublish với history tracking
- Display order management
- Tự động generate slug từ title
- Kiểm tra volume number uniqueness

**Hierarchy**: Novel → Volume → Chapter

---

### 5.1. Lấy danh sách Volumes theo Novel

**Endpoint**: `GET /api/v1/novels/:novel_id/volumes`

**Path Parameters**:
- `novel_id` (string, required): UUID của novel

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| published_only | bool | No | false | Chỉ lấy volumes đã publish |

**Response**:
```json
{
  "success": true,
  "message": "volume.list_success",
  "data": [
    {
      "id": "uuid-string",
      "novel_id": "novel-uuid",
      "volume_number": 1,
      "title": "Tập 1: Khởi Đầu",
      "slug": "tap-1-khoi-dau",
      "cover_image_url": "https://example.com/volume1.jpg",
      "chapter_count": 250,
      "word_count": 500000,
      "display_order": 1,
      "is_published": true,
      "published_at": "2024-01-01T00:00:00Z",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-02T00:00:00Z"
    }
  ]
}
```

**Curl Examples**:
```bash
# Lấy tất cả volumes của novel
curl -X GET "http://localhost:8080/api/v1/novels/{novel_id}/volumes"

# Chỉ lấy volumes đã publish
curl -X GET "http://localhost:8080/api/v1/novels/{novel_id}/volumes?published_only=true"
```

---

### 5.2. Lấy chi tiết Volume

**Endpoint**: `GET /api/v1/volumes/:id`

**Path Parameters**:
- `id` (string, required): UUID của volume

**Response**:
```json
{
  "success": true,
  "message": "volume.get_success",
  "data": {
    "id": "uuid-string",
    "novel_id": "novel-uuid",
    "volume_number": 1,
    "title": "Tập 1: Khởi Đầu",
    "slug": "tap-1-khoi-dau",
    "description": "Mô tả chi tiết tập 1...",
    "cover_image_url": "https://example.com/volume1.jpg",
    "chapter_count": 250,
    "word_count": 500000,
    "display_order": 1,
    "is_published": true,
    "published_at": "2024-01-01T00:00:00Z",
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-02T00:00:00Z"
  }
}
```

**Curl Example**:
```bash
curl -X GET "http://localhost:8080/api/v1/volumes/{volume_id}"
```

**Error Responses**:
- `400 INVALID_ID`: UUID không hợp lệ
- `404 VOLUME_NOT_FOUND`: Volume không tồn tại

---

### 5.3. Tạo Volume mới

**Endpoint**: `POST /api/v1/volumes`

**Headers**:
- `Content-Type: application/json`
- `Authorization: Bearer <token>` (required)

**Request Body**:
```json
{
  "novel_id": "novel-uuid",
  "volume_number": 1,
  "title": "Tập 1: Khởi Đầu",
  "description": "Mô tả chi tiết tập 1...",
  "cover_image_url": "https://example.com/volume1.jpg",
  "display_order": 1,
  "is_published": false
}
```

**Validation**:
- `novel_id`: bắt buộc, phải là UUID hợp lệ
- `volume_number`: bắt buộc, phải >= 1, unique trong novel
- `title`: bắt buộc, độ dài 1-500 ký tự
- `description`: tùy chọn, max 5000 ký tự
- `cover_image_url`: tùy chọn, phải là URL hợp lệ
- `display_order`: bắt buộc, >= 0
- `is_published`: bắt buộc, boolean

**Response**:
```json
{
  "success": true,
  "message": "volume.created_success",
  "data": {
    "id": "new-uuid",
    "novel_id": "novel-uuid",
    "volume_number": 1,
    "title": "Tập 1: Khởi Đầu",
    "slug": "tap-1-khoi-dau",
    "description": "Mô tả chi tiết tập 1...",
    "cover_image_url": "https://example.com/volume1.jpg",
    "chapter_count": 0,
    "word_count": 0,
    "display_order": 1,
    "is_published": false,
    "published_at": null,
    "created_at": "2024-01-16T00:00:00Z",
    "updated_at": "2024-01-16T00:00:00Z"
  }
}
```

**Curl Example**:
```bash
curl -X POST "http://localhost:8080/api/v1/volumes" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "novel_id": "novel-uuid",
    "volume_number": 1,
    "title": "Tập 1: Khởi Đầu",
    "display_order": 1,
    "is_published": false
  }'
```

**Error Responses**:
- `400 INVALID_INPUT`: Dữ liệu đầu vào không hợp lệ
- `400 INVALID_NOVEL_ID`: Novel ID không hợp lệ
- `409 VOLUME_NUMBER_EXISTS`: Volume number đã tồn tại cho novel này

---

### 5.4. Cập nhật Volume (với History Tracking)

**Endpoint**: `PUT /api/v1/volumes/:id`

**Headers**:
- `Content-Type: application/json`
- `Authorization: Bearer <token>` (required)

**Request Body**:
```json
{
  "volume_number": 1,
  "title": "Tập 1: Khởi Đầu (Updated)",
  "description": "Mô tả mới...",
  "cover_image_url": "https://example.com/new-cover.jpg",
  "display_order": 1,
  "is_published": true
}
```

**History Tracking**:
Mọi thay đổi được log vào bảng `volume_history` với thông tin:
- Version number tự động tăng
- Changed fields (các trường đã thay đổi)
- Change summary (tóm tắt thay đổi)
- Request context (IP, user agent, request ID)
- Changed by (user ID)

**Changed Fields Tracked**:
- `volume_number` - Số thứ tự tập
- `title` - Tiêu đề
- `slug` - Slug (tự động update khi title thay đổi)
- `description` - Mô tả
- `cover_image_url` - URL ảnh bìa
- `display_order` - Thứ tự hiển thị
- `is_published` - Trạng thái publish

**Response**: Giống như response của GET volume detail

**Curl Example**:
```bash
curl -X PUT "http://localhost:8080/api/v1/volumes/{volume_id}" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "volume_number": 1,
    "title": "Tập 1: Khởi Đầu (Updated)",
    "display_order": 1,
    "is_published": true
  }'
```

**Error Responses**:
- `404 VOLUME_NOT_FOUND`: Volume không tồn tại
- `400 INVALID_INPUT`: Dữ liệu không hợp lệ
- `409 VOLUME_NUMBER_EXISTS`: Volume number mới đã tồn tại

---

### 5.5. Xóa Volume

**Endpoint**: `DELETE /api/v1/volumes/:id`

**Headers**:
- `Authorization: Bearer <token>` (required)

**Response**:
```json
{
  "success": true,
  "message": "volume.deleted_success"
}
```

**Curl Example**:
```bash
curl -X DELETE "http://localhost:8080/api/v1/volumes/{volume_id}" \
  -H "Authorization: Bearer <token>"
```

**Error Responses**:
- `404 VOLUME_NOT_FOUND`: Volume không tồn tại

**Note**: Đây là soft delete, dữ liệu vẫn được giữ trong database với `deleted_at` timestamp

---

### 5.6. Cập nhật Display Order

**Endpoint**: `PUT /api/v1/volumes/:id/display-order`

**Headers**:
- `Content-Type: application/json`
- `Authorization: Bearer <token>` (required)

**Request Body**:
```json
{
  "display_order": 5
}
```

**Response**:
```json
{
  "success": true,
  "message": "volume.display_order_updated"
}
```

**Curl Example**:
```bash
curl -X PUT "http://localhost:8080/api/v1/volumes/{volume_id}/display-order" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"display_order": 5}'
```

**Use Case**: Sắp xếp lại thứ tự hiển thị của các volumes

---

### 5.7. Publish Volume (với History Tracking)

**Endpoint**: `POST /api/v1/volumes/:id/publish`

**Headers**:
- `Authorization: Bearer <token>` (required)

**Description**: Publish volume và log vào history

**Response**:
```json
{
  "success": true,
  "message": "volume.published_success"
}
```

**Curl Example**:
```bash
curl -X POST "http://localhost:8080/api/v1/volumes/{volume_id}/publish" \
  -H "Authorization: Bearer <token>"
```

**Note**:
- Tự động set `is_published = true`
- Set `published_at` timestamp nếu chưa có
- Log vào `volume_history` với action = 'published'

---

### 5.8. Unpublish Volume (với History Tracking)

**Endpoint**: `POST /api/v1/volumes/:id/unpublish`

**Headers**:
- `Authorization: Bearer <token>` (required)

**Description**: Ẩn volume (unpublish) và log vào history

**Response**:
```json
{
  "success": true,
  "message": "volume.unpublished_success"
}
```

**Curl Example**:
```bash
curl -X POST "http://localhost:8080/api/v1/volumes/{volume_id}/unpublish" \
  -H "Authorization: Bearer <token>"
```

**Note**:
- Set `is_published = false`
- Log vào `volume_history` với action = 'unpublished'

---

## 6. Chapter API (Chương)

### Tổng quan

Chapter (Chương) là đơn vị nội dung nhỏ nhất trong hệ thống. Mỗi chapter thuộc về một novel và có thể thuộc về một volume (hoặc không). Chapter API cung cấp:
- Full CRUD operations với content management
- Publish immediately hoặc schedule for later
- Content và author notes lưu dưới dạng JSONB
- Tự động tính word count và character count
- Hỗ trợ pricing (free/paid)
- View tracking và statistics
- Status management (draft, published, scheduled)
- History tracking cho mọi update

**Status Flow**: Draft → Published hoặc Draft → Scheduled → Published

---

### 6.1. Lấy danh sách Chapters theo Novel

**Endpoint**: `GET /api/v1/novels/:novel_id/chapters`

**Path Parameters**:
- `novel_id` (string, required): UUID của novel

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| status | string | No | - | Filter theo status: `draft`, `published`, `scheduled` |
| volume_id | string | No | - | Filter theo volume ID |
| published_only | bool | No | false | Chỉ lấy chapters đã publish |
| is_free | bool | No | - | Filter theo free/paid |
| sort_by | string | No | - | Sort theo: `chapter_number`, `published_at`, `views` |
| sort_order | string | No | asc | Thứ tự: `asc`, `desc` |
| page | int | No | - | Số trang |
| limit | int | No | - | Số items/trang |

**Response**:
```json
{
  "success": true,
  "message": "chapter.list_success",
  "data": [
    {
      "id": "uuid-string",
      "novel_id": "novel-uuid",
      "volume_id": "volume-uuid",
      "chapter_number": 1,
      "title": "Chương 1: Bắt Đầu",
      "slug": "chuong-1-bat-dau",
      "word_count": 2500,
      "is_free": true,
      "price": null,
      "currency": null,
      "status": "published",
      "view_count": 10000,
      "like_count": 500,
      "comment_count": 50,
      "display_order": 1,
      "published_at": "2024-01-01T00:00:00Z",
      "scheduled_at": null,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-02T00:00:00Z"
    }
  ]
}
```

**Curl Examples**:
```bash
# Lấy tất cả chapters của novel
curl -X GET "http://localhost:8080/api/v1/novels/{novel_id}/chapters"

# Chỉ lấy chapters đã publish
curl -X GET "http://localhost:8080/api/v1/novels/{novel_id}/chapters?published_only=true"

# Filter theo volume
curl -X GET "http://localhost:8080/api/v1/novels/{novel_id}/chapters?volume_id={volume_id}"

# Filter theo status
curl -X GET "http://localhost:8080/api/v1/novels/{novel_id}/chapters?status=draft"

# Chỉ lấy chapters free
curl -X GET "http://localhost:8080/api/v1/novels/{novel_id}/chapters?is_free=true"

# Sort theo views
curl -X GET "http://localhost:8080/api/v1/novels/{novel_id}/chapters?sort_by=views&sort_order=desc"

# Với pagination
curl -X GET "http://localhost:8080/api/v1/novels/{novel_id}/chapters?page=1&limit=50"
```

---

### 6.2. Lấy danh sách Chapters theo Volume

**Endpoint**: `GET /api/v1/volumes/:volume_id/chapters`

**Path Parameters**:
- `volume_id` (string, required): UUID của volume

**Query Parameters**:
| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| published_only | bool | No | false | Chỉ lấy chapters đã publish |

**Response**: Giống như GET chapters by novel

**Curl Examples**:
```bash
# Lấy tất cả chapters của volume
curl -X GET "http://localhost:8080/api/v1/volumes/{volume_id}/chapters"

# Chỉ lấy chapters đã publish
curl -X GET "http://localhost:8080/api/v1/volumes/{volume_id}/chapters?published_only=true"
```

---

### 6.3. Lấy chi tiết Chapter

**Endpoint**: `GET /api/v1/chapters/:id`

**Path Parameters**:
- `id` (string, required): UUID của chapter

**Response**:
```json
{
  "success": true,
  "message": "chapter.get_success",
  "data": {
    "id": "uuid-string",
    "novel_id": "novel-uuid",
    "volume_id": "volume-uuid",
    "chapter_number": 1,
    "title": "Chương 1: Bắt Đầu",
    "slug": "chuong-1-bat-dau",
    "content": "Nội dung chapter đầy đủ ở đây...",
    "word_count": 2500,
    "character_count": 10000,
    "is_free": true,
    "price": null,
    "currency": null,
    "status": "published",
    "view_count": 10000,
    "like_count": 500,
    "comment_count": 50,
    "display_order": 1,
    "author_notes": "Ghi chú của tác giả...",
    "published_at": "2024-01-01T00:00:00Z",
    "scheduled_at": null,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-02T00:00:00Z"
  }
}
```

**Curl Example**:
```bash
curl -X GET "http://localhost:8080/api/v1/chapters/{chapter_id}"
```

**Error Responses**:
- `400 INVALID_ID`: UUID không hợp lệ
- `404 CHAPTER_NOT_FOUND`: Chapter không tồn tại

---

### 6.4. Tạo Chapter mới

**Endpoint**: `POST /api/v1/chapters`

**Headers**:
- `Content-Type: application/json`
- `Authorization: Bearer <token>` (required)

**Request Body**:
```json
{
  "novel_id": "novel-uuid",
  "volume_id": "volume-uuid",
  "chapter_number": 1,
  "title": "Chương 1: Bắt Đầu",
  "content": "Nội dung chapter đầy đủ...",
  "author_notes": "Ghi chú của tác giả...",
  "is_free": true,
  "price": null,
  "currency": null,
  "status": "draft",
  "display_order": 1,
  "scheduled_at": null
}
```

**Validation**:
- `novel_id`: bắt buộc, phải là UUID hợp lệ
- `volume_id`: tùy chọn, phải là UUID hợp lệ (có thể null)
- `chapter_number`: bắt buộc, phải >= 1, unique trong novel
- `title`: bắt buộc, độ dài 1-500 ký tự
- `content`: bắt buộc, không giới hạn độ dài
- `author_notes`: tùy chọn
- `is_free`: bắt buộc, boolean
- `price`: tùy chọn, >= 0 (bắt buộc nếu `is_free = false`)
- `currency`: tùy chọn, phải là mã ISO 4217 (3 ký tự: USD, VND...)
- `status`: bắt buộc, phải là: `draft`, `published`, `scheduled`
- `display_order`: bắt buộc, >= 0
- `scheduled_at`: tùy chọn, ISO 8601 format (bắt buộc nếu `status = scheduled`)

**Chapter Status**:
- `draft`: Bản nháp (chưa publish)
- `published`: Đã publish
- `scheduled`: Đã lên lịch publish

**Pricing**:
- Nếu `is_free = true`: Chapter miễn phí, không cần `price` và `currency`
- Nếu `is_free = false`: Chapter trả phí, phải có `price` và `currency`

**Auto-calculated Fields**:
- `slug`: Tự động generate từ title
- `word_count`: Tự động đếm từ content
- `character_count`: Tự động đếm ký tự từ content

**Response**: Giống như response của GET chapter detail

**Curl Example**:
```bash
curl -X POST "http://localhost:8080/api/v1/chapters" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "novel_id": "novel-uuid",
    "volume_id": "volume-uuid",
    "chapter_number": 1,
    "title": "Chương 1: Bắt Đầu",
    "content": "Nội dung chapter...",
    "is_free": true,
    "status": "draft",
    "display_order": 1
  }'
```

**Error Responses**:
- `400 INVALID_INPUT`: Dữ liệu đầu vào không hợp lệ
- `400 INVALID_NOVEL_ID`: Novel ID không hợp lệ
- `400 INVALID_VOLUME_ID`: Volume ID không hợp lệ
- `400 INVALID_STATUS`: Status không hợp lệ
- `409 CHAPTER_NUMBER_EXISTS`: Chapter number đã tồn tại cho novel này

---

### 6.5. Cập nhật Chapter (với History Tracking)

**Endpoint**: `PUT /api/v1/chapters/:id`

**Headers**:
- `Content-Type: application/json`
- `Authorization: Bearer <token>` (required)

**Request Body**:
```json
{
  "volume_id": "volume-uuid",
  "chapter_number": 1,
  "title": "Chương 1: Bắt Đầu (Updated)",
  "content": "Nội dung mới...",
  "author_notes": "Ghi chú mới...",
  "is_free": false,
  "price": 5000,
  "currency": "VND",
  "status": "published",
  "display_order": 1,
  "scheduled_at": null
}
```

**History Tracking**:
Mọi thay đổi được log vào bảng `chapter_history` với thông tin:
- Version number tự động tăng
- Changed fields (các trường đã thay đổi)
- Change summary (tóm tắt thay đổi)
- Content changed flag (có thay đổi nội dung hay không)
- Request context (IP, user agent, request ID)
- Changed by (user ID)

**Changed Fields Tracked**:
- `chapter_number` - Số thứ tự chương
- `title` - Tiêu đề
- `slug` - Slug (tự động update khi title thay đổi)
- `content` - Nội dung (chỉ log flag, không lưu full content)
- `is_free` - Trạng thái miễn phí/trả phí
- `price` - Giá
- `currency` - Tiền tệ
- `status` - Trạng thái
- `display_order` - Thứ tự hiển thị
- `author_notes` - Ghi chú tác giả

**Note về Content History**:
- Content KHÔNG được lưu full vào history table (để tiết kiệm dung lượng)
- Chỉ log `content_changed = true/false` flag
- Nếu cần version control cho content, implement riêng content versioning system

**Response**: Giống như response của GET chapter detail

**Curl Example**:
```bash
curl -X PUT "http://localhost:8080/api/v1/chapters/{chapter_id}" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "chapter_number": 1,
    "title": "Chương 1: Bắt Đầu (Updated)",
    "content": "Nội dung mới...",
    "is_free": true,
    "status": "published",
    "display_order": 1
  }'
```

**Error Responses**:
- `404 CHAPTER_NOT_FOUND`: Chapter không tồn tại
- `400 INVALID_INPUT`: Dữ liệu không hợp lệ
- `400 INVALID_STATUS`: Status không hợp lệ
- `409 CHAPTER_NUMBER_EXISTS`: Chapter number mới đã tồn tại

---

### 6.6. Xóa Chapter

**Endpoint**: `DELETE /api/v1/chapters/:id`

**Headers**:
- `Authorization: Bearer <token>` (required)

**Response**:
```json
{
  "success": true,
  "message": "chapter.deleted_success"
}
```

**Curl Example**:
```bash
curl -X DELETE "http://localhost:8080/api/v1/chapters/{chapter_id}" \
  -H "Authorization: Bearer <token>"
```

**Error Responses**:
- `404 CHAPTER_NOT_FOUND`: Chapter không tồn tại

**Note**: Đây là soft delete, dữ liệu vẫn được giữ trong database với `deleted_at` timestamp

---

### 6.7. Publish Chapter ngay lập tức (với History Tracking)

**Endpoint**: `POST /api/v1/chapters/:id/publish`

**Headers**:
- `Authorization: Bearer <token>` (required)

**Description**: Publish chapter ngay lập tức và log vào history

**Response**:
```json
{
  "success": true,
  "message": "chapter.published_success"
}
```

**Curl Example**:
```bash
curl -X POST "http://localhost:8080/api/v1/chapters/{chapter_id}/publish" \
  -H "Authorization: Bearer <token>"
```

**Note**:
- Tự động set `status = 'published'`
- Set `published_at` timestamp nếu chưa có
- Clear `scheduled_at` nếu có
- Log vào `chapter_history` với action = 'published'

---

### 6.8. Schedule Chapter để publish sau

**Endpoint**: `POST /api/v1/chapters/:id/schedule`

**Headers**:
- `Content-Type: application/json`
- `Authorization: Bearer <token>` (required)

**Request Body**:
```json
{
  "scheduled_at": "2024-12-31T10:00:00Z"
}
```

**Validation**:
- `scheduled_at`: bắt buộc, phải là ISO 8601 format

**Response**:
```json
{
  "success": true,
  "message": "chapter.scheduled_success"
}
```

**Curl Example**:
```bash
curl -X POST "http://localhost:8080/api/v1/chapters/{chapter_id}/schedule" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "scheduled_at": "2024-12-31T10:00:00Z"
  }'
```

**Note**:
- Set `status = 'scheduled'`
- Set `scheduled_at` timestamp
- Cần có background job để auto-publish chapters khi đến scheduled time

**Use Case**: Lên lịch trước để tự động publish chapter vào thời gian cụ thể

---

### 6.9. Tăng View Count

**Endpoint**: `POST /api/v1/chapters/:id/view`

**Description**: Track lượt xem chapter. Gọi endpoint này mỗi khi user đọc chapter.

**Response**:
```json
{
  "success": true,
  "message": "chapter.view_incremented"
}
```

**Curl Example**:
```bash
curl -X POST "http://localhost:8080/api/v1/chapters/{chapter_id}/view"
```

**Use Case**:
```javascript
// Gọi khi user mở trang đọc chapter
async function trackChapterView(chapterId) {
  // Chỉ track 1 lần mỗi session
  const viewedKey = `chapter_viewed_${chapterId}`;
  if (!sessionStorage.getItem(viewedKey)) {
    await fetch(`/api/v1/chapters/${chapterId}/view`, { method: 'POST' });
    sessionStorage.setItem(viewedKey, 'true');
  }
}
```

---

### 6.10. Cập nhật Statistics

**Endpoint**: `PUT /api/v1/chapters/:id/statistics`

**Headers**:
- `Content-Type: application/json`
- `Authorization: Bearer <token>` (required)

**Request Body**:
```json
{
  "view_count": 10000,
  "like_count": 500,
  "comment_count": 50
}
```

**Note**: Tất cả các fields đều optional. Chỉ update các fields được cung cấp.

**Response**:
```json
{
  "success": true,
  "message": "chapter.statistics_updated"
}
```

**Curl Example**:
```bash
curl -X PUT "http://localhost:8080/api/v1/chapters/{chapter_id}/statistics" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "like_count": 501,
    "comment_count": 51
  }'
```

**Use Case**: Batch update statistics từ external services (likes, comments system)

---

## 7. Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| VALIDATION_FAILED | 400 | Dữ liệu đầu vào không hợp lệ |
| INVALID_INPUT | 400 | Input không đúng format |
| INVALID_ID | 400 | UUID không hợp lệ |
| UNAUTHORIZED | 401 | Chưa đăng nhập |
| NOT_FOUND | 404 | Resource không tồn tại |
| SLUG_EXISTS | 409 | Slug đã tồn tại |
| IN_USE | 409 | Resource đang được sử dụng |
| CIRCULAR_REFERENCE | 400 | Tham chiếu vòng (chỉ Genre) |
| HAS_CHILDREN | 409 | Genre có các genre con (chỉ Genre) |
| PARENT_NOT_FOUND | 400 | Parent genre không tồn tại (chỉ Genre) |
| VOLUME_NOT_FOUND | 404 | Volume không tồn tại |
| VOLUME_NUMBER_EXISTS | 409 | Volume number đã tồn tại cho novel này |
| CHAPTER_NOT_FOUND | 404 | Chapter không tồn tại |
| CHAPTER_NUMBER_EXISTS | 409 | Chapter number đã tồn tại cho novel này |
| INVALID_STATUS | 400 | Status không hợp lệ |
| INVALID_SCHEDULED_TIME | 400 | Thời gian lên lịch không hợp lệ |

---

## 8. Pagination và Sorting

### Pagination

Tất cả các list endpoint đều hỗ trợ pagination:

```bash
# Lấy trang 2, mỗi trang 50 items
curl -X GET "http://localhost:8080/api/v1/genres?page=2&limit=50"
```

Response sẽ chứa `meta` object:
```json
{
  "meta": {
    "page": 2,
    "limit": 50,
    "total_items": 150,
    "total_pages": 3
  }
}
```

### Sorting

Sử dụng `sort_by` và `sort_order`:

```bash
# Sắp xếp genres theo tên tăng dần
curl -X GET "http://localhost:8080/api/v1/genres?sort_by=name&sort_order=asc"

# Sắp xếp authors theo số lượng novel giảm dần
curl -X GET "http://localhost:8080/api/v1/authors?sort_by=novels&sort_order=desc"

# Sắp xếp novels theo rating giảm dần
curl -X GET "http://localhost:8080/api/v1/novels?sort_by=rating&sort_order=desc"
```

### Search

Tìm kiếm theo tên (case-insensitive):

```bash
# Tìm genres có tên chứa "fantasy"
curl -X GET "http://localhost:8080/api/v1/genres?search=fantasy"

# Tìm authors có tên chứa "nhĩ căn"
curl -X GET "http://localhost:8080/api/v1/authors?search=nhĩ+căn"

# Tìm novels trong title và synopsis
curl -X GET "http://localhost:8080/api/v1/novels?search=tiên+nghịch"
```

---

## 9. Filter Options

### Genre Filters
- `active_only`: Chỉ lấy genres đang active

### Author Filters
- `is_verified`: Lọc theo trạng thái verified

### Artist Filters
- `specialization`: Lọc theo chuyên môn (cover_artist, illustrator, etc.)
- `is_verified`: Lọc theo trạng thái verified

### Novel Filters
- `status`: Lọc theo trạng thái (draft, ongoing, completed, hiatus, dropped)
- `original_language`: Lọc theo ngôn ngữ gốc (ISO 639-1: vi, en, zh, ja, ko...)

**Example**:
```bash
# Lấy novels đang ongoing, ngôn ngữ gốc tiếng Trung
curl -X GET "http://localhost:8080/api/v1/novels?status=ongoing&original_language=zh"

# Lấy artists là cover_artist và đã verified
curl -X GET "http://localhost:8080/api/v1/artists?specialization=cover_artist&is_verified=true"
```

---

## 10. Best Practices

### 10.1. Error Handling

Luôn kiểm tra `success` field trong response:

```javascript
const response = await fetch('http://localhost:8080/api/v1/novels');
const data = await response.json();

if (!data.success) {
  console.error('Error:', data.error.code, data.error.message);
  // Handle error
} else {
  console.log('Data:', data.data);
}
```

### 10.2. Pagination

Sử dụng `meta` để implement pagination UI:

```javascript
const { page, limit, total_items, total_pages } = data.meta;

// Hiển thị: "Showing 21-40 of 100 items"
const start = (page - 1) * limit + 1;
const end = Math.min(page * limit, total_items);
console.log(`Showing ${start}-${end} of ${total_items} items`);
```

### 10.3. Search với debounce

Implement debounce khi search để tránh gọi API quá nhiều:

```javascript
let debounceTimer;
function searchNovels(query) {
  clearTimeout(debounceTimer);
  debounceTimer = setTimeout(() => {
    fetch(`http://localhost:8080/api/v1/novels?search=${query}`)
      .then(res => res.json())
      .then(data => {
        // Update UI
      });
  }, 300);
}
```

### 10.4. View Tracking

Track views một cách thông minh:

```javascript
// Chỉ track view một lần mỗi session
let trackedNovels = new Set();

function trackNovelView(novelId) {
  if (!trackedNovels.has(novelId)) {
    fetch(`http://localhost:8080/api/v1/novels/${novelId}/view`, {
      method: 'POST'
    });
    trackedNovels.add(novelId);
  }
}
```

### 10.5. Combine Filters

Kết hợp nhiều filters để lọc chính xác:

```bash
# Tìm novels ongoing, tiếng Trung, sort theo rating, trang 2
curl -X GET "http://localhost:8080/api/v1/novels?status=ongoing&original_language=zh&sort_by=rating&sort_order=desc&page=2&limit=20"
```

---

## 11. JavaScript/TypeScript Examples

### 11.1. Fetch Novels

```typescript
interface Novel {
  id: string;
  title: string;
  slug: string;
  cover_image_url?: string;
  status: string;
  total_chapters: number;
  view_count: number;
  rating_average: number;
  created_at: string;
}

interface PaginationMeta {
  page: number;
  limit: number;
  total_items: number;
  total_pages: number;
}

interface StandardResponse<T> {
  success: boolean;
  message: string;
  data: T;
  meta?: PaginationMeta;
}

async function fetchNovels(
  page: number = 1,
  limit: number = 20,
  search?: string,
  status?: string
): Promise<StandardResponse<Novel[]>> {
  const params = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString(),
  });

  if (search) {
    params.append('search', search);
  }

  if (status) {
    params.append('status', status);
  }

  const response = await fetch(`http://localhost:8080/api/v1/novels?${params}`);
  return response.json();
}

// Usage
const result = await fetchNovels(1, 20, undefined, 'ongoing');
if (result.success) {
  console.log('Novels:', result.data);
  console.log('Meta:', result.meta);
}
```

### 11.2. Create Novel

```typescript
interface CreateNovelRequest {
  title: string;
  synopsis?: string;
  cover_image_url?: string;
  thumbnail_url?: string;
  status: 'draft' | 'ongoing' | 'completed' | 'hiatus' | 'dropped';
  original_language?: string;
  original_title?: string;
  metadata?: string;
}

interface NovelDetail {
  id: string;
  title: string;
  slug: string;
  synopsis: string;
  status: string;
  total_chapters: number;
  view_count: number;
  // ... other fields
}

async function createNovel(data: CreateNovelRequest): Promise<StandardResponse<NovelDetail>> {
  const response = await fetch('http://localhost:8080/api/v1/novels', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });

  return response.json();
}

// Usage
const newNovel = await createNovel({
  title: 'Tiên Nghịch',
  synopsis: 'Thuận thiên giả, sống...',
  status: 'draft',
  original_language: 'zh',
  original_title: '仙逆',
});

if (newNovel.success) {
  console.log('Created novel:', newNovel.data);
}
```

### 11.3. Update Novel Status

```typescript
async function updateNovelStatus(
  id: string,
  status: 'draft' | 'ongoing' | 'completed' | 'hiatus' | 'dropped'
): Promise<StandardResponse<NovelDetail>> {
  // First, get current novel data
  const currentResponse = await fetch(`http://localhost:8080/api/v1/novels/${id}`);
  const current = await currentResponse.json();

  if (!current.success) {
    return current;
  }

  // Update with new status
  const response = await fetch(`http://localhost:8080/api/v1/novels/${id}`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      ...current.data,
      status: status,
    }),
  });

  return response.json();
}

// Usage
const updated = await updateNovelStatus('550e8400-e29b-41d4-a716-446655440000', 'completed');
```

### 11.4. Novel Reader Component Example

```typescript
class NovelReader {
  private novelId: string;
  private viewTracked: boolean = false;

  constructor(novelId: string) {
    this.novelId = novelId;
  }

  async loadNovel(): Promise<NovelDetail | null> {
    const response = await fetch(`http://localhost:8080/api/v1/novels/${this.novelId}`);
    const data = await response.json();

    if (data.success) {
      // Track view on first load
      if (!this.viewTracked) {
        this.trackView();
        this.viewTracked = true;
      }
      return data.data;
    }

    return null;
  }

  private async trackView(): Promise<void> {
    try {
      await fetch(`http://localhost:8080/api/v1/novels/${this.novelId}/view`, {
        method: 'POST',
      });
    } catch (error) {
      console.error('Failed to track view:', error);
    }
  }
}

// Usage
const reader = new NovelReader('550e8400-e29b-41d4-a716-446655440000');
const novel = await reader.loadNovel();
```

---

## 12. Testing với Postman

### Import Collection

1. Tạo một collection mới trong Postman
2. Thêm các requests sau:

**Novel Collection**:
- GET List Novels
- GET Novel Detail (by ID)
- GET Novel Detail (by slug)
- POST Create Novel
- PUT Update Novel
- DELETE Delete Novel
- POST Increment View Count

**Genre Collection**:
- GET List Genres
- GET Genre Detail
- POST Create Genre
- PUT Update Genre
- DELETE Delete Genre

**Author Collection**:
- GET List Authors
- GET Author Detail
- POST Create Author
- PUT Update Author
- DELETE Delete Author

**Artist Collection**:
- GET List Artists
- GET Artist Detail
- POST Create Artist
- PUT Update Artist
- DELETE Delete Artist

### Environment Variables

Tạo environment với variables:
- `base_url`: `http://localhost:8080/api/v1`
- `novel_id`: UUID của novel test
- `genre_id`: UUID của genre test
- `author_id`: UUID của author test
- `artist_id`: UUID của artist test

### Test Scripts Example

```javascript
// Test cho Create Novel
pm.test("Status code is 201", function () {
    pm.response.to.have.status(201);
});

pm.test("Response has novel data", function () {
    var jsonData = pm.response.json();
    pm.expect(jsonData.success).to.eql(true);
    pm.expect(jsonData.data).to.have.property('id');
    pm.expect(jsonData.data).to.have.property('slug');

    // Save novel_id for later use
    pm.environment.set("novel_id", jsonData.data.id);
});
```

---

## 13. Rate Limiting & Performance

### Recommendations

1. **Pagination**: Không request quá 100 items một lúc
2. **Search**: Sử dụng debounce 300ms khi implement search
3. **Caching**: Cache responses ở client side khi có thể
4. **View Tracking**: Chỉ track view một lần mỗi session/user
5. **Batch requests**: Gom nhiều requests lại nếu có thể

---

## 14. Support & Contact

Nếu gặp vấn đề, vui lòng:
1. Kiểm tra error response để biết chi tiết lỗi
2. Xem lại validation requirements
3. Contact support team với error code và request details
