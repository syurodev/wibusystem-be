package auth

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pb "wibusystem/pkg/grpc/tokenvalidation"
)

// GRPCClient wraps the gRPC connection and client for token validation
type GRPCClient struct {
	client pb.TokenValidationServiceClient
	conn   *grpc.ClientConn
	config *Config
	mutex  sync.RWMutex
}

// NewGRPCClient creates a new gRPC client for token validation
func NewGRPCClient(config *Config) (*GRPCClient, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Create gRPC connection
	conn, err := grpc.NewClient(
		config.IdentifyGRPCURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithTimeout(config.GRPCTimeout),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to identify service at %s: %w", config.IdentifyGRPCURL, err)
	}

	client := pb.NewTokenValidationServiceClient(conn)

	return &GRPCClient{
		client: client,
		conn:   conn,
		config: config,
	}, nil
}

// ValidateToken validates a token using the identify service
func (g *GRPCClient) ValidateToken(ctx context.Context, token string, requiredScopes ...string) (*ValidationResult, error) {
	g.mutex.RLock()
	client := g.client
	g.mutex.RUnlock()

	if client == nil {
		return &ValidationResult{
			Valid: false,
			Error: "gRPC client not initialized",
		}, nil
	}

	// Create validation request
	req := &pb.ValidateTokenRequest{
		Token:          token,
		TokenType:      "access_token",
		RequiredScopes: requiredScopes,
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, g.config.GRPCTimeout)
	defer cancel()

	// Call identify service
	resp, err := client.ValidateToken(ctx, req)
	if err != nil {
		// Handle gRPC errors
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.Unavailable:
				return &ValidationResult{
					Valid: false,
					Error: "identify service unavailable",
				}, nil
			case codes.DeadlineExceeded:
				return &ValidationResult{
					Valid: false,
					Error: "token validation timeout",
				}, nil
			default:
				return &ValidationResult{
					Valid: false,
					Error: fmt.Sprintf("validation error: %s", st.Message()),
				}, nil
			}
		}

		return &ValidationResult{
			Valid: false,
			Error: fmt.Sprintf("failed to validate token: %v", err),
		}, nil
	}

	// Handle validation response
	if !resp.Valid {
		return &ValidationResult{
			Valid: false,
			Error: resp.Error,
		}, nil
	}

	// Convert response to UserContext
	userContext, err := g.convertResponseToUserContext(resp)
	if err != nil {
		return &ValidationResult{
			Valid: false,
			Error: fmt.Sprintf("failed to parse user context: %v", err),
		}, nil
	}

	return &ValidationResult{
		Valid:       true,
		UserContext: userContext,
	}, nil
}

// convertResponseToUserContext converts gRPC response to UserContext
func (g *GRPCClient) convertResponseToUserContext(resp *pb.ValidateTokenResponse) (*UserContext, error) {
	userContext := &UserContext{}

	// Extract token info
	if resp.TokenInfo != nil {
		userContext.Scopes = resp.TokenInfo.Scope
		userContext.ClientID = resp.TokenInfo.ClientId
		userContext.TokenType = resp.TokenInfo.TokenType

		if resp.TokenInfo.ExpiresAt > 0 {
			userContext.ExpiresAt = time.Unix(resp.TokenInfo.ExpiresAt, 0)
		}
		if resp.TokenInfo.IssuedAt > 0 {
			userContext.IssuedAt = time.Unix(resp.TokenInfo.IssuedAt, 0)
		}
	}

	// Extract user info
	if resp.UserInfo != nil {
		// Parse user ID
		if resp.UserInfo.Subject != "" {
			if userID, err := uuid.Parse(resp.UserInfo.Subject); err == nil {
				userContext.UserID = userID
			}
		}

		userContext.Username = resp.UserInfo.Username
		userContext.Email = resp.UserInfo.Email
		userContext.Name = resp.UserInfo.Name
		userContext.EmailVerified = resp.UserInfo.EmailVerified
		userContext.Extra = resp.UserInfo.Extra
	}

	return userContext, nil
}

// Close closes the gRPC connection
func (g *GRPCClient) Close() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if g.conn != nil {
		err := g.conn.Close()
		g.conn = nil
		g.client = nil
		return err
	}
	return nil
}

// Reconnect attempts to reconnect to the identify service
func (g *GRPCClient) Reconnect() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Close existing connection
	if g.conn != nil {
		g.conn.Close()
	}

	// Create new connection
	conn, err := grpc.NewClient(
		g.config.IdentifyGRPCURL,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithTimeout(g.config.GRPCTimeout),
	)
	if err != nil {
		return fmt.Errorf("failed to reconnect to identify service: %w", err)
	}

	g.conn = conn
	g.client = pb.NewTokenValidationServiceClient(conn)

	return nil
}

// HealthCheck performs a simple health check by validating an empty token
// This is mainly used to test connectivity to the identify service
func (g *GRPCClient) HealthCheck(ctx context.Context) error {
	g.mutex.RLock()
	client := g.client
	g.mutex.RUnlock()

	if client == nil {
		return fmt.Errorf("gRPC client not initialized")
	}

	ctx, cancel := context.WithTimeout(ctx, g.config.GRPCTimeout)
	defer cancel()

	// Use empty token for health check - should return invalid but no connection error
	req := &pb.ValidateTokenRequest{
		Token:     "",
		TokenType: "access_token",
	}

	_, err := client.ValidateToken(ctx, req)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			// If we get a proper gRPC response (even if validation fails), connection is OK
			if st.Code() == codes.InvalidArgument {
				return nil
			}
		}
		return fmt.Errorf("identify service health check failed: %w", err)
	}

	return nil
}