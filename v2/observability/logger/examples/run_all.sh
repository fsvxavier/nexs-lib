#!/bin/bash

# Script para executar todos os exemplos do Logger v2
# Seguindo as práticas do prompt de engenheiro sênior

set -e

echo "=== Logger v2 - Executando Todos os Exemplos ==="
echo "Data: $(date)"
echo "Diretório: $(pwd)"
echo ""

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Função para log com timestamp
log() {
    echo -e "${BLUE}[$(date +'%H:%M:%S')]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Contador de sucessos e falhas
success_count=0
failure_count=0
total_time=0

# Lista de exemplos para executar
examples=(
    "basic"
    "structured" 
    "context-aware"
    "async"
    "middleware"
    "providers"
    "microservices"
    "web-app"
)

log "Verificando pré-requisitos..."

# Verifica se go está instalado
if ! command -v go &> /dev/null; then
    error "Go não está instalado. Instale Go 1.21+ para continuar."
    exit 1
fi

# Verifica versão do Go
go_version=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | grep -oE '[0-9]+\.[0-9]+')
required_version="1.21"

if [ "$(printf '%s\n' "$required_version" "$go_version" | sort -V | head -n1)" != "$required_version" ]; then
    error "Go versão $required_version+ é necessária. Versão atual: $go_version"
    exit 1
fi

success "Go versão $go_version detectado ✓"

# Função para executar um exemplo
run_example() {
    local example=$1
    local start_time=$(date +%s)
    
    log "Executando exemplo: $example"
    
    if [ ! -d "$example" ]; then
        error "Diretório '$example' não encontrado"
        ((failure_count++))
        return 1
    fi
    
    if [ ! -f "$example/main.go" ]; then
        error "Arquivo '$example/main.go' não encontrado"
        ((failure_count++))
        return 1
    fi
    
    cd "$example"
    
    # Verifica se go.mod existe
    if [ ! -f "go.mod" ]; then
        warning "go.mod não encontrado em $example, criando..."
        go mod init "logger-$example-example"
        echo "replace github.com/fsvxavier/nexs-lib/v2 => ../../../../" >> go.mod
        echo "require github.com/fsvxavier/nexs-lib/v2 v2.0.0-00010101000000-000000000000" >> go.mod
    fi
    
    # Download de dependências
    log "Baixando dependências para $example..."
    if ! go mod tidy &> /dev/null; then
        warning "Falha no go mod tidy para $example"
    fi
    
    # Compilação
    log "Compilando $example..."
    if ! go build -o "../temp_$example" . &> /dev/null; then
        error "Falha na compilação de $example"
        cd ..
        ((failure_count++))
        return 1
    fi
    
    # Execução
    log "Executando $example..."
    if timeout 30s go run main.go > "../output_$example.log" 2>&1; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        total_time=$((total_time + duration))
        
        success "Exemplo $example executado com sucesso (${duration}s)"
        ((success_count++))
    else
        local exit_code=$?
        if [ $exit_code -eq 124 ]; then
            warning "Exemplo $example interrompido por timeout (30s)"
            # Para web-app que pode rodar indefinidamente, consideramos sucesso
            if [ "$example" = "web-app" ]; then
                success "Exemplo $example executado com sucesso (timeout esperado)"
                ((success_count++))
            else
                ((failure_count++))
            fi
        else
            error "Exemplo $example falhou com código de saída $exit_code"
            ((failure_count++))
        fi
    fi
    
    cd ..
    
    # Cleanup
    rm -f "temp_$example"
}

# Executa todos os exemplos
log "Iniciando execução de ${#examples[@]} exemplos..."
echo ""

for example in "${examples[@]}"; do
    echo "----------------------------------------"
    run_example "$example"
    echo ""
done

# Relatório final
echo "========================================"
echo "RELATÓRIO FINAL"
echo "========================================"
echo "Exemplos executados: ${#examples[@]}"
echo "Sucessos: $success_count"
echo "Falhas: $failure_count"
echo "Tempo total: ${total_time}s"
echo ""

if [ $failure_count -eq 0 ]; then
    success "Todos os exemplos executados com sucesso! 🎉"
    echo ""
    echo "Logs de saída salvos em:"
    for example in "${examples[@]}"; do
        if [ -f "output_$example.log" ]; then
            echo "  - output_$example.log"
        fi
    done
else
    error "$failure_count exemplo(s) falharam."
    echo ""
    echo "Verifique os logs para detalhes:"
    for example in "${examples[@]}"; do
        if [ -f "output_$example.log" ]; then
            echo "  - output_$example.log"
        fi
    done
fi

echo ""
log "Execução concluída."

# Sai com código de erro se houve falhas
exit $failure_count
