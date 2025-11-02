package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/allis/rate-limiter/internal/limiter"
)

// MockStorage for testing
type MockStorage struct {
	counters map[string]int64
	blocks   map[string]time.Time
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		counters: make(map[string]int64),
		blocks:   make(map[string]time.Time),
	}
}

func (m *MockStorage) Increment(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	m.counters[key]++
	return m.counters[key], nil
}

func (m *MockStorage) Get(ctx context.Context, key string) (int64, error) {
	return m.counters[key], nil
}

func (m *MockStorage) SetBlock(ctx context.Context, key string, duration time.Duration) error {
	m.blocks[key] = time.Now().Add(duration)
	return nil
}

func (m *MockStorage) IsBlocked(ctx context.Context, key string) (bool, error) {
	if blockUntil, exists := m.blocks[key]; exists {
		return time.Now().Before(blockUntil), nil
	}
	return false, nil
}

func (m *MockStorage) TTL(ctx context.Context, key string) (time.Duration, error) {
	if blockUntil, exists := m.blocks[key]; exists {
		return time.Until(blockUntil), nil
	}
	return 0, nil
}

func (m *MockStorage) Close() error {
	return nil
}

func TestRateLimiterMiddleware_IP(t *testing.T) {
	storage := NewMockStorage()
	config := limiter.Config{
		IPLimit:         3,
		IPBlockDuration: 5 * time.Second,
	}
	rl := limiter.NewRateLimiter(storage, config)

	handler := RateLimiterMiddleware(rl)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Test: Allow requests within limit
	for i := 1; i <= 3; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Request %d: expected status 200, got %d", i, w.Code)
		}
	}

	// Test: Block request exceeding limit
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("Expected status 429, got %d", w.Code)
	}

	expectedMessage := "you have reached the maximum number of requests or actions allowed within a certain time frame"
	if w.Body.String() != expectedMessage {
		t.Fatalf("Expected message '%s', got '%s'", expectedMessage, w.Body.String())
	}
}

func TestRateLimiterMiddleware_Token(t *testing.T) {
	storage := NewMockStorage()
	config := limiter.Config{
		IPLimit:                   2,
		IPBlockDuration:           5 * time.Second,
		DefaultTokenLimit:         5,
		DefaultTokenBlockDuration: 5 * time.Second,
	}
	rl := limiter.NewRateLimiter(storage, config)

	handler := RateLimiterMiddleware(rl)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Test: Token-based limiting (higher limit than IP)
	for i := 1; i <= 5; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		req.Header.Set("API_KEY", "test_token")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Request %d with token: expected status 200, got %d", i, w.Code)
		}
	}

	// Test: Block token after exceeding limit
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	req.Header.Set("API_KEY", "test_token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("Expected status 429 for token, got %d", w.Code)
	}
}

func TestRateLimiterMiddleware_XForwardedFor(t *testing.T) {
	storage := NewMockStorage()
	config := limiter.Config{
		IPLimit:         2,
		IPBlockDuration: 5 * time.Second,
	}
	rl := limiter.NewRateLimiter(storage, config)

	handler := RateLimiterMiddleware(rl)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Test: Use X-Forwarded-For header
	for i := 1; i <= 2; i++ {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Request %d: expected status 200, got %d", i, w.Code)
		}
	}

	// Test: Should be blocked based on X-Forwarded-For IP
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("Expected status 429, got %d", w.Code)
	}
}
