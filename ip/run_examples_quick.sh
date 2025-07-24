#!/bin/bash

# Script rápido para executar exemplos principais (sem servidores HTTP)
# Autor: Gerado automaticamente

set -e

# Cores
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🚀 NEXS-LIB IP - Execução Rápida dos Exemplos Principais${NC}\n"

# Verificar diretório
if [ ! -d "examples" ]; then
    if [ -f "../go.mod" ] && [ -d "../examples" ]; then
        cd ..
    else
        echo "❌ Execute no diretório do módulo IP"
        exit 1
    fi
fi

examples=(
    "optimized_usage:Uso Otimizado"
    "advanced-detection:Detecção Avançada"
    "memory-optimization:Otimização de Memória"
    "security:Sistema de Segurança"
)

for example in "${examples[@]}"; do
    IFS=':' read -r dir name <<< "$example"
    echo -e "${YELLOW}🔍 $name${NC}"
    cd "examples/$dir"
    timeout 30s go run main.go || echo "⏱️ Timeout ou interrompido"
    cd - > /dev/null
    echo
done

echo -e "${GREEN}✅ Exemplos principais executados!${NC}"
