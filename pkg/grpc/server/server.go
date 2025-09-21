package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	"wibusystem/pkg/common/oauth"
	"wibusystem/pkg/grpc/config"
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
	RegisterTokenValidationServiceServer(grpcServer, service)

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

const (
	serviceName             = "wibusystem.grpc.TokenValidationService"
	fullMethodValidateToken = "/" + serviceName + "/ValidateToken"
)

// TokenValidationServiceServer defines the gRPC surface area exposed to other
// services. We intentionally keep the request/response types dynamic via
// structpb.Struct to avoid requiring generated protobuf code at this stage.
type TokenValidationServiceServer interface {
	ValidateToken(context.Context, *structpb.Struct) (*structpb.Struct, error)
}

// tokenValidationService adapts a TokenValidator to the gRPC transport.
type tokenValidationService struct {
	validator TokenValidator
}

func (s *tokenValidationService) ValidateToken(ctx context.Context, req *structpb.Struct) (*structpb.Struct, error) {
	validationReq, err := structToValidationRequest(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	result, err := s.validator.ValidateToken(ctx, validationReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if result == nil {
		return nil, status.Error(codes.Internal, "validator returned nil result")
	}

	respMap := validationResultToMap(result)
	respStruct, err := structpb.NewStruct(respMap)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to encode response: %v", err))
	}

	return respStruct, nil
}

// RegisterTokenValidationServiceServer registers the service implementation
// with the provided gRPC server registrar.
func RegisterTokenValidationServiceServer(s grpc.ServiceRegistrar, srv TokenValidationServiceServer) {
	s.RegisterService(&tokenValidationServiceDesc, srv)
}

var tokenValidationServiceDesc = grpc.ServiceDesc{
	ServiceName: serviceName,
	HandlerType: (*TokenValidationServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ValidateToken",
			Handler:    _TokenValidationService_ValidateToken_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "token_validation",
}

func _TokenValidationService_ValidateToken_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(structpb.Struct)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TokenValidationServiceServer).ValidateToken(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: fullMethodValidateToken,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TokenValidationServiceServer).ValidateToken(ctx, req.(*structpb.Struct))
	}
	return interceptor(ctx, in, info, handler)
}

// structToValidationRequest converts a dynamic protobuf Struct into our shared
// ValidationRequest type, performing basic input validation.
func structToValidationRequest(st *structpb.Struct) (*oauth.ValidationRequest, error) {
	if st == nil {
		return nil, oauth.ErrTokenRequired
	}

	fields := st.GetFields()
	tokenValue, ok := fields["token"]
	if !ok || tokenValue.GetStringValue() == "" {
		return nil, oauth.ErrTokenRequired
	}

	req := &oauth.ValidationRequest{
		Token: tokenValue.GetStringValue(),
	}

	if tokenTypeValue, ok := fields["token_type"]; ok {
		req.TokenType = tokenTypeValue.GetStringValue()
	}

	if scopesValue, ok := fields["scopes"]; ok {
		list := scopesValue.GetListValue()
		if list != nil {
			for _, v := range list.GetValues() {
				req.Scopes = append(req.Scopes, v.GetStringValue())
			}
		}
	}

	return req, nil
}

// validationResultToMap prepares a ValidationResult for transport by converting
// it into a map compatible with structpb.NewStruct.
func validationResultToMap(result *oauth.ValidationResult) map[string]interface{} {
	output := map[string]interface{}{
		"valid": result.Valid,
	}

	if result.Error != "" {
		output["error"] = result.Error
	}

	if info := result.TokenInfo; info != nil {
		tokenInfo := map[string]interface{}{
			"active": info.Active,
		}
		if info.TokenType != "" {
			tokenInfo["token_type"] = info.TokenType
		}
		if len(info.Scope) > 0 {
			tokenInfo["scope"] = info.Scope
		}
		if info.ClientID != "" {
			tokenInfo["client_id"] = info.ClientID
		}
		if len(info.Audience) > 0 {
			tokenInfo["audience"] = info.Audience
		}
		if info.Issuer != "" {
			tokenInfo["issuer"] = info.Issuer
		}
		if info.Subject != "" {
			tokenInfo["subject"] = info.Subject
		}
		if !info.ExpiresAt.IsZero() {
			tokenInfo["expires_at"] = info.ExpiresAt.Format(time.RFC3339Nano)
		}
		if !info.IssuedAt.IsZero() {
			tokenInfo["issued_at"] = info.IssuedAt.Format(time.RFC3339Nano)
		}

		output["token_info"] = tokenInfo
	}

	if user := result.UserInfo; user != nil {
		userInfo := map[string]interface{}{}
		if user.Subject != "" {
			userInfo["subject"] = user.Subject
		}
		if user.Username != "" {
			userInfo["username"] = user.Username
		}
		if user.Email != "" {
			userInfo["email"] = user.Email
		}
		if user.Name != "" {
			userInfo["name"] = user.Name
		}
		if user.Verified {
			userInfo["verified"] = user.Verified
		}
		if len(user.Extra) > 0 {
			extra := make(map[string]interface{}, len(user.Extra))
			for k, v := range user.Extra {
				extra[k] = v
			}
			userInfo["extra"] = extra
		}
		if user.UpdatedAt != nil && !user.UpdatedAt.IsZero() {
			userInfo["updated_at"] = user.UpdatedAt.Format(time.RFC3339Nano)
		}

		output["user_info"] = userInfo
	}

	return output
}
