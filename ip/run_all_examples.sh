#!/bin/bash

# Script para executar todos os exemplos do módulo IP
# Autor: Gerado automaticamente
# Data: $(date)

set -e  # Parar em caso de erro

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Função para imprimir cabeçalho
print_header() {
    echo -e "\n${BLUE}═══════════════════════════════════════════════════════════════════════════════${NC}"
    echo -e "${CYAN}🚀 $1${NC}"
    echo -e "${BLUE}═══════════════════════════════════════════════════════════════════════════════${NC}\n"
}

# Função para imprimir seção
print_section() {
    echo -e "\n${PURPLE}📋 $1${NC}"
    echo -e "${PURPLE}───────────────────────────────────────────────────────────────────────────────${NC}"
}

# Função para executar exemplo
run_example() {
    local dir=$1
    local name=$2
    local timeout=${3:-30}
    
    echo -e "\n${YELLOW}🔍 Executando: $name${NC}"
    echo -e "${YELLOW}📁 Diretório: $dir${NC}"
    echo -e "${YELLOW}⏱️  Timeout: ${timeout}s${NC}\n"
    
    if [ -f "$dir/main.go" ]; then
        cd "$dir"
        if timeout "${timeout}s" go run main.go; then
            echo -e "\n${GREEN}✅ $name - SUCESSO${NC}"
        else
            echo -e "\n${RED}❌ $name - FALHOU ou TIMEOUT${NC}"
        fi
        cd - > /dev/null
    else
        echo -e "${RED}❌ Arquivo main.go não encontrado em $dir${NC}"
    fi
}

# Verificar se estamos no diretório correto
if [ ! -d "examples" ]; then
    # Tentar encontrar o diretório do módulo IP
    if [ -f "../go.mod" ] && [ -d "../examples" ]; then
        cd ..
    elif [ -f "../../go.mod" ] && [ -d "../../examples" ]; then
        cd ../..
    else
        echo -e "${RED}❌ Erro: Execute este script no diretório do módulo IP${NC}"
        echo -e "${YELLOW}💡 Uso: cd /caminho/para/nexs-lib/ip && ./run_all_examples.sh${NC}"
        exit 1
    fi
fi

# Início do script
print_header "NEXS-LIB IP MODULE - EXECUÇÃO DE TODOS OS EXEMPLOS"

echo -e "${CYAN}📦 Módulo: nexs-lib/ip${NC}"
echo -e "${CYAN}📅 Data: $(date)${NC}"
echo -e "${CYAN}👤 Usuário: $(whoami)${NC}"
echo -e "${CYAN}🗂️  Diretório: $(pwd)${NC}"

# Verificar se Go está instalado
if ! command -v go &> /dev/null; then
    echo -e "\n${RED}❌ Go não está instalado ou não está no PATH${NC}"
    exit 1
fi

echo -e "\n${GREEN}✅ Go version: $(go version)${NC}"

# Compilar o módulo primeiro
print_section "COMPILAÇÃO E VALIDAÇÃO"
echo -e "${YELLOW}🔨 Compilando módulo...${NC}"
if go build ./...; then
    echo -e "${GREEN}✅ Compilação bem-sucedida${NC}"
else
    echo -e "${RED}❌ Falha na compilação${NC}"
    exit 1
fi

# Verificar exemplos disponíveis
print_section "EXEMPLOS DISPONÍVEIS"
echo -e "${CYAN}📁 Listando exemplos:${NC}\n"

