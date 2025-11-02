package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/allis/rate-limiter/internal/limiter"
)

const (
	apiKeyHeader     = "API_KEY"
	rateLimitMessage = "you have reached the maximum number of requests or actions allowed within a certain time frame"
)

// RateLimiterMiddleware creates a middleware that applies rate limiting
func RateLimiterMiddleware(rateLimiter *limiter.RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.Background()

			// Check for API_KEY in header
			token := r.Header.Get(apiKeyHeader)

			// If token is present, use token-based rate limiting
			if token != "" {
				allowed, err := rateLimiter.AllowToken(ctx, token)
				if err != nil {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
					return
				}

				if !allowed {
					w.WriteHeader(http.StatusTooManyRequests)
					w.Write([]byte(rateLimitMessage))
					return
				}

				// Token is allowed, proceed with request
				next.ServeHTTP(w, r)
				return
			}

			// No token, use IP-based rate limiting
			ip := getIP(r)
			if ip == "" {
				http.Error(w, "Unable to determine IP address", http.StatusBadRequest)
				return
			}

			allowed, err := rateLimiter.AllowIP(ctx, ip)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if !allowed {
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(rateLimitMessage))
				return
			}

			// Request is allowed, proceed
			next.ServeHTTP(w, r)
		})
	}
}

// getIP extracts the IP address from the request
func getIP(r *http.Request) string {
	// Check X-Forwarded-For header
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// Take the first IP in the list
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Use RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}
