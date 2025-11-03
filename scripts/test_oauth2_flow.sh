#!/bin/bash

# =====================================================
# OAuth2 Authorization Code Flow Test Script
# =====================================================
# This script tests the complete OAuth2 flow:
# 1. Discovery endpoint
# 2. Authorization request
# 3. User login
# 4. User consent
# 5. Token exchange
# 6. UserInfo request
# 7. Token refresh
# =====================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost:8080"
# Use the first client from database (System Admin Dashboard or Test Web Application)
CLIENT_ID="11cacf6d-607b-4128-9b56-4d0b540319da"  # System Admin Dashboard
# To get CLIENT_SECRET, you need to regenerate it via Admin API:
# curl -X POST http://localhost:8080/api/v1/admin/oauth2/clients/$CLIENT_ID/regenerate-secret
CLIENT_SECRET="79Vm_UeWF7D42CK_yafWld72zOBEL7tRXpDnWL84LzA="
REDIRECT_URI="http://localhost:3000/callback"
# Update with your actual registered user credentials
USERNAME="test@example.com"
PASSWORD="StrongPass123"

# Helper functions
print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

# =====================================================
# Setup: Regenerate Client Secret
# =====================================================
setup_client_secret() {
    print_header "Setup: Regenerate Client Secret"

    print_info "Regenerating secret for client: $CLIENT_ID"

    RESPONSE=$(curl -s -X POST "${BASE_URL}/api/v1/admin/oauth2/clients/${CLIENT_ID}/regenerate-secret")

    if echo "$RESPONSE" | grep -q "client_secret"; then
        print_success "Client secret regenerated successfully"
        NEW_SECRET=$(echo "$RESPONSE" | jq -r '.data.client_secret')

        print_info "New Client Secret: ${GREEN}${NEW_SECRET}${NC}"
        print_info "Please update CLIENT_SECRET variable in this script with the above value."

        echo ""
        echo -e "${YELLOW}Copy this line to update the script:${NC}"
        echo -e "${GREEN}CLIENT_SECRET=\"${NEW_SECRET}\"${NC}"
        echo ""

    else
        print_error "Failed to regenerate client secret"
        echo "$RESPONSE" | jq '.'
        exit 1
    fi
}

# =====================================================
# Test 1: Discovery Endpoint
# =====================================================
test_discovery() {
    print_header "Test 1: OpenID Connect Discovery"

    RESPONSE=$(curl -s "${BASE_URL}/.well-known/openid-configuration")

    if echo "$RESPONSE" | grep -q "authorization_endpoint"; then
        print_success "Discovery endpoint working"
        echo "$RESPONSE" | jq '.'
    else
        print_error "Discovery endpoint failed"
        echo "$RESPONSE"
        exit 1
    fi
}

# =====================================================
# Test 2: JWKS Endpoint
# =====================================================
test_jwks() {
    print_header "Test 2: JWKS Endpoint"

    RESPONSE=$(curl -s "${BASE_URL}/.well-known/jwks.json")

    if echo "$RESPONSE" | grep -q "keys"; then
        print_success "JWKS endpoint working"
        echo "$RESPONSE" | jq '.'
    else
        print_error "JWKS endpoint failed"
        echo "$RESPONSE"
        exit 1
    fi
}

