package grpc

import (
	"log"

	"google.golang.org/grpc"

	pb "wibusystem/pkg/grpc/userservice"
	"wibusystem/services/identify/services/interfaces"
)

// RegisterUserServiceOnExistingServer registers UserService on an existing gRPC server
// This is used to add UserService to the same server as TokenValidationService
func RegisterUserServiceOnExistingServer(server *grpc.Server, userService interfaces.UserServiceInterface) {
	handler := NewUserServiceHandler(userService)
	pb.RegisterUserServiceServer(server, handler)
	log.Printf("UserService registered on existing gRPC server")
}