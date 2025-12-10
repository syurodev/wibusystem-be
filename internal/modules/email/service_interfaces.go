package email

import (
	"context"
)

// EmailService defines business email operations
type EmailService interface {
	SendVerificationEmail(ctx context.Context, toEmail, toName, token string) error
	SendPasswordResetEmail(ctx context.Context, toEmail, toName, token string) error
	SendWelcomeEmail(ctx context.Context, toEmail, toName string) error
}