# =====================================================
# Test 3: Authorization Code Flow (Manual)
# =====================================================
test_authorization_manual() {
    print_header "Test 3: Authorization Code Flow (Manual)"

    # Generate state and code_verifier for PKCE
    STATE=$(openssl rand -hex 16)
    CODE_VERIFIER=$(openssl rand -base64 32 | tr -d '=+/' | cut -c1-43)
    CODE_CHALLENGE=$(echo -n "$CODE_VERIFIER" | openssl dgst -binary -sha256 | openssl base64 | tr -d '=+/' | cut -c1-43)

    print_info "Generated State: $STATE"
    print_info "Generated Code Verifier: $CODE_VERIFIER"
    print_info "Generated Code Challenge: $CODE_CHALLENGE"

    # Build authorization URL
    AUTH_URL="${BASE_URL}/oauth2/auth"
    AUTH_URL+="?client_id=${CLIENT_ID}"
    AUTH_URL+="&redirect_uri=${REDIRECT_URI}"
    AUTH_URL+="&response_type=code"
    AUTH_URL+="&scope=openid+profile+email+offline_access"
    AUTH_URL+="&state=${STATE}"
    AUTH_URL+="&code_challenge=${CODE_CHALLENGE}"
    AUTH_URL+="&code_challenge_method=S256"

    print_info "Authorization URL:"
    echo -e "${BLUE}${AUTH_URL}${NC}\n"

    print_info "Steps to test manually:"
    echo "1. Open the URL above in a browser"
    echo "2. Login with: ${USERNAME} / ${PASSWORD}"
    echo "3. Grant consent"
    echo "4. Copy the 'code' from redirect URL"
    echo "5. Exchange code for tokens (see next test)"
}

# =====================================================
# Test 4: Token Exchange (requires authorization code)
# =====================================================
test_token_exchange() {
    print_header "Test 4: Token Exchange"

    if [ -z "$1" ]; then
        print_error "Authorization code required"
        echo "Usage: $0 token <authorization_code> <code_verifier>"
        exit 1
    fi

    AUTH_CODE="$1"
    CODE_VERIFIER="${2:-}"

    print_info "Exchanging authorization code for tokens..."

    # Prepare token request
    TOKEN_REQUEST="grant_type=authorization_code"
    TOKEN_REQUEST+="&code=${AUTH_CODE}"
    TOKEN_REQUEST+="&redirect_uri=${REDIRECT_URI}"

    if [ -n "$CODE_VERIFIER" ]; then
        TOKEN_REQUEST+="&code_verifier=${CODE_VERIFIER}"
    fi

    # Make token request with client credentials
    RESPONSE=$(curl -s -X POST "${BASE_URL}/oauth2/token" \
        -u "${CLIENT_ID}:${CLIENT_SECRET}" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "$TOKEN_REQUEST")

    if echo "$RESPONSE" | grep -q "access_token"; then
        print_success "Token exchange successful"
        echo "$RESPONSE" | jq '.'

        # Extract tokens
        ACCESS_TOKEN=$(echo "$RESPONSE" | jq -r '.access_token')
        REFRESH_TOKEN=$(echo "$RESPONSE" | jq -r '.refresh_token')
        ID_TOKEN=$(echo "$RESPONSE" | jq -r '.id_token')

        # Save to file for later use
        echo "$ACCESS_TOKEN" > /tmp/oauth2_access_token.txt
        echo "$REFRESH_TOKEN" > /tmp/oauth2_refresh_token.txt
        echo "$ID_TOKEN" > /tmp/oauth2_id_token.txt

        print_success "Tokens saved to /tmp/oauth2_*.txt"
    else
        print_error "Token exchange failed"
        echo "$RESPONSE" | jq '.'
        exit 1
    fi
}

# =====================================================
# Test 5: UserInfo Endpoint
# =====================================================
test_userinfo() {
    print_header "Test 5: UserInfo Endpoint"

    if [ ! -f /tmp/oauth2_access_token.txt ]; then
        print_error "Access token not found. Run token exchange first."
        exit 1
    fi

    ACCESS_TOKEN=$(cat /tmp/oauth2_access_token.txt)

    print_info "Requesting user info..."

    RESPONSE=$(curl -s "${BASE_URL}/oauth2/userinfo" \
        -H "Authorization: Bearer ${ACCESS_TOKEN}")

    if echo "$RESPONSE" | grep -q "sub"; then
        print_success "UserInfo request successful"
        echo "$RESPONSE" | jq '.'
    else
        print_error "UserInfo request failed"
        echo "$RESPONSE"
        exit 1
    fi
}

