#!/bin/bash

# API Test Script
# Tests all endpoints of the Pack Calculator API

API_URL="${API_URL:-http://localhost:8080}"

echo "Testing Pack Calculator API at $API_URL"
echo "=========================================="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Test counter
PASSED=0
FAILED=0

test_endpoint() {
    local name=$1
    local method=$2
    local endpoint=$3
    local data=$4
    
    echo -n "Testing $name... "
    
    if [ -z "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X $method "$API_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method "$API_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)
    
    if [ "$http_code" -ge 200 ] && [ "$http_code" -lt 300 ]; then
        echo -e "${GREEN}PASSED${NC} (HTTP $http_code)"
        echo "Response: $body"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}FAILED${NC} (HTTP $http_code)"
        echo "Response: $body"
        FAILED=$((FAILED + 1))
    fi
    echo ""
}

echo ""
echo "1. Health Check"
echo "---------------"
test_endpoint "Health Check" "GET" "/health"

echo "2. Get Pack Sizes"
echo "-----------------"
test_endpoint "Get Pack Sizes" "GET" "/api/packs"

echo "3. Calculate Packs - Simple Case (250 items)"
echo "--------------------------------------------"
test_endpoint "Calculate 250 items" "POST" "/api/calculate" '{"amount": 250}'

echo "4. Calculate Packs - Complex Case (501 items)"
echo "---------------------------------------------"
test_endpoint "Calculate 501 items" "POST" "/api/calculate" '{"amount": 501}'

echo "5. Calculate Packs - Large Order (12001 items)"
echo "-----------------------------------------------"
test_endpoint "Calculate 12001 items" "POST" "/api/calculate" '{"amount": 12001}'

echo "6. Get Order History"
echo "-------------------"
test_endpoint "Get Orders" "GET" "/api/orders?limit=10"

echo "7. Add Pack Size (test with 750)"
echo "--------------------------------"
test_endpoint "Add Pack Size 750" "POST" "/api/packs" '{"size": 750}'

echo "8. Verify Pack Size Added"
echo "------------------------"
test_endpoint "Get Pack Sizes Again" "GET" "/api/packs"

echo "9. Calculate with New Pack Size (800 items)"
echo "-------------------------------------------"
test_endpoint "Calculate 800 items" "POST" "/api/calculate" '{"amount": 800}'

echo "10. Delete Pack Size (750)"
echo "-------------------------"
test_endpoint "Delete Pack Size 750" "DELETE" "/api/packs/750"

echo ""
echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo "Total: $((PASSED + FAILED))"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed.${NC}"
    exit 1
fi

