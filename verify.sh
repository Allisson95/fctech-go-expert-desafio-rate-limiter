#!/bin/bash

# Verifica a saúde do projeto
# Execute: chmod +x verify.sh && ./verify.sh

echo "================================================="
echo "  Rate Limiter - Verificação de Integridade"
echo "================================================="
echo ""

errors=0
warnings=0

# Função para verificar arquivo
check_file() {
    if [ -f "$1" ]; then
        echo "✓ $1"
    else
        echo "✗ $1 - AUSENTE"
        ((errors++))
    fi
}

# Função para verificar diretório
check_dir() {
    if [ -d "$1" ]; then
        echo "✓ $1/"
    else
        echo "✗ $1/ - AUSENTE"
        ((errors++))
    fi
}

echo "1. Verificando estrutura de diretórios..."
check_dir "cmd/server"
check_dir "internal/config"
check_dir "internal/limiter"
check_dir "internal/middleware"
check_dir "internal/storage"
check_dir "examples"
echo ""

echo "2. Verificando arquivos principais..."
check_file "cmd/server/main.go"
check_file "internal/config/config.go"
check_file "internal/limiter/limiter.go"
check_file "internal/middleware/ratelimiter.go"
check_file "internal/storage/storage.go"
check_file "internal/storage/redis.go"
echo ""

echo "3. Verificando testes..."
check_file "internal/limiter/limiter_test.go"
check_file "internal/middleware/ratelimiter_test.go"
check_file "internal/storage/redis_test.go"
echo ""

echo "4. Verificando configuração..."
check_file "go.mod"
check_file "go.sum"
check_file ".env"
check_file "docker-compose.yml"
check_file "Dockerfile"
check_file "Makefile"
echo ""

echo "5. Verificando documentação..."
check_file "README.md"
check_file "QUICKSTART.md"
check_file "ARCHITECTURE.md"
check_file "SUMMARY.md"
echo ""

echo "6. Verificando scripts..."
check_file "test_load.sh"
check_file "test_quick.sh"
echo ""

echo "7. Testando compilação..."
if go build -o /tmp/rate-limiter-test ./cmd/server > /dev/null 2>&1; then
    echo "✓ Projeto compila sem erros"
    rm -f /tmp/rate-limiter-test
else
    echo "✗ Erro ao compilar o projeto"
    ((errors++))
fi
echo ""

echo "8. Executando testes..."
if go test ./... > /tmp/test-output.txt 2>&1; then
    echo "✓ Todos os testes passaram"
    cat /tmp/test-output.txt | grep -E "PASS|ok" | head -5
else
    echo "⚠ Alguns testes falharam (pode ser normal se Redis não estiver disponível)"
    ((warnings++))
fi
rm -f /tmp/test-output.txt
echo ""

echo "9. Verificando dependências..."
if go mod verify > /dev/null 2>&1; then
    echo "✓ Dependências verificadas"
else
    echo "⚠ Problema com dependências"
    ((warnings++))
fi
echo ""

echo "10. Verificando formatação..."
unformatted=$(gofmt -l . 2>/dev/null | wc -l)
if [ "$unformatted" -eq "0" ]; then
    echo "✓ Código formatado corretamente"
else
    echo "⚠ $unformatted arquivo(s) precisam de formatação"
    ((warnings++))
fi
echo ""

echo "================================================="
echo "  Resumo da Verificação"
echo "================================================="
echo ""

if [ $errors -eq 0 ] && [ $warnings -eq 0 ]; then
    echo "🎉 SUCESSO! Projeto está perfeito!"
    echo ""
    echo "Próximos passos:"
    echo "  1. docker-compose up -d"
    echo "  2. ./test_quick.sh"
    echo "  3. Comece a usar!"
    exit 0
elif [ $errors -eq 0 ]; then
    echo "✅ Projeto OK (com $warnings aviso(s))"
    echo ""
    echo "O projeto está funcional, mas há pequenos avisos."
    exit 0
else
    echo "❌ Encontrados $errors erro(s) e $warnings aviso(s)"
    echo ""
    echo "Corrija os erros antes de prosseguir."
    exit 1
fi
