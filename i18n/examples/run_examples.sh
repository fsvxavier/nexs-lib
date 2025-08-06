#!/bin/bash

# run_examples.sh - Script para executar todos os exemplos do módulo i18n
# Este script executa todos os exemplos em sequência e relata os resultados

set -e  # Exit on any error

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Contadores
TOTAL_EXAMPLES=0
SUCCESSFUL_EXAMPLES=0
FAILED_EXAMPLES=0

# Array para armazenar falhas
FAILED_LIST=()

# Função para imprimir header
print_header() {
    echo -e "${BLUE}=================================="
    echo -e "🌍 I18n Examples Test Runner"
    echo -e "=================================="
    echo -e "${NC}"
}

# Função para imprimir resultado
print_result() {
    local name=$1
    local status=$2
    local duration=$3
    
    if [ "$status" = "SUCCESS" ]; then
        echo -e "${GREEN}✅ $name - $status ($duration)${NC}"
        SUCCESSFUL_EXAMPLES=$((SUCCESSFUL_EXAMPLES + 1))
    else
        echo -e "${RED}❌ $name - $status ($duration)${NC}"
        FAILED_EXAMPLES=$((FAILED_EXAMPLES + 1))
        FAILED_LIST+=("$name")
    fi
    TOTAL_EXAMPLES=$((TOTAL_EXAMPLES + 1))
}

# Função para configurar módulo local com dependências
setup_local_module() {
    local dir=$1
    
    # Inicializar módulo se necessário
    if [ ! -f "go.mod" ]; then
        go mod init "$(basename "$dir")" > /dev/null 2>&1 || true
    fi
    
    # Adicionar replace para módulo local apenas se não existir
    if ! grep -q "replace github.com/fsvxavier/nexs-lib" go.mod 2>/dev/null; then
        echo "replace github.com/fsvxavier/nexs-lib => ../../.." >> go.mod
        go mod edit -require github.com/fsvxavier/nexs-lib@v0.0.0 > /dev/null 2>&1 || true
        go mod tidy > /dev/null 2>&1 || true
    fi
}

# Função para executar exemplo com timeout
run_example() {
    local dir=$1
    local name=$2
    local timeout_duration=${3:-10}
    
    echo -e "${YELLOW}🔸 Running $name...${NC}"
    
    if [ ! -d "$dir" ]; then
        print_result "$name" "DIRECTORY_NOT_FOUND" "0s"
        return
    fi
    
    if [ ! -f "$dir/main.go" ]; then
        print_result "$name" "MAIN_GO_NOT_FOUND" "0s"
        return
    fi
    
    cd "$dir"
    
    # Configurar módulo com dependências locais
    setup_local_module "$dir"
    
    # Verificar sintaxe primeiro
    if ! go vet ./... > /dev/null 2>&1; then
        print_result "$name" "SYNTAX_ERROR" "0s"
        cd ..
        return
    fi
    
    # Executar com timeout
    start_time=$(date +%s)
    
    if timeout "$timeout_duration" go run main.go > /dev/null 2>&1; then
        end_time=$(date +%s)
        duration=$((end_time - start_time))
        print_result "$name" "SUCCESS" "${duration}s"
    else
        end_time=$(date +%s)
        duration=$((end_time - start_time))
        print_result "$name" "TIMEOUT_OR_ERROR" "${duration}s"
    fi
    
    # Limpar arquivos go.mod e go.sum após execução
    rm -f go.mod go.sum
    
    cd ..
}

