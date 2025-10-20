#!/bin/bash

# Load Testing Script for Pack Calculator
# Tests throughput, latency, and resource utilization

set -e

API_URL="${API_URL:-http://localhost:8080}"
RESULTS_DIR="load-test-results"
mkdir -p "$RESULTS_DIR"

echo "=========================================="
echo "Pack Calculator - Load Testing Suite"
echo "=========================================="
echo "API URL: $API_URL"
echo "Results: $RESULTS_DIR"
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Test 1: Baseline Performance
echo -e "${YELLOW}Test 1: Baseline Performance (10 requests)${NC}"
echo "Testing single-threaded performance..."

total_time=0
for i in {1..10}; do
  start=$(date +%s%N)
  curl -s -o /dev/null -w "%{http_code}" -X POST "$API_URL/api/calculate" \
    -H "Content-Type: application/json" \
    -d '{"amount": 501}' > /dev/null
  end=$(date +%s%N)
  time_ms=$(( ($end - $start) / 1000000 ))
  total_time=$(( $total_time + $time_ms ))
  echo "  Request $i: ${time_ms}ms"
done

avg_time=$(( $total_time / 10 ))
echo -e "${GREEN}Average response time: ${avg_time}ms${NC}"
echo ""

# Test 2: Concurrent Load
echo -e "${YELLOW}Test 2: Concurrent Load (100 requests, 10 concurrent)${NC}"
echo "Testing concurrent request handling..."

start_time=$(date +%s)
success=0
failed=0

for i in {1..100}; do
  {
    http_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API_URL/api/calculate" \
      -H "Content-Type: application/json" \
      -d "{\"amount\": $((250 + i * 10))}")
    
    if [ "$http_code" = "200" ]; then
      ((success++))
    else
      ((failed++))
    fi
  } &
  
  # Limit concurrency to 10
  if [ $((i % 10)) -eq 0 ]; then
    wait
  fi
done

wait
end_time=$(date +%s)
duration=$((end_time - start_time))

echo -e "${GREEN}Success: $success, Failed: $failed${NC}"
echo "Duration: ${duration}s"
echo "Throughput: $((100 / duration)) req/s"
echo ""

# Test 3: Cache Performance
echo -e "${YELLOW}Test 3: Cache Performance (same request 50 times)${NC}"
echo "Testing cache hit performance..."

# First request (cache miss)
start=$(date +%s%N)
curl -s -X POST "$API_URL/api/calculate" \
  -H "Content-Type: application/json" \
  -d '{"amount": 12001}' > /dev/null
end=$(date +%s%N)
miss_time=$(( ($end - $start) / 1000000 ))

# Subsequent requests (cache hits)
hit_times=()
for i in {1..49}; do
  start=$(date +%s%N)
  curl -s -X POST "$API_URL/api/calculate" \
    -H "Content-Type: application/json" \
    -d '{"amount": 12001}' > /dev/null
  end=$(date +%s%N)
  time_ms=$(( ($end - $start) / 1000000 ))
  hit_times+=($time_ms)
done

# Calculate average hit time
hit_total=0
for time in "${hit_times[@]}"; do
  hit_total=$((hit_total + time))
done
avg_hit=$(( hit_total / 49 ))

echo "Cache miss: ${miss_time}ms"
echo "Cache hit avg: ${avg_hit}ms"
echo -e "${GREEN}Speed improvement: $((miss_time / avg_hit))x${NC}"
echo ""

# Test 4: Different Pack Sizes
echo -e "${YELLOW}Test 4: Various Input Sizes${NC}"
echo "Testing different calculation complexities..."

amounts=(1 100 1000 10000 100000 1000000 5000000)

for amount in "${amounts[@]}"; do
  echo -n "  Amount $amount: "
  start=$(date +%s%N)
  result=$(curl -s -X POST "$API_URL/api/calculate" \
    -H "Content-Type: application/json" \
    -d "{\"amount\": $amount}")
  end=$(date +%s%N)
  time_ms=$(( ($end - $start) / 1000000 ))
  
  if echo "$result" | grep -q "total_items"; then
    echo -e "${GREEN}${time_ms}ms${NC}"
  else
    echo -e "${RED}FAILED${NC}"
  fi
done
echo ""

# Test 5: Edge Cases
echo -e "${YELLOW}Test 5: Edge Cases${NC}"
echo "Testing boundary conditions..."

# Test minimum
echo -n "  Minimum (1 item): "
result=$(curl -s -X POST "$API_URL/api/calculate" \
  -H "Content-Type: application/json" \
  -d '{"amount": 1}')
if echo "$result" | grep -q "total_items"; then
  echo -e "${GREEN}PASS${NC}"
else
  echo -e "${RED}FAIL${NC}"
fi

# Test zero (should fail)
echo -n "  Zero (should error): "
result=$(curl -s -X POST "$API_URL/api/calculate" \
  -H "Content-Type: application/json" \
  -d '{"amount": 0}')
if echo "$result" | grep -q "error"; then
  echo -e "${GREEN}PASS${NC}"
