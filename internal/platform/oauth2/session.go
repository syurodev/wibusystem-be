package oauth2

import (
	"time"

	"github.com/ory/fosite"
)

// Session là custom session implementation cho OpenID Connect
// Sử dụng fosite.DefaultSession trực tiếp vì nó đã đủ cho hầu hết use cases
type Session struct {
	*fosite.DefaultSession
}

// NewSession tạo một session mới với thông tin user
func NewSession(subject string) *Session {
	now := time.Now().UTC()

	return &Session{
		DefaultSession: &fosite.DefaultSession{
			Subject:  subject,
			Username: subject,
			ExpiresAt: map[fosite.TokenType]time.Time{
				fosite.AccessToken:   now.Add(time.Hour * 1),
				fosite.RefreshToken:  now.Add(time.Hour * 24 * 30),
				fosite.AuthorizeCode: now.Add(time.Minute * 10),
			},
		},
	}
}
