package storage

import (
	"context"
	"time"
)

// Storage defines the interface for rate limiter storage
type Storage interface {
	// Increment increments the counter for a key and returns the new value
	// If the key doesn't exist, it creates it with value 1 and sets the expiration
	Increment(ctx context.Context, key string, expiration time.Duration) (int64, error)

	// Get returns the current counter value for a key
	Get(ctx context.Context, key string) (int64, error)

	// SetBlock sets a block for a key with the specified duration
	SetBlock(ctx context.Context, key string, duration time.Duration) error

	// IsBlocked checks if a key is currently blocked
	IsBlocked(ctx context.Context, key string) (bool, error)

	// TTL returns the time to live for a key
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Close closes the storage connection
	Close() error
}