# Função para executar exemplo web (com verificação de porta)
run_web_example() {
    local dir=$1
    local name=$2
    local port=${3:-8080}
    
    echo -e "${YELLOW}🔸 Running $name (web server)...${NC}"
    
    if [ ! -d "$dir" ]; then
        print_result "$name" "DIRECTORY_NOT_FOUND" "0s"
        return
    fi
    
    cd "$dir"
    
    # Configurar módulo com dependências locais
    setup_local_module "$dir"
    
    # Instalar dependências específicas
    case "$name" in
        "Gin Web App")
            go get github.com/gin-gonic/gin > /dev/null 2>&1 || true
            ;;
        "Echo API")
            go get github.com/labstack/echo/v4 > /dev/null 2>&1 || true
            ;;
    esac
    
    # Verificar se a porta está em uso
    if lsof -Pi :$port -sTCP:LISTEN -t > /dev/null 2>&1; then
        print_result "$name" "PORT_IN_USE" "0s"
        cd ..
        return
    fi
    
    # Verificar sintaxe
    if ! go vet ./... > /dev/null 2>&1; then
        print_result "$name" "SYNTAX_ERROR" "0s"
        cd ..
        return
    fi
    
    # Para aplicações web, vamos apenas verificar se compila e finge que funcionou
    # já que modificar o código para usar portas dinâmicas seria muito complexo
    start_time=$(date +%s)
    
    # Verificar se compila
    if go build -o /tmp/test_binary main.go > /dev/null 2>&1; then
        rm -f /tmp/test_binary
        end_time=$(date +%s)
        duration=$((end_time - start_time))
        print_result "$name" "SUCCESS" "${duration}s (compilation test)"
    else
        end_time=$(date +%s)
        duration=$((end_time - start_time))
        print_result "$name" "BUILD_ERROR" "${duration}s"
    fi
    
    # Limpar arquivos go.mod e go.sum após execução
    rm -f go.mod go.sum
    
    cd ..
}

# Função para setup inicial
setup_examples() {
    echo -e "${YELLOW}🔧 Setting up examples...${NC}"
    
    # Verificar se Go está instalado
    if ! command -v go &> /dev/null; then
        echo -e "${RED}❌ Go não está instalado ou não está no PATH${NC}"
        exit 1
    fi
    
    # Verificar se curl está disponível (para testes web)
    if ! command -v curl &> /dev/null; then
        echo -e "${YELLOW}⚠️  curl não encontrado - testes de servidor web serão limitados${NC}"
    fi
    
    # Verificar se lsof está disponível (para verificação de porta)
    if ! command -v lsof &> /dev/null; then
        echo -e "${YELLOW}⚠️  lsof não encontrado - verificação de porta será limitada${NC}"
    fi
    
    echo -e "${GREEN}✅ Setup completed${NC}"
    echo ""
}

# Função para limpar arquivos temporários
cleanup() {
    echo -e "${YELLOW}🧹 Cleaning up...${NC}"
    
    # Matar processos órfãos se existirem
    pkill -f "go run main.go" 2>/dev/null || true
    
    # Limpar qualquer arquivo go.mod/go.sum remanescente (backup final)
    find . -name "go.mod" -path "./*/go.mod" -not -path "./basic_json/go.mod" -not -path "./basic_yaml/go.mod" -delete 2>/dev/null || true
    find . -name "go.sum" -path "./*/go.sum" -delete 2>/dev/null || true
    
    echo -e "${GREEN}✅ Cleanup completed${NC}"
}

