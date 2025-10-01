package oauth2

import (
	"context"
	"testing"

	"github.com/ory/fosite"
	"golang.org/x/crypto/bcrypt"
)

func TestStoreAuthenticateWithBcryptSecret(t *testing.T) {
	hashed, err := bcrypt.GenerateFromPassword([]byte("super-secret"), 10)
	if err != nil {
		t.Fatalf("failed to hash secret: %v", err)
	}

	store := &Store{
		clients: map[string]*Client{
			"client-123": {
				DefaultClient: &fosite.DefaultClient{
					ID:     "client-123",
					Secret: hashed,
				},
			},
		},
	}

	if err := store.Authenticate(context.Background(), "client-123", "super-secret"); err != nil {
		t.Fatalf("expected authentication success, got error: %v", err)
	}

	if err := store.Authenticate(context.Background(), "client-123", "wrong-secret"); err == nil {
		t.Fatalf("expected authentication failure for wrong secret")
	}
}

func TestStoreAuthenticateRejectsMissingSecret(t *testing.T) {
	store := &Store{
		clients: map[string]*Client{
			"public-client": {
				DefaultClient: &fosite.DefaultClient{
					ID: "public-client",
				},
			},
		},
	}

	if err := store.Authenticate(context.Background(), "public-client", "anything"); err == nil {
		t.Fatalf("expected error when client has no stored secret")
	}
}
