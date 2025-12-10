package resend

import (
	"fmt"

	"github.com/resend/resend-go/v2"
	"go.uber.org/zap"

	"system/configs"
)

// Client handles interaction with Resend API.
type Client struct {
	client *resend.Client
	config *configs.EmailConfig
	logger *zap.Logger
}

// NewClient creates a new instance of Client.
func NewClient(config *configs.EmailConfig, logger *zap.Logger) *Client {
	client := resend.NewClient(config.APIKey)

	return &Client{
		client: client,
		config: config,
		logger: logger,
	}
}

// Send sends a generic email.
func (c *Client) Send(to []string, subject string, htmlBody string) error {
	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", c.config.FromName, c.config.FromEmail),
		To:      to,
		Subject: subject,
		Html:    htmlBody,
	}

	sent, err := c.client.Emails.Send(params)
	if err != nil {
		c.logger.Error("Failed to send email via Resend",
			zap.Strings("to", to),
			zap.String("subject", subject),
			zap.Error(err),
		)
		return fmt.Errorf("failed to send email: %w", err)
	}

	c.logger.Info("Email sent successfully",
		zap.Strings("to", to),
		zap.String("email_id", sent.Id),
		zap.String("subject", subject),
	)

	return nil
}
