package auth

import (
	"context"
)

type resetPasswordUseCase struct {
	authService AuthService
}

func NewResetPasswordUseCase(authService AuthService) ResetPasswordUseCase {
	return &resetPasswordUseCase{
		authService: authService,
	}
}

func (uc *resetPasswordUseCase) Execute(ctx context.Context, input ResetPasswordInput) error {
	return uc.authService.ResetPassword(ctx, input.Token, input.NewPassword)
}
