package grpc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "wibusystem/pkg/grpc/userservice"
	"wibusystem/services/identify/services/interfaces"
)

// UserServiceHandler implements the gRPC UserService server
type UserServiceHandler struct {
	pb.UnimplementedUserServiceServer
	userService interfaces.UserServiceInterface
}

// NewUserServiceHandler creates a new UserServiceHandler
func NewUserServiceHandler(userService interfaces.UserServiceInterface) *UserServiceHandler {
	return &UserServiceHandler{
		userService: userService,
	}
}

// GetUser implements the GetUser RPC method
func (h *UserServiceHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	// Validate request
	if req.UserId == "" {
		return &pb.GetUserResponse{
			Result: &pb.GetUserResponse_Error{
				Error: "user_id is required",
			},
		}, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Parse UUID
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return &pb.GetUserResponse{
			Result: &pb.GetUserResponse_Error{
				Error: "invalid user_id format",
			},
		}, status.Error(codes.InvalidArgument, "invalid user_id format")
	}

	// Get user from service
	user, err := h.userService.GetUserByID(ctx, userID)
	if err != nil {
		// Check if it's a "not found" error
		if isNotFoundError(err) {
			return &pb.GetUserResponse{
				Result: &pb.GetUserResponse_Error{
					Error: "user not found",
				},
			}, status.Error(codes.NotFound, "user not found")
		}

		// Internal server error for other cases
		return &pb.GetUserResponse{
			Result: &pb.GetUserResponse_Error{
				Error: "internal server error",
			},
		}, status.Error(codes.Internal, "failed to retrieve user")
	}

	// Convert to proto and return success response
	protoUser := ModelUserToProtoUser(user)
	return &pb.GetUserResponse{
		Result: &pb.GetUserResponse_User{
			User: protoUser,
		},
	}, nil
}

// GetUsers implements the GetUsers RPC method
func (h *UserServiceHandler) GetUsers(ctx context.Context, req *pb.GetUsersRequest) (*pb.GetUsersResponse, error) {
	// Validate request
	if len(req.UserIds) == 0 {
		return &pb.GetUsersResponse{
			Error: "at least one user_id is required",
		}, status.Error(codes.InvalidArgument, "at least one user_id is required")
	}

	// Parse all UUIDs first
	userIDs := make([]uuid.UUID, 0, len(req.UserIds))
	for _, userIDStr := range req.UserIds {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return &pb.GetUsersResponse{
				Error: fmt.Sprintf("invalid user_id format: %s", userIDStr),
			}, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid user_id format: %s", userIDStr))
		}
		userIDs = append(userIDs, userID)
	}

	// Collect users and track missing ones
	var users []*pb.User
	var missingIDs []string

	for i, userID := range userIDs {
		user, err := h.userService.GetUserByID(ctx, userID)
		if err != nil {
			if isNotFoundError(err) {
				// Track missing user ID
				missingIDs = append(missingIDs, req.UserIds[i])
				continue
			}

			// Return internal error for other failures
			return &pb.GetUsersResponse{
				Error: "internal server error",
			}, status.Error(codes.Internal, "failed to retrieve users")
		}

		// Convert and add to results
		if protoUser := ModelUserToProtoUser(user); protoUser != nil {
			users = append(users, protoUser)
		}
	}

	// Return response with found users and missing IDs
	return &pb.GetUsersResponse{
		Users:          users,
		MissingUserIds: missingIDs,
	}, nil
}

// isNotFoundError checks if the error indicates a user was not found
// This is a simple string-based check - in a real implementation,
// you might want to use custom error types
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	// Check common "not found" error patterns
	errStr := err.Error()
	return contains(errStr, "not found") ||
		   contains(errStr, "no rows") ||
		   contains(errStr, "record not found")
}

// contains checks if a string contains a substring (case-insensitive helper)
func contains(str, substr string) bool {
	return len(str) >= len(substr) &&
		   (str == substr ||
		    (len(str) > len(substr) &&
		     someContains(str, substr)))
}

// someContains is a simple substring check helper
func someContains(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}