# =====================================================
# Test 6: Token Refresh
# =====================================================
test_refresh_token() {
    print_header "Test 6: Token Refresh"

    if [ ! -f /tmp/oauth2_refresh_token.txt ]; then
        print_error "Refresh token not found. Run token exchange first."
        exit 1
    fi

    REFRESH_TOKEN=$(cat /tmp/oauth2_refresh_token.txt)

    print_info "Refreshing access token..."

    RESPONSE=$(curl -s -X POST "${BASE_URL}/oauth2/token" \
        -u "${CLIENT_ID}:${CLIENT_SECRET}" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "grant_type=refresh_token&refresh_token=${REFRESH_TOKEN}")

    if echo "$RESPONSE" | grep -q "access_token"; then
        print_success "Token refresh successful"
        echo "$RESPONSE" | jq '.'

        # Update access token
        NEW_ACCESS_TOKEN=$(echo "$RESPONSE" | jq -r '.access_token')
        echo "$NEW_ACCESS_TOKEN" > /tmp/oauth2_access_token.txt
        print_success "New access token saved"
    else
        print_error "Token refresh failed"
        echo "$RESPONSE" | jq '.'
        exit 1
    fi
}

# =====================================================
# Test 7: Client Credentials Grant
# =====================================================
test_client_credentials() {
    print_header "Test 7: Client Credentials Grant"

    API_CLIENT_ID="10000000-0000-0000-0000-000000000004"
    API_CLIENT_SECRET="test-client-secret"

    print_info "Requesting token with client credentials..."

    RESPONSE=$(curl -s -X POST "${BASE_URL}/oauth2/token" \
        -u "${API_CLIENT_ID}:${API_CLIENT_SECRET}" \
        -H "Content-Type: application/x-www-form-urlencoded" \
        -d "grant_type=client_credentials&scope=api:read+api:write")

    if echo "$RESPONSE" | grep -q "access_token"; then
        print_success "Client credentials grant successful"
        echo "$RESPONSE" | jq '.'
    else
        print_error "Client credentials grant failed"
        echo "$RESPONSE" | jq '.'
        exit 1
    fi
}

# =====================================================
# Main Script
# =====================================================
main() {
    case "${1:-all}" in
        setup)
            setup_client_secret
            ;;
        discovery)
            test_discovery
            ;;
        jwks)
            test_jwks
            ;;
        auth)
            test_authorization_manual
            ;;
        token)
            test_token_exchange "$2" "$3"
            ;;
        userinfo)
            test_userinfo
            ;;
        refresh)
            test_refresh_token
            ;;
        client-creds)
            test_client_credentials
            ;;
        all)
            # Check if CLIENT_SECRET is set
            if [ -z "$CLIENT_SECRET" ]; then
                print_error "CLIENT_SECRET is not set!"
                print_info "Run the following command first to generate a client secret:"
                echo -e "${YELLOW}  $0 setup${NC}\n"
                exit 1
            fi

            print_header "OAuth2 Flow Test Suite"
            test_discovery
            test_jwks
            test_authorization_manual
            print_info "\n⚠️  Complete the authorization manually, then run:"
            print_info "  $0 token <code> <code_verifier>"
            print_info "  $0 userinfo"
            print_info "  $0 refresh"
            ;;
        *)
            echo "Usage: $0 {setup|discovery|jwks|auth|token|userinfo|refresh|client-creds|all}"
            echo ""
            echo "Commands:"
            echo "  setup          - Generate new client secret (run this first!)"
            echo "  discovery      - Test OpenID Connect Discovery endpoint"
            echo "  jwks           - Test JWKS endpoint"
            echo "  auth           - Generate authorization URL (manual test)"
            echo "  token <code> <verifier> - Exchange authorization code for tokens"
            echo "  userinfo       - Test UserInfo endpoint"
            echo "  refresh        - Test token refresh"
            echo "  client-creds   - Test client credentials grant"
            echo "  all            - Run all automated tests"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
