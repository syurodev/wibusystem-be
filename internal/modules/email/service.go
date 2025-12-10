package email

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"

	"system/configs"
)

// EmailSender interface for sending raw emails
type EmailSender interface {
	Send(to []string, subject string, htmlBody string) error
}

// Service handles business email logic
type Service struct {
	sender EmailSender
	config *configs.EmailConfig
}

// NewService creates a new Email Service
func NewService(sender EmailSender, config *configs.EmailConfig) *Service {
	return &Service{
		sender: sender,
		config: config,
	}
}

// SendVerificationEmail sends verification email with token
func (s *Service) SendVerificationEmail(ctx context.Context, toEmail, toName, token string) error {
	verificationURL := strings.ReplaceAll(s.config.VerificationURL, "{{.Token}}", token)

	data := map[string]any{
		"Name":            toName,
		"VerificationURL": verificationURL,
		"Token":           token,
	}

	htmlBody, err := s.renderTemplate("verification", data)
	if err != nil {
		return fmt.Errorf("failed to render verification template: %w", err)
	}

	return s.sender.Send([]string{toEmail}, "Verify your email address", htmlBody)
}

// SendPasswordResetEmail sends password reset email with token
func (s *Service) SendPasswordResetEmail(ctx context.Context, toEmail, toName, token string) error {
	resetURL := strings.ReplaceAll(s.config.PasswordResetURL, "{{.Token}}", token)

	data := map[string]any{
		"Name":     toName,
		"ResetURL": resetURL,
		"Token":    token,
	}

	htmlBody, err := s.renderTemplate("password_reset", data)
	if err != nil {
		return fmt.Errorf("failed to render password reset template: %w", err)
	}

	return s.sender.Send([]string{toEmail}, "Reset your password", htmlBody)
}

// SendWelcomeEmail sends welcome email
func (s *Service) SendWelcomeEmail(ctx context.Context, toEmail, toName string) error {
	data := map[string]any{
		"Name":    toName,
		"BaseURL": s.config.BaseURL,
	}

	htmlBody, err := s.renderTemplate("welcome", data)
	if err != nil {
		return fmt.Errorf("failed to render welcome template: %w", err)
	}

	return s.sender.Send([]string{toEmail}, "Welcome to our platform!", htmlBody)
}

// renderTemplate renders email HTML template with data.
func (s *Service) renderTemplate(templateName string, data map[string]any) (string, error) {
	tmpl, err := template.ParseFiles(fmt.Sprintf("web/templates/emails/%s.html", templateName))
	if err != nil {
		return "", fmt.Errorf("failed to parse template files: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
