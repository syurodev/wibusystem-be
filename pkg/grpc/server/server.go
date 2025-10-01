package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	"wibusystem/pkg/common/oauth"
	"wibusystem/pkg/grpc/config"
	pb "wibusystem/pkg/grpc/tokenvalidation"
)

// TokenValidator represents a component capable of validating OAuth tokens and
// returning a shared ValidationResult structure.
type TokenValidator interface {
	ValidateToken(ctx context.Context, req *oauth.ValidationRequest) (*oauth.ValidationResult, error)
}

// Server wraps a gRPC server instance that exposes a token validation API.
type Server struct {
	cfg        *config.ServerConfig
	validator  TokenValidator
	grpcServer *grpc.Server
	listener   net.Listener
	mu         sync.Mutex
	started    bool
}

// NewServer constructs a gRPC server using the provided configuration and
// validator implementation.
func NewServer(cfg *config.ServerConfig, validator TokenValidator) (*Server, error) {
	if validator == nil {
		return nil, errors.New("token validator cannot be nil")
	}
	if cfg == nil {
		return nil, errors.New("grpc server config cannot be nil")
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid grpc server config: %w", err)
	}

	effectiveCfg := cfg.WithDefaults()

	keepaliveParams := keepalive.ServerParameters{
		MaxConnectionIdle:     effectiveCfg.MaxConnectionIdle,
		MaxConnectionAge:      effectiveCfg.MaxConnectionAge,
		MaxConnectionAgeGrace: effectiveCfg.MaxConnectionAgeGrace,
		Time:                  effectiveCfg.Time,
		Timeout:               effectiveCfg.Timeout,
	}

	opts := []grpc.ServerOption{
		grpc.ConnectionTimeout(effectiveCfg.ConnectionTimeout),
		grpc.MaxRecvMsgSize(effectiveCfg.MaxRecvMsgSize),
		grpc.MaxSendMsgSize(effectiveCfg.MaxSendMsgSize),
		grpc.KeepaliveParams(keepaliveParams),
	}

	grpcServer := grpc.NewServer(opts...)
	service := &tokenValidationService{validator: validator}
	pb.RegisterTokenValidationServiceServer(grpcServer, service)

	if effectiveCfg.EnableReflection {
		reflection.Register(grpcServer)
	}

	return &Server{
		cfg:        effectiveCfg,
		validator:  validator,
		grpcServer: grpcServer,
	}, nil
}

// Start begins listening on the configured host/port and blocks until the
// server is stopped or an unrecoverable error occurs.
func (s *Server) Start() error {
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return errors.New("grpc server already started")
	}

	address := s.cfg.Address()
	listener, err := net.Listen("tcp", address)
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("failed to listen on %s: %w", address, err)
	}

	s.listener = listener
	s.started = true
	s.mu.Unlock()

	return s.grpcServer.Serve(listener)
}

// Stop gracefully stops the gRPC server, closing the listener if required.
func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return
	}

	done := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		s.grpcServer.Stop()
	}

	if s.listener != nil {
		_ = s.listener.Close()
		s.listener = nil
	}

	s.started = false
}

// GetAddress returns the configured listen address in host:port form.
func (s *Server) GetAddress() string {
	if s == nil || s.cfg == nil {
		return ""
	}
	return s.cfg.Address()
}

// GetGRPCServer returns the underlying gRPC server instance for registering additional services
func (s *Server) GetGRPCServer() *grpc.Server {
	if s == nil {
		return nil
	}
	return s.grpcServer
}

// tokenValidationService adapts a TokenValidator to the gRPC transport using proto types.
type tokenValidationService struct {
	pb.UnimplementedTokenValidationServiceServer
	validator TokenValidator
}

func (s *tokenValidationService) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	// Convert proto request to internal validation request
	validationReq := &oauth.ValidationRequest{
		Token:     req.Token,
		TokenType: req.TokenType,
		Scopes:    req.RequiredScopes,
	}

	// Validate token using internal validator
	result, err := s.validator.ValidateToken(ctx, validationReq)
	if err != nil {
		return &pb.ValidateTokenResponse{
			Valid: false,
			Error: err.Error(),
		}, nil
	}

	if result == nil {
		return &pb.ValidateTokenResponse{
			Valid: false,
			Error: "validator returned nil result",
		}, nil
	}

	// Convert internal result to proto response
	response := &pb.ValidateTokenResponse{
		Valid: result.Valid,
		Error: result.Error,
	}

	// Convert token info if available
	if result.TokenInfo != nil {
		response.TokenInfo = &pb.TokenInfo{
			Active:    result.TokenInfo.Active,
			TokenType: result.TokenInfo.TokenType,
			Scope:     result.TokenInfo.Scope,
			ClientId:  result.TokenInfo.ClientID,
			Audience:  result.TokenInfo.Audience,
			Issuer:    result.TokenInfo.Issuer,
			Subject:   result.TokenInfo.Subject,
			ExpiresAt: result.TokenInfo.ExpiresAt.Unix(),
			IssuedAt:  result.TokenInfo.IssuedAt.Unix(),
		}
	}

	// Convert user info if available
	if result.UserInfo != nil {
		response.UserInfo = &pb.UserInfo{
			Subject:       result.UserInfo.Subject,
			Username:      result.UserInfo.Username,
			Email:         result.UserInfo.Email,
			Name:          result.UserInfo.Name,
			EmailVerified: result.UserInfo.Verified,
			Extra:         result.UserInfo.Extra,
			TenantId:      result.UserInfo.TenantID, // Add TenantID to gRPC response
		}
		if result.UserInfo.UpdatedAt != nil {
			response.UserInfo.UpdatedAt = result.UserInfo.UpdatedAt.Unix()
		}
	}

	return response, nil
}