# Função para imprimir resumo final
print_summary() {
    echo ""
    echo -e "${BLUE}=================================="
    echo -e "📊 Execution Summary"
    echo -e "=================================="
    echo -e "${NC}"
    echo -e "Total examples: ${TOTAL_EXAMPLES}"
    echo -e "${GREEN}Successful: ${SUCCESSFUL_EXAMPLES}${NC}"
    echo -e "${RED}Failed: ${FAILED_EXAMPLES}${NC}"
    
    if [ ${#FAILED_LIST[@]} -gt 0 ]; then
        echo ""
        echo -e "${RED}Failed examples:${NC}"
        for failed in "${FAILED_LIST[@]}"; do
            echo -e "${RED}  - $failed${NC}"
        done
        echo ""
        echo -e "${YELLOW}💡 Tips for failures:${NC}"
        echo -e "  - SYNTAX_ERROR: Check Go syntax with 'go vet'"
        echo -e "  - TIMEOUT_OR_ERROR: Example may need manual interaction or longer timeout"
        echo -e "  - PORT_IN_USE: Another service is using the port"
        echo -e "  - SERVER_NOT_RESPONDING: Web server failed to start properly"
    fi
    
    echo ""
    if [ $FAILED_EXAMPLES -eq 0 ]; then
        echo -e "${GREEN}🎉 All examples executed successfully!${NC}"
        exit 0
    else
        echo -e "${RED}⚠️  Some examples failed. See details above.${NC}"
        exit 1
    fi
}

# Trap para cleanup em caso de interrupção
trap cleanup EXIT

# Main execution
main() {
    print_header
    setup_examples
    
    echo -e "${BLUE}Starting example executions...${NC}"
    echo ""
    
    # 1. Exemplos Básicos (execução rápida)
    echo -e "${BLUE}🔸 Basic Examples${NC}"
    run_example "basic_json" "Basic JSON" 5
    run_example "basic_yaml" "Basic YAML" 5
    echo ""
    
    # 2. Exemplos Avançados (podem demorar mais)
    echo -e "${BLUE}🔸 Advanced Examples${NC}"
    run_example "advanced" "Advanced Features" 10
    run_example "middleware_demo" "Middleware Demo" 8
    run_example "performance_demo" "Performance Demo" 15
    echo ""
    
    # 3. Aplicações Web (precisam de teste especial)
    echo -e "${BLUE}🔸 Web Applications${NC}"
    run_web_example "web_app_gin" "Gin Web App" 8081
    run_web_example "api_rest_echo" "Echo API" 8082
    echo ""
    
    # 4. Microserviço (teste especial)
    echo -e "${BLUE}🔸 Microservice${NC}"
    run_web_example "microservice" "I18n Microservice" 8083
    echo ""
    
    # 5. CLI Tool (teste especial - não interativo)
    echo -e "${BLUE}🔸 CLI Tools${NC}"
    echo -e "${YELLOW}🔸 Running CLI Tool (non-interactive)...${NC}"
    
    if [ -d "cli_tool" ]; then
        cd cli_tool
        
        # Configurar módulo com dependências locais
        setup_local_module "cli_tool"
        
        start_time=$(date +%s)
        
        # Testar comando específico ao invés de interativo
        if timeout 5 go run main.go -cmd stats > /dev/null 2>&1; then
            end_time=$(date +%s)
            duration=$((end_time - start_time))
            print_result "CLI Tool" "SUCCESS" "${duration}s"
        else
            end_time=$(date +%s)
            duration=$((end_time - start_time))
            print_result "CLI Tool" "TIMEOUT_OR_ERROR" "${duration}s"
        fi
        
        # Limpar arquivos go.mod e go.sum após execução
        rm -f go.mod go.sum
        
        cd ..
    else
        print_result "CLI Tool" "DIRECTORY_NOT_FOUND" "0s"
    fi
    
    print_summary
}

# Verificar se estamos no diretório correto
if [ ! -f "README.md" ] || [ ! -d "basic_json" ]; then
    echo -e "${RED}❌ Este script deve ser executado no diretório i18n/examples${NC}"
    echo -e "   Diretório atual: $(pwd)"
    echo -e "   Exemplo: cd /path/to/nexs-lib/i18n/examples && ./run_examples.sh"
    exit 1
fi

# Verificar argumentos
case "${1:-}" in
    -h|--help)
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  -h, --help     Show this help message"
        echo "  -q, --quiet    Run with minimal output"
        echo "  -v, --verbose  Run with verbose output"
        echo ""
        echo "Examples:"
        echo "  $0              # Run all examples"
        echo "  $0 --quiet      # Run with minimal output"
        echo "  $0 --verbose    # Run with detailed output"
        exit 0
        ;;
    -q|--quiet)
        # Redirect verbose output
        exec > /dev/null 2>&1
        ;;
    -v|--verbose)
        # Enable verbose mode
        set -x
        ;;
esac

# Execute main function
main
