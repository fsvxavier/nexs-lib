#!/bin/bash

# Script para executar todos os exemplos da nexs-lib/httpserver
# Uso: ./run_examples.sh [exemplo]

set -e

EXAMPLES_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
EXAMPLES=("basic" "gin" "echo" "fasthttp" "atreugo" "advanced" "hooks-basic" "middlewares-basic" "complete")

print_header() {
    echo ""
    echo "======================================"
    echo "  $1"
    echo "======================================"
    echo ""
}

print_usage() {
    echo "🚀 Script para executar exemplos da Nexs Lib"
    echo ""
    echo "Uso:"
    echo "  $0                    # Lista todos os exemplos"
    echo "  $0 [exemplo]          # Executa um exemplo específico"
    echo "  $0 build              # Compila todos os exemplos"
    echo "  $0 test               # Testa todos os exemplos"
    echo "  $0 info               # Mostra informações detalhadas"
    echo ""
    echo "Exemplos disponíveis:"
    echo "  🎯 FRAMEWORKS: basic, gin, echo, fasthttp, atreugo, advanced"
    echo "  🔧 RECURSOS:   hooks-basic, middlewares-basic, complete"
    echo ""
}

build_example() {
    local example=$1
    print_header "Compilando $example"
    
    if [ ! -d "$EXAMPLES_DIR/$example" ]; then
        echo "❌ Exemplo '$example' não encontrado!"
        return 1
    fi
    
    cd "$EXAMPLES_DIR/$example"
    echo "📂 Diretório: $(pwd)"
    echo "🔨 Compilando..."
    
    if go build -o "/tmp/nexs-example-$example" .; then
        echo "✅ Compilação bem-sucedida!"
        echo "📦 Binário: /tmp/nexs-example-$example"
    else
        echo "❌ Falha na compilação!"
        return 1
    fi
}

run_example() {
    local example=$1
    print_header "Executando $example"
    
    if [ ! -d "$EXAMPLES_DIR/$example" ]; then
        echo "❌ Exemplo '$example' não encontrado!"
        return 1
    fi
    
    cd "$EXAMPLES_DIR/$example"
    echo "📂 Diretório: $(pwd)"
    echo "🚀 Iniciando servidor..."
    echo "   (Pressione Ctrl+C para parar)"
    echo ""
    
    go run main.go
}

build_all() {
    print_header "Compilando todos os exemplos"
    
    for example in "${EXAMPLES[@]}"; do
        if ! build_example "$example"; then
            echo "❌ Falha ao compilar $example"
            exit 1
        fi
        echo ""
    done
    
    echo "✅ Todos os exemplos compilados com sucesso!"
}

test_all() {
    print_header "Testando todos os exemplos"
    
    for example in "${EXAMPLES[@]}"; do
        echo "🧪 Testando $example..."
        
        if build_example "$example" > /dev/null 2>&1; then
            echo "✅ $example: OK"
        else
            echo "❌ $example: FALHOU"
            exit 1
        fi
    done
    
    echo ""
    echo "✅ Todos os testes passaram!"
}

list_examples() {
    print_header "Exemplos Disponíveis"
    
    echo "🎯 EXEMPLOS BÁSICOS:"
    for example in "basic" "gin" "echo" "fasthttp" "atreugo" "advanced"; do
        if [ -d "$EXAMPLES_DIR/$example" ]; then
            echo "📁 $example"
            if [ -f "$EXAMPLES_DIR/$example/README.md" ]; then
                # Extrair primeira linha de descrição do README
                desc=$(head -n 5 "$EXAMPLES_DIR/$example/README.md" | grep -E "^(Este exemplo|Esta pasta|.*demonstra)" | head -n 1 | sed 's/^[# ]*//' | sed 's/Este exemplo //')
                if [ -n "$desc" ]; then
                    echo "   $desc"
                fi
            fi
            echo ""
        fi
    done
    
    echo ""
    echo "🔧 EXEMPLOS DE RECURSOS:"
    for example in "hooks-basic" "middlewares-basic" "complete"; do
        if [ -d "$EXAMPLES_DIR/$example" ]; then
            echo "📁 $example"
            if [ -f "$EXAMPLES_DIR/$example/README.md" ]; then
                # Extrair primeira linha de descrição do README
                desc=$(head -n 5 "$EXAMPLES_DIR/$example/README.md" | grep -E "^(Este exemplo|Esta pasta|.*demonstra)" | head -n 1 | sed 's/^[# ]*//' | sed 's/Este exemplo //')
                if [ -n "$desc" ]; then
                    echo "   $desc"
                fi
            fi
            echo ""
        fi
    done
    
    echo "💡 TRILHA DE APRENDIZADO RECOMENDADA:"
    echo "   1. basic → gin → echo (frameworks)"
    echo "   2. hooks-basic → middlewares-basic → complete (recursos)"
    echo "   3. fasthttp → atreugo (performance)"
    echo "   4. advanced (produção)"
    echo ""
    echo "Para executar um exemplo:"
    echo "  $0 <nome_do_exemplo>"
    echo ""
    echo "Para ver detalhes:"
    echo "  cat <nome_do_exemplo>/README.md"
}

