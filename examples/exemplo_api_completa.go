package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/allis/rate-limiter/internal/limiter"
	"github.com/allis/rate-limiter/internal/middleware"
	"github.com/allis/rate-limiter/internal/storage"
)

// User representa um usuário na API
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// APIResponse representa uma resposta padrão da API
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

var users = []User{
	{ID: 1, Name: "João Silva", Email: "joao@example.com", CreatedAt: time.Now()},
	{ID: 2, Name: "Maria Santos", Email: "maria@example.com", CreatedAt: time.Now()},
}

func main() {
	// Configurar Redis Storage
	redisStorage, err := storage.NewRedisStorage("localhost:6379", "", 0)
	if err != nil {
		log.Fatalf("Erro ao conectar ao Redis: %v", err)
	}
	defer redisStorage.Close()

	// Configurar Rate Limiter
	config := limiter.Config{
		IPLimit:                   5, // 5 requisições por segundo por IP
		IPBlockDuration:           30 * time.Second,
		DefaultTokenLimit:         20, // 20 requisições por segundo para tokens
		DefaultTokenBlockDuration: 60 * time.Second,
		TokenLimits: map[string]limiter.TokenConfig{
			"premium_token": {
				Limit:         100,
				BlockDuration: 30 * time.Second,
			},
			"basic_token": {
				Limit:         10,
				BlockDuration: 60 * time.Second,
			},
		},
	}

	rateLimiter := limiter.NewRateLimiter(redisStorage, config)

	// Configurar rotas
	mux := http.NewServeMux()

	// Health check (sem rate limiting)
	mux.HandleFunc("/health", handleHealth)

	// API routes (com rate limiting)
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/api/users", handleUsers)
	apiMux.HandleFunc("/api/stats", handleStats)

	// Aplicar middleware de rate limiting apenas nas rotas /api/*
	mux.Handle("/api/", middleware.RateLimiterMiddleware(rateLimiter)(apiMux))

	// Iniciar servidor
	addr := ":8080"
	fmt.Printf("🚀 Servidor iniciado em http://localhost%s\n", addr)
	fmt.Println("📝 Endpoints disponíveis:")
	fmt.Println("   GET  /health        - Health check")
	fmt.Println("   GET  /api/users     - Lista usuários (rate limited)")
	fmt.Println("   POST /api/users     - Cria usuário (rate limited)")
	fmt.Println("   GET  /api/stats     - Estatísticas (rate limited)")
	fmt.Println()
	fmt.Println("🔑 Tokens configurados:")
	fmt.Println("   premium_token - 100 req/s")
	fmt.Println("   basic_token   - 10 req/s")
	fmt.Println()
	fmt.Println("⚡ Rate Limits:")
	fmt.Printf("   IP: %d req/s (bloqueio: %v)\n", config.IPLimit, config.IPBlockDuration)
	fmt.Printf("   Token padrão: %d req/s (bloqueio: %v)\n", config.DefaultTokenLimit, config.DefaultTokenBlockDuration)
	fmt.Println()

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Erro ao iniciar servidor: %v", err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "Service is healthy",
	})
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Listar usuários
		respondJSON(w, http.StatusOK, APIResponse{
			Success: true,
			Data:    users,
		})

	case http.MethodPost:
		// Criar novo usuário
		var newUser User
		if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
			respondJSON(w, http.StatusBadRequest, APIResponse{
				Success: false,
				Message: "Invalid request body",
			})
			return
		}

		newUser.ID = len(users) + 1
		newUser.CreatedAt = time.Now()
		users = append(users, newUser)

		respondJSON(w, http.StatusCreated, APIResponse{
			Success: true,
			Data:    newUser,
			Message: "User created successfully",
		})

	default:
		respondJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Success: false,
			Message: "Method not allowed",
		})
	}
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondJSON(w, http.StatusMethodNotAllowed, APIResponse{
			Success: false,
			Message: "Method not allowed",
		})
		return
	}

	stats := map[string]interface{}{
		"total_users":  len(users),
		"server_time":  time.Now(),
		"api_version":  "1.0.0",
		"rate_limited": true,
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    stats,
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
