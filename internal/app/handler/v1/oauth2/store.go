package oauth2

import (
	"context"
	"fmt"
)

// User là một struct đơn giản đại diện cho người dùng.
// Trong một ứng dụng thực tế, nó sẽ nằm trong package domain.
type User struct {
	ID      string
	Name    string
	Email   string
	Picture string
}

// Store là interface định nghĩa các phương thức truy cập dữ liệu cần thiết cho handler.
type Store interface {
	GetUserByID(ctx context.Context, id string) (*User, error)
}

// MockStore là một triển khai giả (mock) của Store để phục vụ cho việc phát triển.
type MockStore struct{}

// NewMockStore tạo một instance mới của MockStore.
func NewMockStore() *MockStore {
	return &MockStore{}
}

// GetUserByID trả về một người dùng giả với ID được cung cấp.
func (ms *MockStore) GetUserByID(ctx context.Context, id string) (*User, error) {
	if id == "1" { // Chỉ trả về user nếu ID là "1"
		return &User{
			ID:      "1",
			Name:    "John Doe",
			Email:   "john.doe@example.com",
			Picture: "https://example.com/avatar.jpg",
		}, nil
	}
	return nil, fmt.Errorf("user with id %s not found", id)
}
