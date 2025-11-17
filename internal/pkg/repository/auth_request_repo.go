package repository

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
)

const (
	authRequestKeyPrefix = "auth_request:%s"
	authUserIDKeyPrefix  = "auth_request_user:%s"
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
		return fmt.Errorf("failed to marshal auth request: %w", err)
	}

	return r.client.Set(ctx, key, data, ttl)
}

// GetAuthRequest lấy authorization request từ Redis
func (r *authRequestRepository) GetAuthRequest(ctx context.Context, requestID string) (fosite.AuthorizeRequester, error) {
	key := fmt.Sprintf(authRequestKeyPrefix, requestID)

	data, err := r.client.Get(ctx, key)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, fmt.Errorf("auth request not found or expired")
		}
		return nil, err
	}

	// Deserialize từ JSON
	var stored storedAuthorizeRequest
	if err := json.Unmarshal([]byte(data), &stored); err != nil {
		return nil, fmt.Errorf("failed to unmarshal auth request: %w", err)
	}

	// Load lại client từ DB
	clientUUID, err := uuid.FromString(stored.ClientID)
	if err != nil {
		return nil, fmt.Errorf("invalid client id in stored request")
	}

	domainClient, err := r.clientRepo.GetClientByID(ctx, clientUUID)
	if err != nil {
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
