# Rate Limiter em Go

Um rate limiter robusto e configurável implementado em Go, capaz de limitar requisições HTTP por endereço IP ou token de acesso, com suporte a Redis para persistência distribuída.

## 📋 Descrição

Este projeto implementa um rate limiter que pode ser usado como middleware em servidores web Go. Ele oferece:

- **Limitação por IP**: Restringe o número de requisições de um endereço IP específico
- **Limitação por Token**: Permite diferentes limites para diferentes tokens de acesso
- **Prioridade de Token**: Configurações de token sobrepõem as de IP
- **Bloqueio Temporário**: IPs/tokens que excedem o limite são bloqueados por um período configurável
- **Strategy Pattern**: Interface de storage que permite trocar Redis por outros mecanismos
- **Separação de Responsabilidades**: Lógica do limiter separada do middleware

## 🏗️ Arquitetura

```
rate-limiter/
├── cmd/
│   └── server/
│       └── main.go              # Ponto de entrada da aplicação
├── internal/
│   ├── config/
│   │   └── config.go            # Carregamento de configurações
│   ├── limiter/
│   │   ├── limiter.go           # Lógica do rate limiter
│   │   └── limiter_test.go      # Testes unitários
│   ├── middleware/
│   │   ├── ratelimiter.go       # Middleware HTTP
│   │   └── ratelimiter_test.go  # Testes do middleware
│   └── storage/
│       ├── storage.go           # Interface Storage (Strategy Pattern)
│       └── redis.go             # Implementação Redis
├── .env                         # Variáveis de ambiente
├── docker-compose.yml           # Orquestração de containers
├── Dockerfile                   # Imagem Docker da aplicação
├── Makefile                     # Comandos úteis
└── README.md                    # Esta documentação
```

## 🚀 Funcionalidades

### 1. Rate Limiting por IP

- Limita requisições baseadas no endereço IP do cliente
- Suporta detecção de IP via `X-Forwarded-For`, `X-Real-IP` ou `RemoteAddr`
- Configurável via variáveis de ambiente

### 2. Rate Limiting por Token

- Token informado no header `API_KEY`
- Cada token pode ter limites e durações de bloqueio personalizados
- Tokens não configurados usam limite padrão
- **Token sobrepõe IP**: Se presente, usa configuração do token

### 3. Bloqueio Temporário

- Quando o limite é excedido, o IP/token é bloqueado
- Período de bloqueio configurável (padrão: 5 minutos)
- Durante o bloqueio, todas as requisições retornam HTTP 429

### 4. Strategy Pattern

A interface `Storage` permite trocar facilmente o mecanismo de persistência:

```go
type Storage interface {
    Increment(ctx context.Context, key string, expiration time.Duration) (int64, error)
    Get(ctx context.Context, key string) (int64, error)
    SetBlock(ctx context.Context, key string, duration time.Duration) error
    IsBlocked(ctx context.Context, key string) (bool, error)
    TTL(ctx context.Context, key string) (time.Duration, error)
    Close() error
}
```

Implementações disponíveis:
- `RedisStorage`: Usa Redis para armazenamento distribuído
- `MockStorage`: Implementação em memória para testes

## ⚙️ Configuração

### Variáveis de Ambiente

Crie um arquivo `.env` na raiz do projeto:

```env
# Redis Configuration
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# IP Rate Limiter Configuration
IP_RATE_LIMIT=10                # Máximo de requisições por segundo por IP
IP_BLOCK_DURATION=300           # Tempo de bloqueio em segundos (5 minutos)

# Token Rate Limiter Configuration
TOKEN_RATE_LIMIT=100            # Limite padrão para tokens
TOKEN_BLOCK_DURATION=300        # Tempo de bloqueio padrão para tokens

# Custom Token Configuration (format: TOKEN:LIMIT:BLOCK_DURATION)
API_KEY_abc123=100:300          # Token 'abc123' com 100 req/s e 300s de bloqueio
API_KEY_xyz789=200:600          # Token 'xyz789' com 200 req/s e 600s de bloqueio

# Server Configuration
SERVER_PORT=8080
```

### Configuração de Tokens Personalizados

Tokens personalizados seguem o formato:
```
API_KEY_<nome_do_token>=<limite>:<duracao_bloqueio_segundos>
```

Exemplo:
```env
API_KEY_premium=1000:60      # 1000 req/s, bloqueio de 1 minuto
API_KEY_basic=50:600         # 50 req/s, bloqueio de 10 minutos
```

## 🐳 Docker e Docker Compose

### Subir a aplicação com Docker Compose

```bash
# Construir e iniciar os serviços
docker-compose up -d

# Ver logs
docker-compose logs -f

# Parar os serviços
docker-compose down
```

Ou usando o Makefile:

```bash
make docker-up      # Inicia os serviços
make docker-logs    # Visualiza logs
make docker-down    # Para os serviços
```

### Serviços

- **Redis**: `localhost:6379`
- **Aplicação**: `http://localhost:8080`

## 🧪 Testes

### Testes Unitários

