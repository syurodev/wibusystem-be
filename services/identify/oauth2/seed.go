package oauth2

import (
    "context"
    "log"

    "github.com/jackc/pgx/v5/pgxpool"
)

func SeedDefaultClients(pool *pgxpool.Pool, issuer string) {
    ctx := context.Background()

    // Seed a public SPA client for dev
    _, err := pool.Exec(ctx, `
        INSERT INTO oauth2_clients (
            id, client_secret_hash, redirect_uris, grant_types, response_types,
            scopes, audience, public, client_name
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        ON CONFLICT (id) DO NOTHING
    `,
        "spa-client",
        "", // no secret for public client
        []string{issuer + "/callback", "http://localhost:3000/callback", "http://localhost:3000/api/auth/callback/oidc"},
        []string{"authorization_code", "refresh_token"},
        []string{"code"},
        []string{"openid", "profile", "email", "offline_access"},
        []string{},
        true,
        "Dev SPA Client",
    )
    if err != nil {
        log.Printf("seed: failed to insert spa-client: %v", err)
    }

    // Seed wibutime_web_client for NextAuth
    _, err = pool.Exec(ctx, `
        INSERT INTO oauth2_clients (
            id, client_secret_hash, redirect_uris, grant_types, response_types,
            scopes, audience, public, client_name
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        ON CONFLICT (id) DO NOTHING
    `,
        "wibutime_web_client",
        "db953e2a62fd4c6aa4599e4087d8dec3cab5f5b8badcfbd42dd4d0d6bfacc2f5", // client secret hash
        []string{"http://localhost:3000/api/auth/callback/oidc"},
        []string{"authorization_code", "refresh_token"},
        []string{"code"},
        []string{"openid", "profile", "email", "offline_access"},
        []string{},
        false, // not public, has client secret
        "Wibutime Web Client",
    )
    if err != nil {
        log.Printf("seed: failed to insert wibutime_web_client: %v", err)
    }
}

// EnsureNextAuthRedirects appends NextAuth callback redirect URIs to known dev clients if missing
func EnsureNextAuthRedirects(pool *pgxpool.Pool) {
    ctx := context.Background()
    nextAuthCB := "http://localhost:3000/api/auth/callback/oidc"

    // Update known clients to include NextAuth callback
    for _, cid := range []string{"spa-client", "wibutime_web_client"} {
        _, err := pool.Exec(ctx, `
            UPDATE oauth2_clients 
            SET redirect_uris = (SELECT ARRAY(SELECT DISTINCT unnest(redirect_uris || $2::text[])))
            WHERE id = $1
        `, cid, []string{nextAuthCB})
        if err != nil {
            log.Printf("seed: failed to ensure NextAuth redirect for %s: %v", cid, err)
        }
    }
}
