#!/bin/bash

# Script para executar todos os exemplos do domainerrors
set -e

echo "🚀 Executando todos os exemplos do DomainErrors"
echo "================================================="

# Função para executar um exemplo
run_example() {
    local example_dir=$1
    local example_name=$2
    
    echo ""
    echo "📁 Executando exemplo: $example_name"
    echo "-----------------------------------"
    
    if [ -d "$example_dir" ]; then
        cd "$example_dir"
        if [ -f "main.go" ]; then
            echo "▶️  Executando: go run main.go"
            go run main.go
            echo "✅ Exemplo $example_name executado com sucesso"
        else
            echo "❌ Arquivo main.go não encontrado em $example_dir"
        fi
        cd ..
    else
        echo "❌ Diretório $example_dir não encontrado"
    fi
}

# Verificar se estamos no diretório correto
if [ ! -f "README.md" ]; then
    echo "❌ Este script deve ser executado do diretório examples/"
    exit 1
fi

echo "📍 Diretório atual: $(pwd)"
echo "🔍 Exemplos disponíveis:"
for dir in */; do
    if [ -d "$dir" ] && [ -f "$dir/main.go" ]; then
        echo "  - $dir"
    fi
done

echo ""
echo "⚡ Iniciando execução..."

# Executar exemplos em ordem
run_example "basic" "Básico"
run_example "global" "Configuração Global"
run_example "advanced" "Avançado"

echo ""
echo "🎉 Todos os exemplos foram executados!"
echo "====================================="
echo ""
echo "📚 Para mais informações:"
echo "  - Documentação: ../README.md"
echo "  - Cada exemplo tem seu próprio README.md"
echo "  - Código fonte está comentado"
echo ""
echo "🔧 Para executar um exemplo específico:"
echo "  cd <exemplo>/ && go run main.go"
echo ""
echo "💡 Sugestões:"
echo "  1. Comece com o exemplo básico"
echo "  2. Explore configuração global"
echo "  3. Implemente padrões avançados"
echo ""
