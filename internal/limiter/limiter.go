package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/allis/rate-limiter/internal/storage"
)

// Config holds the configuration for rate limiter
type Config struct {
	IPLimit                   int
	IPBlockDuration           time.Duration
	TokenLimits               map[string]TokenConfig
	DefaultTokenLimit         int
	DefaultTokenBlockDuration time.Duration
}

// TokenConfig holds token-specific configuration
type TokenConfig struct {
	Limit         int
	BlockDuration time.Duration
}

// RateLimiter handles rate limiting logic
type RateLimiter struct {
	storage storage.Storage
	config  Config
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter(storage storage.Storage, config Config) *RateLimiter {
	return &RateLimiter{
		storage: storage,
		config:  config,
	}
}

// AllowIP checks if a request from an IP is allowed
func (rl *RateLimiter) AllowIP(ctx context.Context, ip string) (bool, error) {
	key := fmt.Sprintf("ip:%s", ip)

	// Check if IP is blocked
	blocked, err := rl.storage.IsBlocked(ctx, key)
	if err != nil {
		return false, fmt.Errorf("failed to check if IP is blocked: %w", err)
	}
	if blocked {
		return false, nil
	}

	// Increment counter
	count, err := rl.storage.Increment(ctx, key, time.Second)
	if err != nil {
		return false, fmt.Errorf("failed to increment IP counter: %w", err)
	}

	// Check if limit exceeded
	if count > int64(rl.config.IPLimit) {
		// Block the IP
		if err := rl.storage.SetBlock(ctx, key, rl.config.IPBlockDuration); err != nil {
			return false, fmt.Errorf("failed to block IP: %w", err)
		}
		return false, nil
	}

	return true, nil
}

// AllowToken checks if a request with a token is allowed
func (rl *RateLimiter) AllowToken(ctx context.Context, token string) (bool, error) {
	key := fmt.Sprintf("token:%s", token)

	// Check if token is blocked
	blocked, err := rl.storage.IsBlocked(ctx, key)
	if err != nil {
		return false, fmt.Errorf("failed to check if token is blocked: %w", err)
	}
	if blocked {
		return false, nil
	}

	// Get token configuration
	tokenConfig, exists := rl.config.TokenLimits[token]
	if !exists {
		// Use default token configuration
		tokenConfig = TokenConfig{
			Limit:         rl.config.DefaultTokenLimit,
			BlockDuration: rl.config.DefaultTokenBlockDuration,
		}
	}

	// Increment counter
	count, err := rl.storage.Increment(ctx, key, time.Second)
	if err != nil {
		return false, fmt.Errorf("failed to increment token counter: %w", err)
	}

	// Check if limit exceeded
	if count > int64(tokenConfig.Limit) {
		// Block the token
		if err := rl.storage.SetBlock(ctx, key, tokenConfig.BlockDuration); err != nil {
			return false, fmt.Errorf("failed to block token: %w", err)
		}
		return false, nil
	}

	return true, nil
}

// GetBlockTTL returns the remaining block duration for a key
func (rl *RateLimiter) GetBlockTTL(ctx context.Context, identifier string) (time.Duration, error) {
	return rl.storage.TTL(ctx, identifier)
}
