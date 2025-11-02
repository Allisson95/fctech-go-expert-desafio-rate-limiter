package storage

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

// TestRedisStorage_Integration tests Redis storage with a real Redis instance
// Skip this test if Redis is not available
func TestRedisStorage_Integration(t *testing.T) {
	// Check if Redis is available
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Skip("Skipping integration test: Redis not available")
	}
	client.Close()

	// Create Redis storage
	storage, err := NewRedisStorage("localhost:6379", "", 0)
	if err != nil {
		t.Fatalf("Failed to create Redis storage: %v", err)
	}
	defer storage.Close()

	testKey := "test:integration:key"

	// Clean up before and after test
	ctx = context.Background()
	defer func() {
		storage.client.Del(ctx, "counter:"+testKey)
		storage.client.Del(ctx, "block:"+testKey)
	}()

	// Test Increment
	count, err := storage.Increment(ctx, testKey, time.Second)
	if err != nil {
		t.Fatalf("Increment failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}

	count, err = storage.Increment(ctx, testKey, time.Second)
	if err != nil {
		t.Fatalf("Increment failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}

	// Test Get
	value, err := storage.Get(ctx, testKey)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if value != 2 {
		t.Errorf("Expected value 2, got %d", value)
	}

	// Test SetBlock
	if err := storage.SetBlock(ctx, testKey, 2*time.Second); err != nil {
		t.Fatalf("SetBlock failed: %v", err)
	}

	// Test IsBlocked
	blocked, err := storage.IsBlocked(ctx, testKey)
	if err != nil {
		t.Fatalf("IsBlocked failed: %v", err)
	}
	if !blocked {
		t.Error("Expected key to be blocked")
	}

	// Test TTL
	ttl, err := storage.TTL(ctx, testKey)
	if err != nil {
		t.Fatalf("TTL failed: %v", err)
	}
	if ttl <= 0 || ttl > 2*time.Second {
		t.Errorf("Expected TTL between 0 and 2s, got %v", ttl)
	}

	// Wait for block to expire
	time.Sleep(3 * time.Second)

	blocked, err = storage.IsBlocked(ctx, testKey)
	if err != nil {
		t.Fatalf("IsBlocked failed: %v", err)
	}
	if blocked {
		t.Error("Expected key to not be blocked after expiration")
	}
}
