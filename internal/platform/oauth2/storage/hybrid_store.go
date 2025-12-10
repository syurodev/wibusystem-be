package storage

import (
	"context"
	"errors"

	"github.com/ory/fosite"
)

// HybridStore kết hợp các triển khai lưu trữ khác nhau (SQL, Redis) thành một store duy nhất
// cho Fosite. Nó sử dụng tính năng embedding của Go để tự động kế thừa các phương thức.
// Các phương thức cho RefreshTokenStorage được implement tường minh để thực hiện chiến lược cache-aside.
type HybridStore struct {
	*SQLStore
	*RedisStore
}

// NewHybridStore tạo một instance mới của HybridStore.
func NewHybridStore(sqlStore *SQLStore, redisStore *RedisStore) *HybridStore {
	return &HybridStore{
		SQLStore:   sqlStore,
		RedisStore: redisStore,
	}
}

// --- RefreshTokenStorage Orchestration (Cache-Aside) --- //

// CreateRefreshTokenSession lưu session vào SQL trước, sau đó vào Redis.
func (s *HybridStore) CreateRefreshTokenSession(ctx context.Context, signature string, accessSignature string, requester fosite.Requester) error {
	// 1. Ghi vào SQL làm nguồn tin cậy (source of truth)
	if err := s.SQLStore.CreateRefreshTokenSession(ctx, signature, accessSignature, requester); err != nil {
		return err
	}
	// 2. Ghi vào Redis cache. Nếu lỗi, có thể bỏ qua vì dữ liệu đã an toàn trong SQL.
	_ = s.RedisStore.CreateRefreshTokenSession(ctx, signature, accessSignature, requester)
	return nil
}

// GetRefreshTokenSession thử đọc từ Redis trước, nếu không có thì đọc từ SQL và làm mới cache.
func (s *HybridStore) GetRefreshTokenSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	// 1. Thử đọc từ Redis cache trước
	req, err := s.RedisStore.GetRefreshTokenSession(ctx, signature, session)
	if err == nil {
		// Cache hit, trả về ngay
		return req, nil
	}

	// Nếu lỗi không phải là 'not found', trả về lỗi ngay
	if !errors.Is(err, fosite.ErrNotFound) {
		return nil, err
	}

	// 2. Cache miss, đọc từ SQL
	req, err = s.SQLStore.GetRefreshTokenSession(ctx, signature, session)
	if err != nil {
		return nil, err
	}

	// 3. Làm mới (re-hydrate) cache bằng dữ liệu từ SQL.
	_ = s.RedisStore.CreateRefreshTokenSession(ctx, signature, "", req)

	return req, nil
}

// DeleteRefreshTokenSession xóa session ở cả SQL và Redis.
func (s *HybridStore) DeleteRefreshTokenSession(ctx context.Context, signature string) error {
	// Xóa ở cả hai nơi để đảm bảo tính nhất quán
	_ = s.RedisStore.DeleteRefreshTokenSession(ctx, signature)
	return s.SQLStore.DeleteRefreshTokenSession(ctx, signature)
}

// RotateRefreshToken xoá session ở cả SQL và Redis.
func (s *HybridStore) RotateRefreshToken(ctx context.Context, requestID string, signature string) error {
	_ = s.RedisStore.DeleteRefreshTokenSession(ctx, signature)
	return s.SQLStore.RotateRefreshToken(ctx, requestID, signature)
}
