# Examples

Este diretório contém exemplos de uso do Rate Limiter.

## exemplo_api_completa.go

Um exemplo mais completo de API que usa o rate limiter com múltiplos endpoints e funcionalidades adicionais.

### Executar o exemplo

```bash
# Certifique-se de que o Redis está rodando
docker run -d -p 6379:6379 redis:7-alpine

# Execute o exemplo
go run examples/exemplo_api_completa.go
```

### Endpoints disponíveis

- `GET /health` - Health check (sem rate limiting)
- `GET /api/users` - Lista usuários (com rate limiting)
- `POST /api/users` - Cria usuário (com rate limiting)
- `GET /api/stats` - Estatísticas (com rate limiting por token)

### Testar

```bash
# Requests normais
curl http://localhost:8080/api/users

# Com token
curl -H "API_KEY: premium_token" http://localhost:8080/api/stats

# Testar rate limit
for i in {1..15}; do curl http://localhost:8080/api/users; echo ""; done
```
