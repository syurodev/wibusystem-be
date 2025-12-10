package auth

import (
	"context"

	"github.com/gofrs/uuid/v5"

	"system/internal/domain"
)

// AuthService interface định nghĩa business logic cho Auth module
type AuthService interface {
	RegisterUser(ctx context.Context, email, password, fullName string) (*domain.User, string, error)
	CreateEmailVerificationToken(ctx context.Context, userID uuid.UUID) (string, error)
	VerifyEmail(ctx context.Context, tokenStr string) error
	CreatePasswordResetToken(ctx context.Context, email string) (string, error)
	ResetPassword(ctx context.Context, tokenStr, newPassword string) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)
}
