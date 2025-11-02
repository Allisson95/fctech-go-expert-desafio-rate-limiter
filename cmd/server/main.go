package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/allis/rate-limiter/internal/config"
	"github.com/allis/rate-limiter/internal/limiter"
	"github.com/allis/rate-limiter/internal/middleware"
	"github.com/allis/rate-limiter/internal/storage"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Redis storage
	redisStorage, err := storage.NewRedisStorage(
		cfg.Redis.Addr,
		cfg.Redis.Password,
		cfg.Redis.DB,
	)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisStorage.Close()

	log.Println("Connected to Redis successfully")

	// Create rate limiter
	rateLimiter := limiter.NewRateLimiter(redisStorage, limiter.Config{
		IPLimit:                   cfg.RateLimiter.IPLimit,
		IPBlockDuration:           cfg.RateLimiter.IPBlockDuration,
		TokenLimits:               cfg.RateLimiter.TokenLimits,
		DefaultTokenLimit:         cfg.RateLimiter.DefaultTokenLimit,
		DefaultTokenBlockDuration: cfg.RateLimiter.DefaultTokenBlockDuration,
	})

	// Create HTTP server with rate limiter middleware
	mux := http.NewServeMux()

	// Add a simple handler for testing
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Request successful!"))
	})

	// Add health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Apply rate limiter middleware
	handler := middleware.RateLimiterMiddleware(rateLimiter)(mux)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Starting server on %s", addr)
	log.Printf("Rate Limiter Config:")
	log.Printf("  - IP Limit: %d req/s", cfg.RateLimiter.IPLimit)
	log.Printf("  - IP Block Duration: %v", cfg.RateLimiter.IPBlockDuration)
	log.Printf("  - Default Token Limit: %d req/s", cfg.RateLimiter.DefaultTokenLimit)
	log.Printf("  - Default Token Block Duration: %v", cfg.RateLimiter.DefaultTokenBlockDuration)
	log.Printf("  - Custom Token Limits: %d configured", len(cfg.RateLimiter.TokenLimits))

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
