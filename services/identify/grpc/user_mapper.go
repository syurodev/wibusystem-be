package grpc

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	m "wibusystem/pkg/common/model"
	pb "wibusystem/pkg/grpc/userservice"
)

// ModelUserToProtoUser converts a model.User to proto User message
func ModelUserToProtoUser(user *m.User) *pb.User {
	if user == nil {
		return nil
	}

	protoUser := &pb.User{
		Id:          user.ID.String(),
		DisplayName: user.DisplayName,
		Username:    user.Username,
		Email:       user.Email,
		CreatedAt:   timestamppb.New(user.CreatedAt),
	}

	// Handle avatar_url (pointer field)
	if user.AvatarURL != nil {
		protoUser.AvatarUrl = *user.AvatarURL
	}

	return protoUser
}

// ModelUsersToProtoUsers converts a slice of model.User to proto User messages
func ModelUsersToProtoUsers(users []*m.User) []*pb.User {
	if users == nil {
		return nil
	}

	protoUsers := make([]*pb.User, 0, len(users))
	for _, user := range users {
		if protoUser := ModelUserToProtoUser(user); protoUser != nil {
			protoUsers = append(protoUsers, protoUser)
		}
	}

	return protoUsers
}