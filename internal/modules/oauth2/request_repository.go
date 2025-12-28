package oauth2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"system/internal/domain"
	"system/internal/platform/database"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/ory/fosite"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	authRequestKeyPrefix    = "auth_request:%s" // Stores original query string
	authUserIDKeyPrefix     = "auth_request_user:%s"
	authRequestParamsPrefix = "auth_request_params:%s" // Stores original OAuth2 params
	authPasskeyPromptPrefix = "auth_passkey_prompt:%s" // Stores passkey prompt flag
)

// storedAuthorizeRequest là dạng đã được "strip" client để có thể lưu/khôi phục.
type storedAuthorizeRequest struct {
	Request  *fosite.AuthorizeRequest `json:"request"`
	ClientID string                   `json:"client_id"`
}

// authRequestRepository triển khai AuthRequestRepository sử dụng Redis
type authRequestRepository struct {
	client     *database.RedisClient
	clientRepo domain.OAuth2ClientRepository
}

// NewAuthRequestRepository tạo một instance mới của authRequestRepository
func NewAuthRequestRepository(client *database.RedisClient, clientRepo domain.OAuth2ClientRepository) domain.AuthRequestRepository {
	return &authRequestRepository{client: client, clientRepo: clientRepo}
}

// SaveAuthRequest lưu authorization request vào Redis
func (r *authRequestRepository) SaveAuthRequest(ctx context.Context, requestID string, ar fosite.AuthorizeRequester, ttl time.Duration) error {
	key := fmt.Sprintf(authRequestKeyPrefix, requestID)

	authorizeReq, ok := ar.(*fosite.AuthorizeRequest)
	if !ok {
		return fmt.Errorf("unexpected authorize request type")
	}

	// Sao chép và bỏ client để (de)serialize được, client sẽ được load lại từ DB
	sanitized := *authorizeReq
	sanitized.Client = nil
	sanitized.Session = nil

	payload := storedAuthorizeRequest{
		Request:  &sanitized,
		ClientID: authorizeReq.GetClient().GetID(),
	}

	// Serialize authorization request to JSON
	data, err := json.Marshal(payload)
	if err != nil {
		zap.L().Error("Failed to marshal auth request",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		return fmt.Errorf("failed to marshal auth request: %w", err)
	}

	if err := r.client.Set(ctx, key, data, ttl); err != nil {
		zap.L().Error("Failed to save auth request to Redis",
			zap.String("key", key),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// GetAuthRequest lấy authorization request từ Redis
func (r *authRequestRepository) GetAuthRequest(ctx context.Context, requestID string) (fosite.AuthorizeRequester, error) {
	key := fmt.Sprintf(authRequestKeyPrefix, requestID)

	data, err := r.client.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			zap.L().Warn("Auth request not found in Redis",
				zap.String("key", key),
				zap.String("request_id", requestID),
			)
			return nil, fmt.Errorf("auth request not found or expired")
		}
		zap.L().Error("Redis error fetching auth request",
			zap.String("key", key),
			zap.Error(err),
		)
		return nil, err
	}

	// Deserialize từ JSON
	var stored storedAuthorizeRequest
	if err := json.Unmarshal([]byte(data), &stored); err != nil {
		zap.L().Error("Failed to unmarshal auth request",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to unmarshal auth request: %w", err)
	}

	// Load lại client từ DB
	clientUUID, err := uuid.FromString(stored.ClientID)
	if err != nil {
		zap.L().Error("Invalid client ID in stored request",
			zap.String("client_id", stored.ClientID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("invalid client id in stored request")
	}

	domainClient, err := r.clientRepo.GetClientByID(ctx, clientUUID)
	if err != nil {
		zap.L().Error("Failed to load client from database",
			zap.String("client_id", stored.ClientID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to load client: %w", err)
	}

	// Map sang fosite.DefaultClient
	fositeClient := &fosite.DefaultClient{
		ID:            domainClient.ID.String(),
		Secret:        []byte(domainClient.SecretHash),
		RedirectURIs:  domainClient.RedirectURIs,
		GrantTypes:    domainClient.GrantTypes,
		ResponseTypes: domainClient.ResponseTypes,
		Scopes:        domainClient.Scopes,
		Public:        domainClient.IsPublic,
	}

	stored.Request.Client = fositeClient

	return stored.Request, nil
}

// DeleteAuthRequest xóa authorization request khỏi Redis
func (r *authRequestRepository) DeleteAuthRequest(ctx context.Context, requestID string) error {
	key := fmt.Sprintf(authRequestKeyPrefix, requestID)
	userKey := fmt.Sprintf(authUserIDKeyPrefix, requestID)

	// Delete both keys
	if err := r.client.Del(ctx, key); err != nil {
		return err
	}
	return r.client.Del(ctx, userKey)
}

// SaveAuthRequestWithUserID lưu authorization request cùng với userID
func (r *authRequestRepository) SaveAuthRequestWithUserID(ctx context.Context, requestID string, ar fosite.AuthorizeRequester, userID string, ttl time.Duration) error {
	// Save authorization request
	if err := r.SaveAuthRequest(ctx, requestID, ar, ttl); err != nil {
		return err
	}

	// Save userID separately for quick lookup
	userKey := fmt.Sprintf(authUserIDKeyPrefix, requestID)
	return r.client.Set(ctx, userKey, userID, ttl)
}

// GetUserIDFromAuthRequest lấy userID từ stored authorization request
func (r *authRequestRepository) GetUserIDFromAuthRequest(ctx context.Context, requestID string) (string, error) {
	userKey := fmt.Sprintf(authUserIDKeyPrefix, requestID)

	userID, err := r.client.Get(ctx, userKey)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", fmt.Errorf("user id not found for auth request")
		}
		return "", err
	}

	return userID, nil
}

// SaveQueryParams lưu original OAuth2 query params string
func (r *authRequestRepository) SaveQueryParams(ctx context.Context, requestID string, queryParams string, ttl time.Duration) error {
	key := fmt.Sprintf(authRequestParamsPrefix, requestID)
	return r.client.Set(ctx, key, queryParams, ttl)
}

// GetQueryParams lấy original OAuth2 query params string
func (r *authRequestRepository) GetQueryParams(ctx context.Context, requestID string) (string, error) {
	key := fmt.Sprintf(authRequestParamsPrefix, requestID)

	params, err := r.client.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", fmt.Errorf("query params not found or expired")
		}
		return "", err
	}

	return params, nil
}

// SavePasskeyPromptFlag lưu flag hiển thị passkey prompt
func (r *authRequestRepository) SavePasskeyPromptFlag(ctx context.Context, requestID string, show bool, ttl time.Duration) error {
	key := fmt.Sprintf(authPasskeyPromptPrefix, requestID)
	value := "0"
	if show {
		value = "1"
	}
	return r.client.Set(ctx, key, value, ttl)
}

// GetPasskeyPromptFlag lấy flag hiển thị passkey prompt
func (r *authRequestRepository) GetPasskeyPromptFlag(ctx context.Context, requestID string) (bool, error) {
	key := fmt.Sprintf(authPasskeyPromptPrefix, requestID)

	value, err := r.client.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil // Default to false if not found
		}
		return false, err
	}

	return value == "1", nil
}
