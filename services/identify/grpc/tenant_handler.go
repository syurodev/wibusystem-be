package grpc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "wibusystem/pkg/grpc/tenantservice"
	"wibusystem/services/identify/services/interfaces"
)

// TenantServiceHandler implements the gRPC TenantService server
type TenantServiceHandler struct {
	pb.UnimplementedTenantServiceServer
	tenantService interfaces.TenantServiceInterface
}

// NewTenantServiceHandler creates a new TenantServiceHandler
func NewTenantServiceHandler(tenantService interfaces.TenantServiceInterface) *TenantServiceHandler {
	return &TenantServiceHandler{
		tenantService: tenantService,
	}
}

// GetTenant implements the GetTenant RPC method
func (h *TenantServiceHandler) GetTenant(ctx context.Context, req *pb.GetTenantRequest) (*pb.GetTenantResponse, error) {
	// Validate request
	if req.TenantId == "" {
		return &pb.GetTenantResponse{
			Result: &pb.GetTenantResponse_Error{
				Error: "tenant_id is required",
			},
		}, nil
	}

	// Parse tenant ID
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return &pb.GetTenantResponse{
			Result: &pb.GetTenantResponse_Error{
				Error: "invalid tenant_id format",
			},
		}, nil
	}

	// Get tenant from service
	tenant, err := h.tenantService.GetTenantByID(ctx, tenantID)
	if err != nil {
		return &pb.GetTenantResponse{
			Result: &pb.GetTenantResponse_Error{
				Error: fmt.Sprintf("tenant not found: %v", err),
			},
		}, nil
	}

	// Convert to protobuf message
	pbTenant := &pb.Tenant{
		Id:          tenant.ID.String(),
		Name:        tenant.Name,
		Slug:        tenant.Slug,
		Description: "",
		Status:      tenant.Status,
		CreatedAt:   timestamppb.New(tenant.CreatedAt),
		UpdatedAt:   timestamppb.New(tenant.UpdatedAt),
	}

	if tenant.Description != nil {
		pbTenant.Description = *tenant.Description
	}

	return &pb.GetTenantResponse{
		Result: &pb.GetTenantResponse_Tenant{
			Tenant: pbTenant,
		},
	}, nil
}

// GetTenants implements the GetTenants RPC method
func (h *TenantServiceHandler) GetTenants(ctx context.Context, req *pb.GetTenantsRequest) (*pb.GetTenantsResponse, error) {
	if len(req.TenantIds) == 0 {
		return &pb.GetTenantsResponse{
			Error: "at least one tenant_id is required",
		}, nil
	}

	var tenants []*pb.Tenant
	var missingTenantIDs []string

	for _, tenantIDStr := range req.TenantIds {
		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			missingTenantIDs = append(missingTenantIDs, tenantIDStr)
			continue
		}

		tenant, err := h.tenantService.GetTenantByID(ctx, tenantID)
		if err != nil {
			missingTenantIDs = append(missingTenantIDs, tenantIDStr)
			continue
		}

		pbTenant := &pb.Tenant{
			Id:          tenant.ID.String(),
			Name:        tenant.Name,
			Slug:        tenant.Slug,
			Description: "",
			Status:      tenant.Status,
			CreatedAt:   timestamppb.New(tenant.CreatedAt),
			UpdatedAt:   timestamppb.New(tenant.UpdatedAt),
		}

		if tenant.Description != nil {
			pbTenant.Description = *tenant.Description
		}

		tenants = append(tenants, pbTenant)
	}

	return &pb.GetTenantsResponse{
		Tenants:           tenants,
		MissingTenantIds:  missingTenantIDs,
	}, nil
}

// GetTenantBySlug implements the GetTenantBySlug RPC method
func (h *TenantServiceHandler) GetTenantBySlug(ctx context.Context, req *pb.GetTenantBySlugRequest) (*pb.GetTenantResponse, error) {
	// Validate request
	if req.Slug == "" {
		return &pb.GetTenantResponse{
			Result: &pb.GetTenantResponse_Error{
				Error: "slug is required",
			},
		}, nil
	}

	// Get tenant from service
	tenant, err := h.tenantService.GetTenantBySlug(ctx, req.Slug)
	if err != nil {
		return &pb.GetTenantResponse{
			Result: &pb.GetTenantResponse_Error{
				Error: fmt.Sprintf("tenant not found: %v", err),
			},
		}, nil
	}

	// Convert to protobuf message
	pbTenant := &pb.Tenant{
		Id:          tenant.ID.String(),
		Name:        tenant.Name,
		Slug:        tenant.Slug,
		Description: "",
		Status:      tenant.Status,
		CreatedAt:   timestamppb.New(tenant.CreatedAt),
		UpdatedAt:   timestamppb.New(tenant.UpdatedAt),
	}

	if tenant.Description != nil {
		pbTenant.Description = *tenant.Description
	}

	return &pb.GetTenantResponse{
		Result: &pb.GetTenantResponse_Tenant{
			Tenant: pbTenant,
		},
	}, nil
}