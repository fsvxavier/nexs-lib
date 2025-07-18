#!/bin/bash

# Script para executar todos os exemplos do PostgreSQL
echo "=== Executando todos os exemplos do PostgreSQL ==="

# Função para executar um exemplo com timeout
run_example() {
    local dir=$1
    local name=$2
    local timeout=${3:-30}
    
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
    if ! go build -v .; then
        echo "❌ Erro na compilação"
        return 1
    fi
    
    echo "Executando (timeout: ${timeout}s)..."
    timeout ${timeout}s ./$(basename "$dir") 2>&1 &
    local pid=$!
    
    # Wait for process to complete or timeout
    wait $pid 2>/dev/null
    local exit_code=$?
    
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

run_example "./basic" "Exemplo Básico" 40
run_example "./pool" "Exemplo Pool de Conexões" 40
run_example "./advanced" "Exemplo Avançado" 70
run_example "./replicas" "Exemplo Read Replicas" 40

echo ""
echo "========================================"
echo "Todos os exemplos foram executados!"
echo "========================================"
