# API Quick Reference

## Base URL
```
http://localhost:8080/api/v1
```

## Endpoints Summary

### Genre (Thể loại)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/genres` | Lấy danh sách genres |
| GET | `/genres/:id` | Lấy chi tiết genre |
| POST | `/genres` | Tạo genre mới |
| PUT | `/genres/:id` | Cập nhật genre |
| DELETE | `/genres/:id` | Xóa genre |

### Author (Tác giả)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/authors` | Lấy danh sách authors |
| GET | `/authors/:id` | Lấy chi tiết author |
| POST | `/authors` | Tạo author mới |
| PUT | `/authors/:id` | Cập nhật author |
| DELETE | `/authors/:id` | Xóa author |

### Artist (Hoạ sĩ)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/artists` | Lấy danh sách artists |
| GET | `/artists/:id` | Lấy chi tiết artist |
| POST | `/artists` | Tạo artist mới |
| PUT | `/artists/:id` | Cập nhật artist |
| DELETE | `/artists/:id` | Xóa artist |

### Novel (Tiểu thuyết)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/novels` | Lấy danh sách novels |
| GET | `/novels/:id` | Lấy chi tiết novel (UUID hoặc slug) |
| POST | `/novels` | Tạo novel mới |
| PUT | `/novels/:id` | Cập nhật novel |
| DELETE | `/novels/:id` | Xóa novel |
| POST | `/novels/:id/view` | Tăng view count |

## Common Query Parameters

### List Endpoints
- `page` (int): Số trang, default = 1
- `limit` (int): Số items/trang, default = 20, max = 100
- `search` (string): Tìm kiếm theo tên
- `sort_by` (string): Trường để sắp xếp
- `sort_order` (string): `asc` hoặc `desc`

### Genre Sort Options
- `name`, `views`, `series`, `created`, `updated`

### Author Sort Options
- `name`, `views`, `novels`, `created`

### Artist Sort Options
- `name`, `novels`, `created`

### Novel Sort Options
- `created_at`, `rating`, `views`, `last_chapter`

## Novel Status Values
- `draft` - Bản nháp
- `ongoing` - Đang cập nhật
- `completed` - Đã hoàn thành
- `hiatus` - Tạm ngưng
- `dropped` - Đã drop

## Quick Examples

### Lấy danh sách với pagination
```bash
curl -X GET "http://localhost:8080/api/v1/genres?page=1&limit=20"
curl -X GET "http://localhost:8080/api/v1/novels?page=1&limit=20"
```

### Tìm kiếm
```bash
# Tìm authors
curl -X GET "http://localhost:8080/api/v1/authors?search=nhĩ+căn"

# Tìm novels
curl -X GET "http://localhost:8080/api/v1/novels?search=tiên+nghịch"
```

### Sắp xếp
```bash
# Sort genres theo views
curl -X GET "http://localhost:8080/api/v1/genres?sort_by=views&sort_order=desc"

# Sort novels theo rating
curl -X GET "http://localhost:8080/api/v1/novels?sort_by=rating&sort_order=desc"
```

### Filter
```bash
# Novels đang ongoing
curl -X GET "http://localhost:8080/api/v1/novels?status=ongoing"

# Novels tiếng Trung
curl -X GET "http://localhost:8080/api/v1/novels?original_language=zh"

# Artists là cover_artist
curl -X GET "http://localhost:8080/api/v1/artists?specialization=cover_artist"
```

### Tạo mới
```bash
# Tạo genre
curl -X POST "http://localhost:8080/api/v1/genres" \
  -H "Content-Type: application/json" \
  -d '{"name": "Fantasy", "description": "Mô tả"}'

# Tạo novel
curl -X POST "http://localhost:8080/api/v1/novels" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Tiên Nghịch",
    "synopsis": "Thuận thiên giả, sống...",
    "status": "draft",
    "original_language": "zh"
  }'
```

### Cập nhật
```bash
# Update genre
curl -X PUT "http://localhost:8080/api/v1/genres/{id}" \
  -H "Content-Type: application/json" \
  -d '{"name": "New Name", "is_active": true}'

# Update novel status
curl -X PUT "http://localhost:8080/api/v1/novels/{id}" \
  -H "Content-Type: application/json" \
  -d '{"title": "Title", "synopsis": "...", "status": "ongoing"}'
```

### Xóa
```bash
curl -X DELETE "http://localhost:8080/api/v1/genres/{id}"
curl -X DELETE "http://localhost:8080/api/v1/novels/{id}"
```

### Track Novel View
```bash
curl -X POST "http://localhost:8080/api/v1/novels/{id}/view"
```

## Response Format

### Success Response
```json
{
  "success": true,
  "message": "operation.success",
  "data": { ... },
  "meta": {
    "page": 1,
    "limit": 20,
    "total_items": 100,
    "total_pages": 5
  }
}
```

### Error Response
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "error.message",
    "details": "Chi tiết lỗi"
  }
}
```

## Common Error Codes
- `400 VALIDATION_FAILED`: Dữ liệu không hợp lệ
- `400 INVALID_INPUT`: Input không đúng format
- `400 INVALID_ID`: UUID không hợp lệ
- `401 UNAUTHORIZED`: Chưa đăng nhập
- `404 NOT_FOUND`: Không tìm thấy
- `409 SLUG_EXISTS`: Slug đã tồn tại
- `409 IN_USE`: Resource đang được sử dụng

## Filter Quick Reference

### Genre
- `active_only=true` - Chỉ lấy genres active

### Author & Artist
- `is_verified=true` - Chỉ lấy đã verified

### Artist
- `specialization=cover_artist` - Lọc theo chuyên môn

### Novel
- `status=ongoing` - Lọc theo status
- `original_language=zh` - Lọc theo ngôn ngữ gốc (vi, en, zh, ja, ko...)

## Complex Query Examples

```bash
# Novels ongoing, tiếng Trung, sort theo rating
curl -X GET "http://localhost:8080/api/v1/novels?status=ongoing&original_language=zh&sort_by=rating&sort_order=desc"

# Authors verified, sort theo số novels
curl -X GET "http://localhost:8080/api/v1/authors?is_verified=true&sort_by=novels&sort_order=desc"

# Tìm kiếm novels + filter + sort
curl -X GET "http://localhost:8080/api/v1/novels?search=tiên&status=ongoing&sort_by=views&page=1"
```

---

Xem [API_CLIENT_GUIDE.md](./API_CLIENT_GUIDE.md) để biết chi tiết đầy đủ.