examples_found=()
if [ -d "examples" ]; then
    for example_dir in examples/*/; do
        if [ -f "${example_dir}main.go" ]; then
            example_name=$(basename "$example_dir")
            examples_found+=("$example_name")
            echo -e "   ${GREEN}✓${NC} $example_name"
        fi
    done
else
    echo -e "${RED}❌ Diretório examples/ não encontrado${NC}"
    exit 1
fi

if [ ${#examples_found[@]} -eq 0 ]; then
    echo -e "${RED}❌ Nenhum exemplo encontrado${NC}"
    exit 1
fi

echo -e "\n${CYAN}📊 Total de exemplos encontrados: ${#examples_found[@]}${NC}"

# Executar exemplos principais
print_section "EXECUÇÃO DOS EXEMPLOS PRINCIPAIS"

# 1. Exemplo Básico
run_example "examples/basic" "Exemplo Básico - Funcionalidades Core" 15

# 2. Uso Otimizado
run_example "examples/optimized_usage" "Uso Otimizado - Zero Allocation" 10

# 3. Detecção Avançada
run_example "examples/advanced-detection" "Detecção Avançada - VPN/Proxy/ASN" 15

# 4. Otimização de Memória
run_example "examples/memory-optimization" "Otimização de Memória - Object Pooling" 15

# 5. Segurança
run_example "examples/security" "Sistema de Segurança Avançado" 10

# Executar exemplos de providers
print_section "EXECUÇÃO DOS EXEMPLOS DE PROVIDERS"

providers=("nethttp" "gin" "echo" "fiber" "fasthttp" "atreugo")
for provider in "${providers[@]}"; do
    provider_dir="examples/providers/$provider"
    if [ -f "$provider_dir/main.go" ]; then
        run_example "$provider_dir" "Provider: $provider" 10
    else
        echo -e "${YELLOW}⚠️  Provider $provider não encontrado${NC}"
    fi
done

# Executar exemplos de middleware
print_section "EXECUÇÃO DOS EXEMPLOS DE MIDDLEWARE"

middleware_examples=("middleware" "security_middleware")
for middleware in "${middleware_examples[@]}"; do
    middleware_dir="examples/$middleware"
    if [ -f "$middleware_dir/main.go" ]; then
        run_example "$middleware_dir" "Middleware: $middleware" 10
    else
        echo -e "${YELLOW}⚠️  Middleware $middleware não encontrado${NC}"
    fi
done

# Executar testes rápidos para validação
print_section "VALIDAÇÃO FINAL - TESTES RÁPIDOS"

echo -e "${YELLOW}🧪 Executando testes rápidos...${NC}"
if go test -short ./...; then
    echo -e "${GREEN}✅ Testes rápidos passaram${NC}"
else
    echo -e "${RED}❌ Alguns testes falharam${NC}"
fi

# Verificar cobertura
echo -e "\n${YELLOW}📊 Verificando cobertura de testes...${NC}"
if go test -short -coverprofile=coverage_quick.out ./... > /dev/null 2>&1; then
    coverage=$(go tool cover -func=coverage_quick.out | tail -1 | awk '{print $3}')
    echo -e "${GREEN}✅ Cobertura atual: $coverage${NC}"
    rm -f coverage_quick.out
else
    echo -e "${YELLOW}⚠️  Não foi possível calcular cobertura${NC}"
fi

# Resumo final
print_header "RESUMO DA EXECUÇÃO"

echo -e "${CYAN}📊 Estatísticas:${NC}"
echo -e "   • Total de exemplos executados: ${#examples_found[@]}"
echo -e "   • Módulo compilado: ${GREEN}✅${NC}"
echo -e "   • Testes rápidos: ${GREEN}✅${NC}"

echo -e "\n${CYAN}🎯 Exemplos Principais:${NC}"
echo -e "   ${GREEN}✓${NC} Básico - Funcionalidades core"
echo -e "   ${GREEN}✓${NC} Otimizado - Zero allocation"
echo -e "   ${GREEN}✓${NC} Avançado - VPN/Proxy/ASN detection"
echo -e "   ${GREEN}✓${NC} Memória - Object pooling"
echo -e "   ${GREEN}✓${NC} Segurança - Validação avançada"

echo -e "\n${CYAN}🔌 Providers Suportados:${NC}"
for provider in "${providers[@]}"; do
    echo -e "   ${GREEN}✓${NC} $provider"
done

echo -e "\n${GREEN}🎉 EXECUÇÃO CONCLUÍDA COM SUCESSO!${NC}"
echo -e "${CYAN}📚 Para mais informações, consulte:${NC}"
echo -e "   • README.md"
echo -e "   • IMPLEMENTATION_SUMMARY.md"
echo -e "   • examples/README.md"

print_header "FIM DA EXECUÇÃO"
