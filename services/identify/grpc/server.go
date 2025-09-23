// Package grpc provides gRPC server setup for the Identity Service
package grpc

import (
	"fmt"
	"log"

	"wibusystem/pkg/grpc/config"
	grpcserver "wibusystem/pkg/grpc/server"
	"wibusystem/services/identify/oauth2"
	"wibusystem/services/identify/services/interfaces"
)

// SetupGRPCServer creates and configures a gRPC server with both token validation and user services
func SetupGRPCServer(provider *oauth2.Provider, userService interfaces.UserServiceInterface, cfg *config.ServerConfig) (*grpcserver.Server, error) {
	// Create fosite token validator
	validator := NewFositeTokenValidator(provider)

	// Create gRPC server
	server, err := grpcserver.NewServer(cfg, validator)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC server: %w", err)
	}

	// Register UserService on the same gRPC server
	RegisterUserServiceOnExistingServer(server.GetGRPCServer(), userService)

	log.Printf("gRPC server configured with TokenValidation and UserService on %s", server.GetAddress())
	return server, nil
}

// StartGRPCServer starts the gRPC server in a goroutine
func StartGRPCServer(server *grpcserver.Server) {
	go func() {
		log.Printf("Starting gRPC server on %s", server.GetAddress())
		if err := server.Start(); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()
}