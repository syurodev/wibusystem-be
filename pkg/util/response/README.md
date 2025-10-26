```go
import "system/pkg/util/response" // Giả sử đường dẫn file là pkg/util/response/response.go

// Sử dụng hàm đã được dịch và đặt tên ngắn gọn
response.Success(c, http.StatusOK, "user_created_success", userRes, nil)

// Giả sử lỗi validation trả về data là gin.H{"field": "email"}
response.Error(c, http.StatusBadRequest, "E4001", "validation.required", gin.H{"field": "email"})
```