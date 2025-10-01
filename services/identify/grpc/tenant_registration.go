package grpc

import (
	"log"

	"google.golang.org/grpc"

	pb "wibusystem/pkg/grpc/tenantservice"
	"wibusystem/services/identify/services/interfaces"
)

// RegisterTenantServiceOnExistingServer registers TenantService on an existing gRPC server
// This is used to add TenantService to the same server as TokenValidationService and UserService
func RegisterTenantServiceOnExistingServer(server *grpc.Server, tenantService interfaces.TenantServiceInterface) {
	handler := NewTenantServiceHandler(tenantService)
	pb.RegisterTenantServiceServer(server, handler)
	log.Printf("TenantService registered on existing gRPC server")
}