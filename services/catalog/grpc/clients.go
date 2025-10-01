package grpc

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	userpb "wibusystem/pkg/grpc/userservice"
	tenantpb "wibusystem/pkg/grpc/tenantservice"
	d "wibusystem/pkg/common/dto"
)

// ClientManager manages gRPC clients for external services
type ClientManager struct {
	userClient   userpb.UserServiceClient
	tenantClient tenantpb.TenantServiceClient
	userConn     *grpc.ClientConn
	tenantConn   *grpc.ClientConn
}

// NewClientManager creates a new gRPC client manager
func NewClientManager(identifyServiceURL string) (*ClientManager, error) {
	// Create connection options
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithTimeout(10 * time.Second),
	}

	// Connect to identify service for user service
	userConn, err := grpc.NewClient(identifyServiceURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}

	// Connect to identify service for tenant service (same URL)
	tenantConn, err := grpc.NewClient(identifyServiceURL, opts...)
	if err != nil {
		userConn.Close()
		return nil, fmt.Errorf("failed to connect to tenant service: %w", err)
	}

	// Create clients
	userClient := userpb.NewUserServiceClient(userConn)
	tenantClient := tenantpb.NewTenantServiceClient(tenantConn)

	log.Printf("gRPC clients connected to identify service at %s", identifyServiceURL)

	return &ClientManager{
		userClient:   userClient,
		tenantClient: tenantClient,
		userConn:     userConn,
		tenantConn:   tenantConn,
	}, nil
}

// GetUsers retrieves multiple users by their IDs
func (c *ClientManager) GetUsers(ctx context.Context, userIDs []string) (map[string]*d.UserSummary, error) {
	if len(userIDs) == 0 {
		return make(map[string]*d.UserSummary), nil
	}

	// Remove duplicates
	uniqueIDs := make(map[string]bool)
	var filteredIDs []string
	for _, id := range userIDs {
		if id != "" && !uniqueIDs[id] {
			uniqueIDs[id] = true
			filteredIDs = append(filteredIDs, id)
		}
	}

	if len(filteredIDs) == 0 {
		return make(map[string]*d.UserSummary), nil
	}

	req := &userpb.GetUsersRequest{
		UserIds: filteredIDs,
	}

	resp, err := c.userClient.GetUsers(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get users via gRPC: %w", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("user service error: %s", resp.Error)
	}

	// Convert to map
	result := make(map[string]*d.UserSummary)
	for _, user := range resp.Users {
		result[user.Id] = &d.UserSummary{
			ID:          user.Id,
			DisplayName: user.DisplayName,
		}
	}

	return result, nil
}

// GetTenants retrieves multiple tenants by their IDs
func (c *ClientManager) GetTenants(ctx context.Context, tenantIDs []string) (map[string]*d.TenantSummary, error) {
	if len(tenantIDs) == 0 {
		return make(map[string]*d.TenantSummary), nil
	}

	// Remove duplicates
	uniqueIDs := make(map[string]bool)
	var filteredIDs []string
	for _, id := range tenantIDs {
		if id != "" && !uniqueIDs[id] {
			uniqueIDs[id] = true
			filteredIDs = append(filteredIDs, id)
		}
	}

	if len(filteredIDs) == 0 {
		return make(map[string]*d.TenantSummary), nil
	}

	req := &tenantpb.GetTenantsRequest{
		TenantIds: filteredIDs,
	}

	resp, err := c.tenantClient.GetTenants(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenants via gRPC: %w", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("tenant service error: %s", resp.Error)
	}

	// Convert to map
	result := make(map[string]*d.TenantSummary)
	for _, tenant := range resp.Tenants {
		result[tenant.Id] = &d.TenantSummary{
			ID:   tenant.Id,
			Name: tenant.Name,
		}
	}

	return result, nil
}

// Close closes all gRPC connections
func (c *ClientManager) Close() error {
	var errs []error

	if err := c.userConn.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close user connection: %w", err))
	}

	if err := c.tenantConn.Close(); err != nil {
		errs = append(errs, fmt.Errorf("failed to close tenant connection: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}

	log.Printf("gRPC clients closed")
	return nil
}