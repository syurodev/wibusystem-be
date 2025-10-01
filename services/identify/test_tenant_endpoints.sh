#!/bin/bash

# Tenant API Testing Script
# Make sure the Identity Service is running on localhost:8080

BASE_URL="http://localhost:8080/api/v1"
echo "🧪 Testing Tenant API Endpoints..."
echo "=================================="

# First, let's check if service is running
echo "📋 Checking service health..."
curl -s -o /dev/null -w "%{http_code}" $BASE_URL/../.well-known/openid_configuration
if [ $? -ne 0 ]; then
    echo "❌ Service is not running. Please start with: go run main.go"
    exit 1
fi
echo "✅ Service is running"

# Test without authentication first (should fail)
echo -e "\n🔒 Testing without authentication (should fail)..."
echo "POST /api/v1/tenants"
curl -s -X POST $BASE_URL/tenants \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Tenant"}' \
  | jq . 2>/dev/null || echo "Response not JSON"

echo -e "\nGET /api/v1/tenants"
curl -s -X GET $BASE_URL/tenants \
  | jq . 2>/dev/null || echo "Response not JSON"

# Note: For authenticated testing, you would need to:
# 1. Create a user and get an access token
# 2. Use that token in Authorization header

echo -e "\n📝 To test with authentication:"
echo "1. Create a user account first"
echo "2. Login to get access token"
echo "3. Use token in Authorization: Bearer <token> header"

echo -e "\n📄 Example authenticated requests:"
echo "# Create tenant"
echo 'curl -X POST '"$BASE_URL"'/tenants \'
echo '  -H "Content-Type: application/json" \'
echo '  -H "Authorization: Bearer YOUR_TOKEN" \'
echo '  -d '"'"'{"name": "My Company", "slug": "my-company", "description": "Test company"}'"'"

echo -e "\n# List tenants"
echo 'curl -X GET '"$BASE_URL"'/tenants \'
echo '  -H "Authorization: Bearer YOUR_TOKEN"'

echo -e "\n✅ Basic connectivity test completed"
echo "🔗 See test_tenant_api.md for detailed API documentation"
