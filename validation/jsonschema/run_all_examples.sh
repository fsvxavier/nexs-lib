#!/bin/bash

# Script para executar todos os exemplos do JSON Schema Validation
# Usage: ./run_all_examples.sh

set -e  # Exit on any error

echo "🚀 Executando todos os exemplos do JSON Schema Validation"
echo "========================================================"

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to run example
run_example() {
    local dir=$1
    local file=$2
    local name=$3
    
    echo ""
    echo -e "${BLUE}📁 Executando: ${name}${NC}"
    echo "----------------------------------------"
    
    if [ -d "$dir" ] && [ -f "$dir/$file" ]; then
        cd "$dir"
        
        # Check if go.mod exists, if not, create it temporarily
        if [ ! -f "go.mod" ]; then
            echo "go 1.23" > go.mod.tmp
            echo "" >> go.mod.tmp
            echo "require github.com/fsvxavier/nexs-lib v0.0.0" >> go.mod.tmp
            echo "" >> go.mod.tmp
            echo "replace github.com/fsvxavier/nexs-lib => ../../../../" >> go.mod.tmp
            mv go.mod.tmp go.mod
            temp_mod=true
        fi
        
        if go run "$file"; then
            echo -e "${GREEN}✅ $name executado com sucesso${NC}"
        else
            echo -e "${RED}❌ $name falhou${NC}"
            exit_code=1
        fi
        
        # Clean up temporary go.mod
        if [ "$temp_mod" = true ]; then
            rm -f go.mod
        fi
        
        cd - > /dev/null
    else
        echo -e "${YELLOW}⚠️  $name não encontrado em $dir/$file${NC}"
    fi
}

# Check if we're in the right directory
if [ ! -d "examples" ]; then
    echo -e "${RED}❌ Este script deve ser executado da raiz do módulo validation/jsonschema${NC}"
    echo "   Uso: cd validation/jsonschema && ./run_all_examples.sh"
    exit 1
fi

exit_code=0

# Navigate to examples directory
cd examples

# Run all examples
run_example "basic" "basic_validation.go" "Basic Validation Example"
run_example "hooks" "hooks_example.go" "Hooks System Example"
run_example "checks" "checks_example.go" "Custom Checks Example"
run_example "providers" "providers_example.go" "Providers Comparison Example"
run_example "migration" "migration_example.go" "Migration Example"
run_example "real_world" "real_world_examples.go" "Real World Examples"

cd - > /dev/null

echo ""
echo "========================================================"
if [ $exit_code -eq 0 ]; then
    echo -e "${GREEN}🎉 Todos os exemplos executados com sucesso!${NC}"
    echo ""
    echo "💡 Para executar exemplos individuais:"
    echo "   cd examples/basic && go run basic_validation.go"
    echo "   cd examples/hooks && go run hooks_example.go"
    echo "   cd examples/checks && go run checks_example.go"
    echo "   cd examples/providers && go run providers_example.go"
    echo "   cd examples/migration && go run migration_example.go"
    echo "   cd examples/real_world && go run real_world_examples.go"
    echo ""
    echo "📚 Consulte examples/README.md para mais informações"
else
    echo -e "${RED}❌ Alguns exemplos falharam${NC}"
    exit $exit_code
fi
