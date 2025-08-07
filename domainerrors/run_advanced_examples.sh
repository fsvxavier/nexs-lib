#!/bin/bash

# Script para executar exemplos das funcionalidades avançadas do domainerrors

echo "🚀 Executando Exemplos das Funcionalidades Avançadas - Domain Errors"
echo "=================================================================="

# Função para executar exemplo com tratamento de erro
run_example() {
    local example_name="$1"
    local example_path="$2"
    
    echo ""
    echo "📁 Executando: $example_name"
    echo "-----------------------------------"
    
    if [ -d "$example_path" ]; then
        cd "$example_path" || {
            echo "❌ Erro: Não foi possível acessar $example_path"
            return 1
        }
        
        if go run . ; then
            echo "✅ $example_name executado com sucesso!"
        else
            echo "❌ Erro ao executar $example_name"
            return 1
        fi
        
        cd - > /dev/null || exit 1
    else
        echo "❌ Erro: Diretório $example_path não encontrado"
        return 1
    fi
}

# Navegar para diretório base do domainerrors
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR" || {
    echo "❌ Erro: Não foi possível acessar o diretório do script"
    exit 1
}

echo "📍 Diretório atual: $(pwd)"

# Executar testes primeiro
echo ""
echo "🧪 Executando Testes"
echo "-------------------"
if go test ./advanced -v; then
    echo "✅ Todos os testes passaram!"
else
    echo "❌ Alguns testes falharam"
    echo "Continuando com exemplos..."
fi

# Executar benchmarks
echo ""
echo "📊 Executando Benchmarks de Performance"
echo "---------------------------------------"
if go test ./advanced -bench=BenchmarkAdvancedFeatures -benchmem -count=1 -timeout=30s; then
    echo "✅ Benchmarks executados com sucesso!"
else
    echo "⚠️ Alguns benchmarks podem ter falhado"
fi

# Executar exemplos
echo ""
echo "📋 Executando Exemplos"
echo "======================="

# Lista de exemplos para executar
examples=(
    "Básico:examples/basic"
    "Global Hooks:examples/global"
    "Avançado:examples/advanced"
    "Outros Casos:examples/outros"
    "Funcionalidades Avançadas:examples/advanced_features"
)

# Contador de sucessos e falhas
successful=0
failed=0

# Executar cada exemplo
for example in "${examples[@]}"; do
    IFS=':' read -r name path <<< "$example"
    
    if run_example "$name" "$path"; then
        ((successful++))
    else
        ((failed++))
    fi
done

# Resumo final
echo ""
echo "📈 RESUMO DA EXECUÇÃO"
echo "====================="
echo "✅ Exemplos executados com sucesso: $successful"
echo "❌ Exemplos com falhas: $failed"
echo "📊 Total de exemplos: $((successful + failed))"

if [ $failed -eq 0 ]; then
    echo ""
    echo "🎉 Todos os exemplos foram executados com sucesso!"
    echo ""
    echo "🎯 Funcionalidades Demonstradas:"
    echo "  • Error Aggregation - Agregação inteligente de erros"
    echo "  • Conditional Hooks - Hooks baseados em condições"
    echo "  • Retry Mechanism - Sistema de retry com backoff"
    echo "  • Error Recovery - Recuperação automática de erros"
    echo "  • Performance Pools - Otimizações de memória"
    echo "  • Lazy Stack Traces - Captura otimizada de stack"
    echo "  • String Interning - Otimização de strings comuns"
    echo ""
    echo "📚 Para mais informações, consulte:"
    echo "  • README.md - Documentação principal"
    echo "  • NEXT_STEPS.md - Roadmap de funcionalidades"
    echo "  • examples/*/README.md - Documentação específica"
    echo ""
    exit 0
else
    echo ""
    echo "⚠️ Alguns exemplos falharam. Verifique os logs acima."
    exit 1
fi
