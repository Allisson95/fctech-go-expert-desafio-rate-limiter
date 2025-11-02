package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	blockPrefix   = "block:"
	counterPrefix = "counter:"
)

// RedisStorage implements Storage interface using Redis
type RedisStorage struct {
	client *redis.Client
}

// NewRedisStorage creates a new Redis storage instance
func NewRedisStorage(addr, password string, db int) (*RedisStorage, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStorage{
		client: client,
	}, nil
}

// Increment increments the counter for a key
func (r *RedisStorage) Increment(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	counterKey := counterPrefix + key

	pipe := r.client.Pipeline()
	incr := pipe.Incr(ctx, counterKey)
	pipe.Expire(ctx, counterKey, expiration)

	if _, err := pipe.Exec(ctx); err != nil {
		return 0, fmt.Errorf("failed to increment counter: %w", err)
	}

	return incr.Val(), nil
}

// Get returns the current counter value for a key
func (r *RedisStorage) Get(ctx context.Context, key string) (int64, error) {
	counterKey := counterPrefix + key
	val, err := r.client.Get(ctx, counterKey).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get counter: %w", err)
	}
	return val, nil
}

// SetBlock sets a block for a key
func (r *RedisStorage) SetBlock(ctx context.Context, key string, duration time.Duration) error {
	blockKey := blockPrefix + key
	if err := r.client.Set(ctx, blockKey, "1", duration).Err(); err != nil {
		return fmt.Errorf("failed to set block: %w", err)
	}
	return nil
}

// IsBlocked checks if a key is blocked
func (r *RedisStorage) IsBlocked(ctx context.Context, key string) (bool, error) {
	blockKey := blockPrefix + key
	val, err := r.client.Get(ctx, blockKey).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check block: %w", err)
	}
	return val == "1", nil
}

// TTL returns the time to live for a key
func (r *RedisStorage) TTL(ctx context.Context, key string) (time.Duration, error) {
	blockKey := blockPrefix + key
	ttl, err := r.client.TTL(ctx, blockKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}
	return ttl, nil
}

// Close closes the Redis connection
func (r *RedisStorage) Close() error {
	return r.client.Close()
}
