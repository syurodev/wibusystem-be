package auth

import (
	"context"
)

type registerUserUseCase struct {
	authService AuthService
}

func NewRegisterUserUseCase(authService AuthService) RegisterUserUseCase {
	return &registerUserUseCase{
		authService: authService,
	}
}

func (uc *registerUserUseCase) Execute(ctx context.Context, input RegisterUserInput) (*RegisterUserResult, error) {
	user, token, err := uc.authService.RegisterUser(ctx, input.Email, input.Password, input.FullName)
	if err != nil {
		return nil, err
	}
	return &RegisterUserResult{
		User:              user,
		VerificationToken: token,
	}, nil
}
