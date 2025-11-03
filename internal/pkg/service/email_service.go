package service

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"

	"github.com/resend/resend-go/v2"
	"go.uber.org/zap"

	"system/configs"
)

// EmailService xử lý gửi email qua Resend.
type EmailService struct {
	client *resend.Client
	config *configs.EmailConfig
	logger *zap.Logger
}

// NewEmailService tạo instance mới của EmailService.
func NewEmailService(config *configs.EmailConfig, logger *zap.Logger) *EmailService {
	client := resend.NewClient(config.APIKey)

	return &EmailService{
		client: client,
		config: config,
		logger: logger,
	}
}

// EmailData chứa dữ liệu chung cho email templates.
type EmailData struct {
	ToEmail   string
	ToName    string
	Subject   string
	Variables map[string]interface{}
}

// SendVerificationEmail gửi email xác thực với token.
func (s *EmailService) SendVerificationEmail(ctx context.Context, toEmail, toName, token string) error {
	// Build verification URL
	verificationURL := strings.ReplaceAll(s.config.VerificationURL, "{{.Token}}", token)

	// Debug log
	s.logger.Debug("Sending verification email",
		zap.String("to_email", toEmail),
		zap.String("to_name", toName),
		zap.String("token", token),
		zap.String("verification_url", verificationURL),
	)

	// Prepare template data
	data := map[string]interface{}{
		"Name":            toName,
		"VerificationURL": verificationURL,
		"Token":           token,
	}

	// Render HTML template
	htmlBody, err := s.renderTemplate("verification", data)
	if err != nil {
		s.logger.Error("Failed to render verification email template", zap.Error(err))
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// Send email via Resend
	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail),
		To:      []string{toEmail},
		Subject: "Verify your email address",
		Html:    htmlBody,
	}

	sent, err := s.client.Emails.Send(params)
	if err != nil {
		s.logger.Error("Failed to send verification email",
			zap.String("to_email", toEmail),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send verification email: %w", err)
	}

	s.logger.Info("Verification email sent successfully",
		zap.String("to_email", toEmail),
		zap.String("email_id", sent.Id),
		zap.String("token", token),
		zap.String("verification_url", verificationURL),
	)

	return nil
}

// SendPasswordResetEmail gửi email reset password với token.
func (s *EmailService) SendPasswordResetEmail(ctx context.Context, toEmail, toName, token string) error {
	// Build reset URL
	resetURL := strings.ReplaceAll(s.config.PasswordResetURL, "{{.Token}}", token)

	// Debug log
	s.logger.Debug("Sending password reset email",
		zap.String("to_email", toEmail),
		zap.String("to_name", toName),
		zap.String("token", token),
		zap.String("reset_url", resetURL),
	)

	// Prepare template data
	data := map[string]interface{}{
		"Name":     toName,
		"ResetURL": resetURL,
		"Token":    token,
	}

	// Render HTML template
	htmlBody, err := s.renderTemplate("password_reset", data)
	if err != nil {
		s.logger.Error("Failed to render password reset email template", zap.Error(err))
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// Send email via Resend
	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail),
		To:      []string{toEmail},
		Subject: "Reset your password",
		Html:    htmlBody,
	}

	sent, err := s.client.Emails.Send(params)
	if err != nil {
		s.logger.Error("Failed to send password reset email",
			zap.String("to_email", toEmail),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send password reset email: %w", err)
	}

	s.logger.Info("Password reset email sent successfully",
		zap.String("to_email", toEmail),
		zap.String("email_id", sent.Id),
		zap.String("token", token),
		zap.String("reset_url", resetURL),
	)

	return nil
}

// SendWelcomeEmail gửi email chào mừng sau khi verify thành công.
func (s *EmailService) SendWelcomeEmail(ctx context.Context, toEmail, toName string) error {
	// Prepare template data
	data := map[string]interface{}{
		"Name":    toName,
		"BaseURL": s.config.BaseURL,
	}

	// Render HTML template
	htmlBody, err := s.renderTemplate("welcome", data)
	if err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	// Send email via Resend
	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromEmail),
		To:      []string{toEmail},
		Subject: "Welcome to our platform!",
		Html:    htmlBody,
	}

	_, err = s.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("failed to send welcome email: %w", err)
	}

	return nil
}

// renderTemplate renders email HTML template with data.
func (s *EmailService) renderTemplate(templateName string, data map[string]interface{}) (string, error) {
	tmpl, err := template.ParseFiles(fmt.Sprintf("web/templates/emails/%s.html", templateName))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