```bash
# Executar todos os testes
go test ./...

# Executar com verbose
go test -v ./...

# Executar com cobertura
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

Ou usando o Makefile:

```bash
make test              # Testes básicos
make test-coverage     # Testes com relatório de cobertura
```

### Testes de Carga

Execute o script de teste de carga:

```bash
chmod +x test_load.sh
./test_load.sh
```

Ou use o Makefile:

```bash
make load-test
```

### Testes Manuais com curl

```bash
# Teste limitação por IP
make curl-test-ip

# Teste limitação por token
make curl-test-token

# Teste token premium
make curl-test-premium

# Health check
make health-check
```

## 📊 Exemplos de Uso

### Exemplo 1: Limitação por IP

```bash
# Configuração: 10 req/s por IP

# Requisições 1-10: Sucesso (200 OK)
curl http://localhost:8080/
# Response: "Request successful!"

# Requisição 11: Bloqueada (429 Too Many Requests)
curl http://localhost:8080/
# Response: "you have reached the maximum number of requests or actions allowed within a certain time frame"
```

### Exemplo 2: Limitação por Token

```bash
# Token com limite de 100 req/s

# Com token - mais requisições permitidas
curl -H "API_KEY: abc123" http://localhost:8080/

# Sem token - limite por IP (10 req/s)
curl http://localhost:8080/
```

### Exemplo 3: Token Sobrepõe IP

```bash
# IP limitado a 10 req/s, token 'premium' a 1000 req/s

# Requisições com token usam limite do token (1000 req/s)
for i in {1..50}; do
  curl -H "API_KEY: premium" http://localhost:8080/
done
# Todas bem-sucedidas

# Requisições sem token usam limite por IP (10 req/s)
for i in {1..15}; do
  curl http://localhost:8080/
done
# Primeiras 10 bem-sucedidas, demais bloqueadas
```

## 🔧 Desenvolvimento

### Executar localmente (sem Docker)

```bash
# Instalar dependências
go mod download

# Subir Redis separadamente
docker run -d -p 6379:6379 redis:7-alpine

# Executar aplicação
go run cmd/server/main.go
```

### Comandos úteis do Makefile

```bash
make help              # Lista todos os comandos disponíveis
make build             # Compila a aplicação
make run               # Executa localmente
make test              # Executa testes
make docker-up         # Sobe com Docker Compose
make docker-logs       # Visualiza logs
make redis-cli         # Conecta ao Redis CLI
make redis-monitor     # Monitora comandos Redis
make redis-flush       # Limpa banco Redis
```

## 🔍 Monitoramento Redis

### Conectar ao Redis CLI

```bash
make redis-cli
```

### Visualizar chaves e valores

```bash
# No Redis CLI
KEYS *                 # Lista todas as chaves
GET counter:ip:192.168.1.1    # Valor do contador
GET block:ip:192.168.1.1      # Status de bloqueio
TTL block:ip:192.168.1.1      # Tempo restante de bloqueio
```

### Monitorar em tempo real

```bash
make redis-monitor
```

## 📝 Resposta HTTP 429

Quando o limite é excedido, a aplicação retorna:

- **Código HTTP**: `429 Too Many Requests`
- **Mensagem**: `you have reached the maximum number of requests or actions allowed within a certain time frame`

## 🧩 Extensibilidade

### Adicionar nova implementação de Storage

1. Implemente a interface `Storage`:

```go
type MyStorage struct {
    // seus campos
}

func (s *MyStorage) Increment(ctx context.Context, key string, expiration time.Duration) (int64, error) {
    // sua implementação
}

// Implemente os demais métodos...
```

2. Use no `main.go`:

```go
myStorage := NewMyStorage()
rateLimiter := limiter.NewRateLimiter(myStorage, config)
```

### Personalizar middleware

O middleware pode ser customizado para adicionar logs, métricas, etc:

```go
func CustomRateLimiterMiddleware(rateLimiter *limiter.RateLimiter, logger *log.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Adicione sua lógica personalizada aqui
            logger.Printf("Request from %s", r.RemoteAddr)
            
            // Chame o middleware original
            middleware.RateLimiterMiddleware(rateLimiter)(next).ServeHTTP(w, r)
        })
    }
}
```

## 🐛 Troubleshooting

### Problema: Não conecta ao Redis

```bash
# Verifique se Redis está rodando
docker ps | grep redis

# Verifique logs do Redis
docker logs rate-limiter-redis

# Teste conexão
docker exec -it rate-limiter-redis redis-cli ping
```

### Problema: Todas as requisições são bloqueadas

```bash
# Limpe o banco Redis
make redis-flush

# Ou reinicie os serviços
make docker-restart
```

### Problema: Limite não está sendo aplicado

- Verifique as variáveis de ambiente no `.env`
- Confirme que o Docker Compose está usando as variáveis corretas
- Verifique logs da aplicação: `make docker-logs`

## 📚 Tecnologias Utilizadas

- **Go 1.21**: Linguagem de programação
- **Redis 7**: Armazenamento de dados
- **go-redis/redis/v8**: Cliente Redis para Go
- **godotenv**: Carregamento de variáveis de ambiente
- **Docker & Docker Compose**: Containerização

## 📄 Licença

Este projeto foi desenvolvido como desafio técnico para o curso Pós Go Expert da Full Cycle.

## 👤 Autor

Desenvolvido como parte do desafio técnico de Rate Limiter.
