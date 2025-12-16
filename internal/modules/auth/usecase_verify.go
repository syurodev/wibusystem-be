package auth

import (
	"context"
)

type verifyEmailUseCase struct {
	authService AuthService
}

func NewVerifyEmailUseCase(authService AuthService) VerifyEmailUseCase {
	return &verifyEmailUseCase{
		authService: authService,
	}
}

func (uc *verifyEmailUseCase) Execute(ctx context.Context, input VerifyEmailInput) error {
	return uc.authService.VerifyEmail(ctx, input.Token)
}
