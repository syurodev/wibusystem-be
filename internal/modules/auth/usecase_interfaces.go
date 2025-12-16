package auth

import (
	"context"

	"system/internal/domain"
)

type RegisterUserInput struct {
	Email    string
	Password string
	FullName string
}

type RegisterUserResult struct {
	User              *domain.User
	VerificationToken string
}

type RegisterUserUseCase interface {
	Execute(ctx context.Context, input RegisterUserInput) (*RegisterUserResult, error)
}

type VerifyEmailInput struct {
	Token string
}

type VerifyEmailUseCase interface {
	Execute(ctx context.Context, input VerifyEmailInput) error
}

type ForgotPasswordInput struct {
	Email string
}

type ForgotPasswordUseCase interface {
	Execute(ctx context.Context, input ForgotPasswordInput) (string, error)
}

type ResetPasswordInput struct {
	Token       string
	NewPassword string
}

type ResetPasswordUseCase interface {
	Execute(ctx context.Context, input ResetPasswordInput) error
}
