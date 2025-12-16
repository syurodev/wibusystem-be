package auth

import (
	"context"
)

type forgotPasswordUseCase struct {
	authService AuthService
}

func NewForgotPasswordUseCase(authService AuthService) ForgotPasswordUseCase {
	return &forgotPasswordUseCase{
		authService: authService,
	}
}

func (uc *forgotPasswordUseCase) Execute(ctx context.Context, input ForgotPasswordInput) (string, error) {
	return uc.authService.CreatePasswordResetToken(ctx, input.Email)
}
