package session

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Manager struct {
	Secret   []byte
	Cookie   string
	MaxAge   time.Duration
	Secure   bool
	SameSite http.SameSite
	Domain   string
	Path     string
}

type payload struct {
	Sub string    `json:"sub"`
	Exp time.Time `json:"exp"`
}

func New(secret string, cookie string, maxAge time.Duration, secure bool) *Manager {
	return &Manager{
		Secret:   []byte(secret),
		Cookie:   cookie,
		MaxAge:   maxAge,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	}
}

func (m *Manager) Set(c *gin.Context, userID string) error {
	p := payload{Sub: userID, Exp: time.Now().Add(m.MaxAge)}
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	mac := hmac.New(sha256.New, m.Secret)
	mac.Write(data)
	sig := mac.Sum(nil)
	token := base64.RawURLEncoding.EncodeToString(data) + "." + base64.RawURLEncoding.EncodeToString(sig)
	cookie := &http.Cookie{
		Name:  m.Cookie,
		Value: token,
		Path:  m.Path,
		// Domain not set for localhost development to avoid cookie issues
		HttpOnly: true,
		Secure:   m.Secure,
		SameSite: http.SameSiteLaxMode, // Lax mode for localhost development
		Expires:  p.Exp,
	}
	http.SetCookie(c.Writer, cookie)
	return nil
}

func (m *Manager) Clear(c *gin.Context) {
	cookie := &http.Cookie{
		Name:  m.Cookie,
		Value: "",
		Path:  m.Path,
		// Domain not set for localhost development to match Set method
		HttpOnly: true,
		Secure:   m.Secure,
		SameSite: http.SameSiteLaxMode, // Match SameSite from Set method
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	}
	http.SetCookie(c.Writer, cookie)
}

func (m *Manager) Get(c *gin.Context) (string, error) {
	cookie, err := c.Request.Cookie(m.Cookie)
	if err != nil {
		return "", err
	}
	parts := []byte(cookie.Value)
	dot := -1
	for i := range parts {
		if parts[i] == '.' {
			dot = i
			break
		}
	}
	if dot <= 0 {
		return "", fmt.Errorf("invalid session")
	}
	dataB, err := base64.RawURLEncoding.DecodeString(string(parts[:dot]))
	if err != nil {
		return "", err
	}
	sigB, err := base64.RawURLEncoding.DecodeString(string(parts[dot+1:]))
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, m.Secret)
	mac.Write(dataB)
	if !hmac.Equal(mac.Sum(nil), sigB) {
		return "", fmt.Errorf("invalid signature")
	}
	var p payload
	if err := json.Unmarshal(dataB, &p); err != nil {
		return "", err
	}
	if time.Now().After(p.Exp) {
		return "", fmt.Errorf("session expired")
	}
	return p.Sub, nil
}
