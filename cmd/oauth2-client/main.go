package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"wibusystem/internal/infrastructure/database"
	"wibusystem/internal/platform/config"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Command line flags
	clientID := flag.String("id", "", "Client ID (required)")
	clientName := flag.String("name", "", "Client name (required)")
	clientSecret := flag.String("secret", "", "Client secret (will be hashed)")
	redirectURIs := flag.String("redirect-uris", "http://localhost:3000/auth/callback", "Comma-separated redirect URIs")
	grantTypes := flag.String("grant-types", "authorization_code,refresh_token", "Comma-separated grant types")
	scopes := flag.String("scopes", "openid,profile,email,offline_access", "Comma-separated scopes")
	isPublic := flag.Bool("public", false, "Is this a public client (no secret required)")

	flag.Parse()

	// Validate required fields
	if *clientID == "" || *clientName == "" {
		log.Fatal("Both -id and -name are required")
	}

	if !*isPublic && *clientSecret == "" {
		log.Fatal("Client secret is required for confidential clients (use -public for public clients)")
	}

	// Load configuration
	cfg := config.Load()

	// Connect to database
	db, err := database.New(context.Background(), &cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Hash the client secret if provided
	var secretHash *string
	if *clientSecret != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(*clientSecret), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash client secret: %v", err)
		}
		hashStr := string(hash)
		secretHash = &hashStr
	}

	// Parse arrays
	redirectURIArray := strings.Split(*redirectURIs, ",")
	grantTypeArray := strings.Split(*grantTypes, ",")
	scopeArray := strings.Split(*scopes, ",")

	// Insert or update client
	query := `
		INSERT INTO identity.oauth2_clients (
			id,
			client_secret_hash,
			redirect_uris,
			grant_types,
			response_types,
			scopes,
			audience,
			public,
			client_name,
			token_endpoint_auth_method
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		)
		ON CONFLICT (id) DO UPDATE SET
			client_name = EXCLUDED.client_name,
			client_secret_hash = EXCLUDED.client_secret_hash,
			redirect_uris = EXCLUDED.redirect_uris,
			grant_types = EXCLUDED.grant_types,
			scopes = EXCLUDED.scopes,
			public = EXCLUDED.public,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, client_name, created_at, updated_at
	`

	var (
		id        string
		name      string
		createdAt time.Time
		updatedAt time.Time
	)

	err = db.Pool().QueryRow(
		context.Background(),
		query,
		*clientID,
		secretHash,
		redirectURIArray,
		grantTypeArray,
		[]string{"code"},
		scopeArray,
		[]string{"wibusystem-api"},
		*isPublic,
		*clientName,
		"client_secret_basic",
	).Scan(&id, &name, &createdAt, &updatedAt)

	if err != nil {
		log.Fatalf("Failed to create OAuth2 client: %v", err)
	}

	// Display result
	fmt.Println("\n✅ OAuth2 Client Created Successfully!")
	fmt.Println("=====================================")
	fmt.Printf("Client ID:        %s\n", id)
	fmt.Printf("Client Name:      %s\n", name)
	if *clientSecret != "" {
		fmt.Printf("Client Secret:    %s\n", *clientSecret)
	}
	fmt.Printf("Public Client:    %v\n", *isPublic)
	fmt.Printf("Redirect URIs:    %s\n", strings.Join(redirectURIArray, ", "))
	fmt.Printf("Grant Types:      %s\n", strings.Join(grantTypeArray, ", "))
	fmt.Printf("Scopes:           %s\n", strings.Join(scopeArray, ", "))
	fmt.Printf("Created At:       %s\n", createdAt)
	fmt.Printf("Updated At:       %s\n", updatedAt)
	fmt.Println("\n⚠️  IMPORTANT: Store the client secret securely!")
	fmt.Println("   It cannot be retrieved later, only reset.")
}
