package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/allis/rate-limiter/internal/limiter"
	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	Redis       RedisConfig
	RateLimiter RateLimiterConfig
	Server      ServerConfig
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// RateLimiterConfig holds rate limiter configuration
type RateLimiterConfig struct {
	IPLimit                   int
	IPBlockDuration           time.Duration
	DefaultTokenLimit         int
	DefaultTokenBlockDuration time.Duration
	TokenLimits               map[string]limiter.TokenConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	cfg := &Config{
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		RateLimiter: RateLimiterConfig{
			IPLimit:                   getEnvAsInt("IP_RATE_LIMIT", 10),
			IPBlockDuration:           time.Duration(getEnvAsInt("IP_BLOCK_DURATION", 300)) * time.Second,
			DefaultTokenLimit:         getEnvAsInt("TOKEN_RATE_LIMIT", 100),
			DefaultTokenBlockDuration: time.Duration(getEnvAsInt("TOKEN_BLOCK_DURATION", 300)) * time.Second,
			TokenLimits:               make(map[string]limiter.TokenConfig),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
	}

	// Load token-specific configurations
	loadTokenConfigs(cfg)

	return cfg, nil
}

// loadTokenConfigs loads token-specific rate limit configurations
func loadTokenConfigs(cfg *Config) {
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "API_KEY_") {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) != 2 {
				continue
			}

			// Extract token name (remove API_KEY_ prefix)
			token := strings.TrimPrefix(parts[0], "API_KEY_")

			// Parse value: format is "LIMIT:BLOCK_DURATION"
			valueParts := strings.Split(parts[1], ":")
			if len(valueParts) != 2 {
				continue
			}

			limit, err := strconv.Atoi(valueParts[0])
			if err != nil {
				continue
			}

			blockDuration, err := strconv.Atoi(valueParts[1])
			if err != nil {
				continue
			}

			cfg.RateLimiter.TokenLimits[token] = limiter.TokenConfig{
				Limit:         limit,
				BlockDuration: time.Duration(blockDuration) * time.Second,
			}
		}
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as int with a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		fmt.Printf("Warning: invalid value for %s, using default %d\n", key, defaultValue)
		return defaultValue
	}

	return value
}
