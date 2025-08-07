#!/bin/bash

# Script para executar todos os exemplos do domainerrors

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
EXAMPLES_DIR="${SCRIPT_DIR}"

echo "🚀 Executando todos os exemplos do domainerrors..."
echo ""

# Função para executar um exemplo
run_example() {
    local example_name=$1
    local example_dir="${EXAMPLES_DIR}/${example_name}"
    
    if [ ! -d "$example_dir" ]; then
        echo "❌ Exemplo $example_name não encontrado em $example_dir"
        return 1
    fi
    
    echo "📁 Executando exemplo: $example_name"
    echo "   Diretório: $example_dir"
    
    cd "$example_dir"
    
    # Compilar o exemplo
    if ! go build -o "${example_name}-example" main.go; then
        echo "❌ Falha na compilação do exemplo $example_name"
        return 1
    fi
    
    echo "   ✅ Compilação bem-sucedida"
    
    # Executar o exemplo (com timeout de 30s para segurança)
    echo "   🔄 Executando..."
    if timeout 30s "./${example_name}-example" > "/tmp/${example_name}_output.log" 2>&1; then
        echo "   ✅ Execução bem-sucedida"
        
        # Mostrar as primeiras e últimas linhas da saída
        echo "   📄 Primeiras linhas da saída:"
        head -5 "/tmp/${example_name}_output.log" | sed 's/^/      /'
        echo "      ..."
        echo "   📄 Últimas linhas da saída:"
        tail -3 "/tmp/${example_name}_output.log" | sed 's/^/      /'
    else
        echo "   ❌ Falha na execução do exemplo $example_name"
        echo "   📄 Saída de erro:"
        tail -10 "/tmp/${example_name}_output.log" | sed 's/^/      /'
        return 1
    fi
    
    # Limpar arquivo executável
    rm -f "${example_name}-example"
    
    echo ""
}

# Lista de exemplos para executar
examples=("basic" "global" "advanced" "outros")

success_count=0
total_count=${#examples[@]}

# Executar cada exemplo
for example in "${examples[@]}"; do
    if run_example "$example"; then
        ((success_count++))
    else
        echo "❌ Exemplo $example falhou"
        echo ""
    fi
done

# Resumo final
echo "================================================="
echo "📊 Resumo da Execução:"
echo "   Total de exemplos: $total_count"
echo "   Bem-sucedidos: $success_count"
echo "   Falharam: $((total_count - success_count))"

if [ $success_count -eq $total_count ]; then
    echo "   🎉 Todos os exemplos executaram com sucesso!"
    exit 0
else
    echo "   ⚠️  Alguns exemplos falharam"
    exit 1
fi
