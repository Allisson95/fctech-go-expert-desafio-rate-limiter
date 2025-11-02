package limiter

import (
	"context"
	"testing"
	"time"
)

// MockStorage is a mock implementation of Storage for testing
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

func (m *MockStorage) Reset() {
	m.counters = make(map[string]int64)
	m.blocks = make(map[string]time.Time)
}

func TestRateLimiter_AllowIP(t *testing.T) {
	storage := NewMockStorage()
	config := Config{
		IPLimit:         5,
		IPBlockDuration: 5 * time.Second,
	}
	rl := NewRateLimiter(storage, config)
	ctx := context.Background()

	// Test: Allow requests within limit
	for i := 1; i <= 5; i++ {
		allowed, err := rl.AllowIP(ctx, "192.168.1.1")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !allowed {
			t.Fatalf("Request %d should be allowed", i)
		}
	}

	// Test: Block request exceeding limit
	allowed, err := rl.AllowIP(ctx, "192.168.1.1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if allowed {
		t.Fatal("Request should be blocked after exceeding limit")
	}

	// Test: Subsequent requests should also be blocked
	allowed, err = rl.AllowIP(ctx, "192.168.1.1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if allowed {
		t.Fatal("Request should remain blocked")
	}

	// Test: Different IP should not be affected
	storage.Reset()
	allowed, err = rl.AllowIP(ctx, "192.168.1.2")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !allowed {
		t.Fatal("Different IP should be allowed")
	}
}

func TestRateLimiter_AllowToken(t *testing.T) {
	storage := NewMockStorage()
	config := Config{
		DefaultTokenLimit:         10,
		DefaultTokenBlockDuration: 5 * time.Second,
		TokenLimits: map[string]TokenConfig{
			"premium": {
				Limit:         20,
				BlockDuration: 10 * time.Second,
			},
		},
	}
	rl := NewRateLimiter(storage, config)
	ctx := context.Background()

	// Test: Allow requests within default limit
	for i := 1; i <= 10; i++ {
		allowed, err := rl.AllowToken(ctx, "standard")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !allowed {
			t.Fatalf("Request %d should be allowed", i)
		}
	}

	// Test: Block request exceeding default limit
	allowed, err := rl.AllowToken(ctx, "standard")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if allowed {
		t.Fatal("Request should be blocked after exceeding default limit")
	}

	// Test: Premium token with custom limit
	storage.Reset()
	for i := 1; i <= 20; i++ {
		allowed, err := rl.AllowToken(ctx, "premium")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !allowed {
			t.Fatalf("Premium request %d should be allowed", i)
		}
	}

	// Test: Block premium token after exceeding custom limit
	allowed, err = rl.AllowToken(ctx, "premium")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if allowed {
		t.Fatal("Premium request should be blocked after exceeding custom limit")
	}
}

func TestRateLimiter_TokenOverridesIP(t *testing.T) {
	storage := NewMockStorage()
	config := Config{
		IPLimit:                   5,
		IPBlockDuration:           5 * time.Second,
		DefaultTokenLimit:         15,
		DefaultTokenBlockDuration: 5 * time.Second,
	}
	rl := NewRateLimiter(storage, config)
	ctx := context.Background()

	// Simulate same client using token (should use token limit, not IP limit)
	// In practice, middleware checks token first

	// Test: Token limit is higher than IP limit
	for i := 1; i <= 15; i++ {
		allowed, err := rl.AllowToken(ctx, "test_token")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !allowed {
			t.Fatalf("Request %d with token should be allowed", i)
		}
	}

	// Test: Should block after token limit, not IP limit
	allowed, err := rl.AllowToken(ctx, "test_token")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if allowed {
		t.Fatal("Request should be blocked after exceeding token limit")
	}
}
