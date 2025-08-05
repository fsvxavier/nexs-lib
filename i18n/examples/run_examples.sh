#!/bin/bash

set -e  # Exit on error

# Cores para output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

run_example() {
    local example_name=$1
    echo -e "\n${BLUE}=== Executando exemplo: ${example_name} ===${NC}"
    cd "$example_name"
    
    echo -e "${GREEN}Compilando...${NC}"
    go build -v
    
    echo -e "${GREEN}Executando...${NC}"
    ./"$(basename "$(pwd)")"
    
    cd ..
    echo -e "${GREEN}✓ Exemplo $example_name concluído${NC}"
}

# Diretório base dos exemplos
cd "$(dirname "$0")"

# Cabeçalho
echo -e "${BLUE}=============================="
echo -e "   Exemplos do módulo i18n"
echo -e "==============================${NC}"

# Executando exemplos básicos
echo -e "\n${BLUE}Executando exemplos autocontidos:${NC}"
for example in "basic" "factory" "cached" "advanced" "formats" "hooks" "http"; do
    run_example "$example"
done

# Instruções para o exemplo HTTP
echo -e "\n${BLUE}=== Exemplo HTTP ===${NC}"
echo -e "${GREEN}Para testar o exemplo HTTP, execute:${NC}"
echo "cd http && go run main.go"
echo "Depois acesse: http://localhost:8080"
echo "Endpoints disponíveis:"
echo "  - / -> Tradução simples"
echo "  - /with-vars -> Tradução com variáveis"
echo "  - /plural -> Tradução com pluralização"
echo -e "\n${BLUE}========== Fim ==========${NC}"