show_info() {
    print_header "Informações Detalhadas dos Exemplos"
    
    echo "📊 MATRIZ DE FUNCIONALIDADES:"
    echo ""
    printf "%-15s %-10s %-10s %-12s %-10s\n" "Exemplo" "Framework" "Hooks" "Middlewares" "Complexidade"
    echo "================================================================="
    printf "%-15s %-10s %-10s %-12s %-10s\n" "basic" "Fiber" "❌" "❌" "⭐"
    printf "%-15s %-10s %-10s %-12s %-10s\n" "gin" "Gin" "✅ Todos" "❌" "⭐⭐⭐"
    printf "%-15s %-10s %-10s %-12s %-10s\n" "echo" "Echo" "✅ Todos" "❌" "⭐⭐⭐"
    printf "%-15s %-10s %-10s %-12s %-10s\n" "fasthttp" "FastHTTP" "✅ Todos" "❌" "⭐⭐⭐⭐"
    printf "%-15s %-10s %-10s %-12s %-10s\n" "atreugo" "Atreugo" "✅ Todos" "❌" "⭐⭐⭐⭐"
    printf "%-15s %-10s %-10s %-12s %-10s\n" "advanced" "Fiber" "✅ Todos" "❌" "⭐⭐⭐⭐⭐"
    printf "%-15s %-10s %-10s %-12s %-10s\n" "hooks-basic" "Gin" "✅ Básicos" "❌" "⭐⭐"
    printf "%-15s %-10s %-10s %-12s %-10s\n" "middlewares" "Gin" "❌" "✅ Básicos" "⭐⭐⭐"
    printf "%-15s %-10s %-10s %-12s %-10s\n" "complete" "Gin" "✅ Todos" "✅ Todos" "⭐⭐⭐⭐⭐"
    echo ""
    
    echo "🚀 PERFORMANCE ESPERADA:"
    echo "  basic:           ~100k req/s"
    echo "  gin:             ~200k req/s"
    echo "  echo:            ~300k req/s"
    echo "  fasthttp:        ~500k req/s (máxima)"
    echo "  atreugo:         ~380k req/s"
    echo "  advanced:        ~180k req/s (produção)"
    echo "  hooks-basic:     ~150k req/s"
    echo "  middlewares:     ~120k req/s"
    echo "  complete:        ~100k req/s (completo)"
    echo ""
    
    echo "📚 TRILHA RECOMENDADA:"
    echo "  👶 Iniciante:   basic → gin → hooks-basic"
    echo "  🎯 Intermediário: echo → middlewares-basic → complete"
    echo "  ⚡ Performance:  fasthttp → atreugo"
    echo "  🏭 Produção:     advanced → complete"
    echo ""
}

# Função principal
main() {
    case "${1:-}" in
        ""|"list")
            list_examples
            ;;
        "build")
            build_all
            ;;
        "test")
            test_all
            ;;
        "info")
            show_info
            ;;
        "help"|"-h"|"--help")
            print_usage
            ;;
        *)
            if [[ " ${EXAMPLES[*]} " =~ " $1 " ]]; then
                run_example "$1"
            else
                echo "❌ Exemplo '$1' não encontrado!"
                echo ""
                print_usage
                exit 1
            fi
            ;;
    esac
}

# Executar função principal
main "$@"
