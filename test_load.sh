#!/bin/bash

# Script to test rate limiter under load

BASE_URL="http://localhost:8080"
TOTAL_REQUESTS=20
CONCURRENT=1

echo "=========================================="
echo "Rate Limiter Load Test"
echo "=========================================="
echo ""

# Test 1: IP-based rate limiting
echo "Test 1: IP-based rate limiting (should allow 10 req/s, then block)"
echo "Sending $TOTAL_REQUESTS requests..."
echo ""

success_count=0
blocked_count=0

for i in $(seq 1 $TOTAL_REQUESTS); do
    response=$(curl -s -o /dev/null -w "%{http_code}" $BASE_URL)
    
    if [ "$response" == "200" ]; then
        echo "Request $i: ✓ Success (200)"
        ((success_count++))
    elif [ "$response" == "429" ]; then
        echo "Request $i: ✗ Blocked (429)"
        ((blocked_count++))
    else
        echo "Request $i: ? Unexpected ($response)"
    fi
    
    # Small delay to simulate real requests
    sleep 0.05
done

echo ""
echo "Results: $success_count successful, $blocked_count blocked"
echo ""
echo "=========================================="

# Test 2: Token-based rate limiting
echo ""
echo "Test 2: Token-based rate limiting (token limit: 100 req/s)"
echo "Sending $TOTAL_REQUESTS requests with API_KEY..."
echo ""

success_count=0
blocked_count=0

for i in $(seq 1 $TOTAL_REQUESTS); do
    response=$(curl -s -o /dev/null -w "%{http_code}" -H "API_KEY: test_token" $BASE_URL)
    
    if [ "$response" == "200" ]; then
        echo "Request $i: ✓ Success (200)"
        ((success_count++))
    elif [ "$response" == "429" ]; then
        echo "Request $i: ✗ Blocked (429)"
        ((blocked_count++))
    else
        echo "Request $i: ? Unexpected ($response)"
    fi
    
    sleep 0.05
done

echo ""
echo "Results: $success_count successful, $blocked_count blocked"
echo ""
echo "=========================================="

# Test 3: Custom token with specific limit
echo ""
echo "Test 3: Custom token 'abc123' (limit: 100 req/s)"
echo "Sending $TOTAL_REQUESTS requests..."
echo ""

success_count=0
blocked_count=0

for i in $(seq 1 $TOTAL_REQUESTS); do
    response=$(curl -s -o /dev/null -w "%{http_code}" -H "API_KEY: abc123" $BASE_URL)
    
    if [ "$response" == "200" ]; then
        echo "Request $i: ✓ Success (200)"
        ((success_count++))
    elif [ "$response" == "429" ]; then
        echo "Request $i: ✗ Blocked (429)"
        ((blocked_count++))
    else
        echo "Request $i: ? Unexpected ($response)"
    fi
    
    sleep 0.05
done

echo ""
echo "Results: $success_count successful, $blocked_count blocked"
echo ""
echo "=========================================="

# Test 4: Rapid requests to trigger blocking
echo ""
echo "Test 4: Rapid requests (no delay) - should trigger blocking"
echo "Sending 15 rapid requests..."
echo ""

success_count=0
blocked_count=0

for i in $(seq 1 15); do
    response=$(curl -s -o /dev/null -w "%{http_code}" $BASE_URL)
    
    if [ "$response" == "200" ]; then
        echo "Request $i: ✓ Success (200)"
        ((success_count++))
    elif [ "$response" == "429" ]; then
        echo "Request $i: ✗ Blocked (429)"
        ((blocked_count++))
    else
        echo "Request $i: ? Unexpected ($response)"
    fi
done

echo ""
echo "Results: $success_count successful, $blocked_count blocked"
echo ""
echo "=========================================="
echo "Load test completed!"
echo "=========================================="
