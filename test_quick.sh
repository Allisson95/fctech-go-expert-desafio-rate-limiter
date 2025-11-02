#!/bin/bash

# Script para testar a aplicação rapidamente

echo "==========================================="
echo "Rate Limiter - Quick Test"
echo "==========================================="
echo ""

# Verifica se a aplicação está rodando
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "❌ Error: Application is not running on port 8080"
    echo "Please start the application with: docker-compose up -d"
    echo "Or: make docker-up"
    exit 1
fi

echo "✓ Application is running"
echo ""

# Test 1: Health Check
echo "Test 1: Health Check"
response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health)
if [ "$response" == "200" ]; then
    echo "✓ Health check passed"
else
    echo "✗ Health check failed (Status: $response)"
fi
echo ""

# Test 2: Normal request
echo "Test 2: Normal Request (should succeed)"
response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/)
if [ "$response" == "200" ]; then
    echo "✓ Normal request succeeded (Status: 200)"
else
    echo "✗ Normal request failed (Status: $response)"
fi
echo ""

# Test 3: Multiple rapid requests (should trigger rate limit)
echo "Test 3: Rapid requests (testing rate limit)"
success=0
blocked=0

for i in {1..15}; do
    response=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/)
    if [ "$response" == "200" ]; then
        ((success++))
    elif [ "$response" == "429" ]; then
        ((blocked++))
    fi
done

echo "Results: $success successful, $blocked blocked"
if [ $blocked -gt 0 ]; then
    echo "✓ Rate limiting is working (some requests were blocked)"
else
    echo "⚠ Warning: No requests were blocked. Rate limit may be too high."
fi
echo ""

# Test 4: Request with token
echo "Test 4: Request with API_KEY (should have higher limit)"
response=$(curl -s -o /dev/null -w "%{http_code}" -H "API_KEY: abc123" http://localhost:8080/)
if [ "$response" == "200" ]; then
    echo "✓ Token-based request succeeded (Status: 200)"
else
    echo "✗ Token-based request failed (Status: $response)"
fi
echo ""

echo "==========================================="
echo "Quick test completed!"
echo "==========================================="
echo ""
echo "For more detailed tests, run:"
echo "  make load-test          # Full load testing"
echo "  make curl-test-ip       # Test IP limiting"
echo "  make curl-test-token    # Test token limiting"