else
  echo -e "${RED}FAIL${NC}"
fi

# Test negative (should fail)
echo -n "  Negative (should error): "
result=$(curl -s -X POST "$API_URL/api/calculate" \
  -H "Content-Type: application/json" \
  -d '{"amount": -100}')
if echo "$result" | grep -q "error"; then
  echo -e "${GREEN}PASS${NC}"
else
  echo -e "${RED}FAIL${NC}"
fi

# Test maximum
echo -n "  Maximum (10M items): "
start=$(date +%s)
result=$(curl -s -X POST "$API_URL/api/calculate" \
  -H "Content-Type: application/json" \
  -d '{"amount": 10000000}')
end=$(date +%s)
duration=$((end - start))
if echo "$result" | grep -q "total_items"; then
  echo -e "${GREEN}PASS${NC} (${duration}s)"
else
  echo -e "${RED}FAIL${NC}"
fi

# Test over maximum (should fail)
echo -n "  Over maximum (should error): "
result=$(curl -s -X POST "$API_URL/api/calculate" \
  -H "Content-Type: application/json" \
  -d '{"amount": 10000001}')
if echo "$result" | grep -q "error"; then
  echo -e "${GREEN}PASS${NC}"
else
  echo -e "${RED}FAIL${NC}"
fi

echo ""

# Test 6: Rate Limiting
echo -e "${YELLOW}Test 6: Rate Limiting${NC}"
echo "Testing rate limiter (sending 50 rapid requests)..."

rate_limited=0
for i in {1..50}; do
  http_code=$(curl -s -o /dev/null -w "%{http_code}" -X POST "$API_URL/api/calculate" \
    -H "Content-Type: application/json" \
    -d '{"amount": 100}')
  
  if [ "$http_code" = "429" ]; then
    ((rate_limited++))
  fi
done

if [ $rate_limited -gt 0 ]; then
  echo -e "${GREEN}Rate limiting working: $rate_limited requests blocked${NC}"
else
  echo -e "${YELLOW}No rate limiting detected (may need faster requests)${NC}"
fi
echo ""

# Test 7: Compression
echo -e "${YELLOW}Test 7: HTTP Compression${NC}"
echo "Testing gzip compression..."

# Without compression
size_uncompressed=$(curl -s -X POST "$API_URL/api/calculate" \
  -H "Content-Type: application/json" \
  -d '{"amount": 12001}' | wc -c)

# With compression
size_compressed=$(curl -s -H "Accept-Encoding: gzip" -X POST "$API_URL/api/calculate" \
  -H "Content-Type: application/json" \
  -d '{"amount": 12001}' --compressed | wc -c)

echo "Uncompressed: ${size_uncompressed} bytes"
echo "Compressed: ${size_compressed} bytes"

if [ $size_compressed -lt $size_uncompressed ]; then
  reduction=$((100 - (size_compressed * 100 / size_uncompressed)))
  echo -e "${GREEN}Compression working: ${reduction}% size reduction${NC}"
else
  echo -e "${YELLOW}Compression may not be working${NC}"
fi
echo ""

# Test 8: API Endpoints
echo -e "${YELLOW}Test 8: All API Endpoints${NC}"

echo -n "  GET /health: "
if curl -s "$API_URL/health" | grep -q "healthy"; then
  echo -e "${GREEN}PASS${NC}"
else
  echo -e "${RED}FAIL${NC}"
fi

echo -n "  GET /api/packs: "
if curl -s "$API_URL/api/packs" | grep -q '\['; then
  echo -e "${GREEN}PASS${NC}"
else
  echo -e "${RED}FAIL${NC}"
fi

echo -n "  GET /api/orders: "
if curl -s "$API_URL/api/orders" | grep -q '\['; then
  echo -e "${GREEN}PASS${NC}"
else
  echo -e "${RED}FAIL${NC}"
fi

echo ""

# Summary
echo "=========================================="
echo -e "${GREEN}Load Testing Complete!${NC}"
echo "=========================================="
echo ""
echo "Key Metrics:"
echo "  - Average latency: ${avg_time}ms"
echo "  - Cache speedup: $((miss_time / avg_hit))x"
echo "  - Throughput: $((100 / duration)) req/s"
echo ""
echo "All results saved to: $RESULTS_DIR/"
echo ""

# Save summary
cat > "$RESULTS_DIR/summary.txt" << EOF
Pack Calculator Load Test Results
Generated: $(date)

Baseline Performance:
  Average response time: ${avg_time}ms

Concurrent Load:
  Requests: 100
  Success: $success
  Failed: $failed
  Duration: ${duration}s
  Throughput: $((100 / duration)) req/s

Cache Performance:
  Cache miss: ${miss_time}ms
  Cache hit: ${avg_hit}ms
  Speedup: $((miss_time / avg_hit))x

Rate Limiting:
  Blocked requests: $rate_limited / 50

Compression:
  Size reduction: $((100 - (size_compressed * 100 / size_uncompressed)))%
EOF

echo "Summary saved to: $RESULTS_DIR/summary.txt"

