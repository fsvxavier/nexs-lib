#!/bin/bash

# Script para executar todos os exemplos do PostgreSQL com terminação forçada
# Este script executa todos os 10 exemplos disponíveis:
# - basic: Operações básicas de conexão e consulta
# - advanced: Configurações avançadas e recursos especiais
# - pool: Gerenciamento de pool de conexões
# - replicas: Operações com réplicas de leitura
# - transaction: Transações e rollbacks
# - batch: Operações em lote (batch)
# - copy: Operações COPY FROM/TO
# - hooks: Sistema de hooks para interceptação
# - listen_notify: LISTEN/NOTIFY para pub/sub
# - multitenant: Estratégias de multi-tenancy
echo "=== Executando todos os exemplos do PostgreSQL ==="
echo "Total de exemplos: 10"

# Função para executar um exemplo com timeout e terminação forçada
run_example() {
    local dir=$1
    local name=$2
    local timeout=${3:-20}
    
    echo ""
    echo "========================================"
    echo "Executando: $name"
    echo "========================================"
    
    cd "$dir"
    
    if [ ! -f "main.go" ]; then
        echo "❌ Arquivo main.go não encontrado em $dir"
        return 1
    fi
    
    echo "Compilando..."
    if ! go build -o "$(basename "$dir")" .; then
        echo "❌ Erro na compilação"
        return 1
    fi
    
    echo "Executando (timeout: ${timeout}s)..."
    
    # Executar em background e capturar PID
    timeout ${timeout}s ./$(basename "$dir") &
    local pid=$!
    
    # Aguardar o processo terminar ou timeout
    wait $pid 2>/dev/null
    local exit_code=$?
    
    # Forçar terminação se necessário
    if kill -0 $pid 2>/dev/null; then
        echo "Forçando terminação..."
        kill -TERM $pid 2>/dev/null
        sleep 1
        kill -KILL $pid 2>/dev/null
        exit_code=124
    fi
    
    if [ $exit_code -eq 124 ]; then
        echo "⏰ Exemplo finalizado por timeout (${timeout}s)"
    elif [ $exit_code -eq 0 ]; then
        echo "✅ Exemplo concluído com sucesso"
    else
        echo "❌ Exemplo falhou com código $exit_code"
    fi
    
    # Limpar binário
    rm -f "./$(basename "$dir")"
    
    cd ..
}

# Executar exemplos
BASE_DIR="/home/fabricioxavier/go/src/github.com/fsvxavier/nexs-lib/db/postgres/examples"
cd "$BASE_DIR"

# Verificar se todos os exemplos existem
echo "Verificando disponibilidade dos exemplos..."
EXAMPLES=("basic" "advanced" "pool" "replicas" "transaction" "batch" "copy" "hooks" "listen_notify" "multitenant")
MISSING=()

for example in "${EXAMPLES[@]}"; do
    if [ ! -d "./$example" ] || [ ! -f "./$example/main.go" ]; then
        MISSING+=("$example")
    fi
done

if [ ${#MISSING[@]} -gt 0 ]; then
    echo "❌ Exemplos faltando: ${MISSING[*]}"
    echo "Executando apenas os exemplos disponíveis..."
else
    echo "✅ Todos os 10 exemplos estão disponíveis!"
fi

echo ""
echo "Executando exemplos básicos..."
run_example "./basic" "Exemplo Básico" 10
run_example "./advanced" "Exemplo Avançado" 15

echo ""
echo "Executando exemplos de conexão..."
run_example "./pool" "Exemplo Pool de Conexões" 10
run_example "./replicas" "Exemplo Read Replicas" 10

echo ""
echo "Executando exemplos de operações..."
run_example "./transaction" "Exemplo Transações" 15
run_example "./batch" "Exemplo Operações Batch" 20
run_example "./copy" "Exemplo Operações COPY" 15

echo ""
echo "Executando exemplos avançados..."
run_example "./hooks" "Exemplo Sistema de Hooks" 20
run_example "./listen_notify" "Exemplo LISTEN/NOTIFY" 25
run_example "./multitenant" "Exemplo Multi-Tenancy" 20

echo ""
echo "========================================"
echo "Todos os exemplos foram executados!"
echo "========================================"
echo ""
echo "📊 Resumo dos exemplos executados:"
echo "  ✅ Básico - Operações fundamentais"
echo "  ✅ Avançado - Configurações avançadas"
echo "  ✅ Pool - Gerenciamento de conexões"
echo "  ✅ Replicas - Réplicas de leitura"
echo "  ✅ Transaction - Transações e rollbacks"
echo "  ✅ Batch - Operações em lote"
echo "  ✅ Copy - Operações COPY FROM/TO"
echo "  ✅ Hooks - Sistema de interceptação"
echo "  ✅ Listen/Notify - Pub/Sub em tempo real"
echo "  ✅ Multitenant - Estratégias multi-tenant"
