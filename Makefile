.PHONY: help build run test clean docker-up docker-down docker-build docker-logs

help: ## Display this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the Go application
	go build -o bin/server ./cmd/server

run: ## Run the application locally
	go run ./cmd/server/main.go

test: ## Run unit tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

mod: ## Download Go modules
	go mod download
	go mod tidy

docker-build: ## Build Docker image
	docker-compose build

docker-up: ## Start services with docker-compose
	docker-compose up -d

docker-down: ## Stop services
	docker-compose down

docker-logs: ## View logs from docker-compose
	docker-compose logs -f

docker-restart: docker-down docker-up ## Restart docker services

load-test: ## Run load test (requires service to be running)
	chmod +x test_load.sh
	./test_load.sh

curl-test-ip: ## Test with IP limiting
	@echo "Testing IP-based rate limiting..."
	@for i in 1 2 3 4 5 6 7 8 9 10 11 12; do \
		echo "Request $$i:"; \
		curl -w "\nStatus: %{http_code}\n\n" http://localhost:8080/; \
	done

curl-test-token: ## Test with token limiting
	@echo "Testing token-based rate limiting..."
	@for i in 1 2 3 4 5 6 7 8 9 10 11 12; do \
		echo "Request $$i:"; \
		curl -w "\nStatus: %{http_code}\n\n" -H "API_KEY: test_token" http://localhost:8080/; \
	done

curl-test-premium: ## Test with premium token (abc123)
	@echo "Testing premium token rate limiting..."
	@for i in 1 2 3 4 5; do \
		echo "Request $$i:"; \
		curl -w "\nStatus: %{http_code}\n\n" -H "API_KEY: abc123" http://localhost:8080/; \
	done

health-check: ## Check service health
	@curl http://localhost:8080/health

redis-cli: ## Connect to Redis CLI
	docker exec -it rate-limiter-redis redis-cli

redis-monitor: ## Monitor Redis commands
	docker exec -it rate-limiter-redis redis-cli MONITOR

redis-flush: ## Flush Redis database
	docker exec -it rate-limiter-redis redis-cli FLUSHALL
	@echo "Redis database flushed"
