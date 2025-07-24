#!/bin/bash

# Nexs Infrastructure Validation Script
# Validates configuration files syntax and structure

set -e

echo "🔍 Validating Nexs Observability Infrastructure..."
echo ""

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
}

validate_file() {
    local file="$1"
    local description="$2"
    
    if [[ -f "$file" ]]; then
        success "$description: $file"
        return 0
    else
        error "$description: $file (missing)"
        return 1
    fi
}

validate_yaml() {
    local file="$1"
    local description="$2"
    
    if [[ -f "$file" ]]; then
        # Try to parse YAML with Python
        if command -v python3 >/dev/null 2>&1; then
            if python3 -c "import yaml; yaml.safe_load(open('$file'))" >/dev/null 2>&1; then
                success "$description: $file (valid YAML)"
                return 0
            else
                error "$description: $file (invalid YAML)"
                return 1
            fi
        else
            warning "$description: $file (YAML validation skipped - python3 not available)"
            return 0
        fi
    else
        error "$description: $file (missing)"
        return 1
    fi
}

# Validation counters
passed=0
failed=0

echo "📁 Validating file structure..."
echo ""

# Core files
files=(
    "docker-compose.yml:Docker Compose configuration"
    "manage.sh:Management script"
    "Makefile:Build automation"
    "README.md:Documentation"
    "NEXT_STEPS.md:Roadmap"
)

for item in "${files[@]}"; do
    IFS=':' read -r file desc <<< "$item"
    if validate_file "$file" "$desc"; then
        ((passed++))
    else
        ((failed++))
    fi
done

echo ""
echo "⚙️  Validating configuration files..."
echo ""

# YAML configurations
yaml_files=(
    "configs/otel-collector-config.yaml:OpenTelemetry Collector"
    "configs/tempo.yaml:Tempo configuration"
    "configs/prometheus.yml:Prometheus configuration"
)

for item in "${yaml_files[@]}"; do
    IFS=':' read -r file desc <<< "$item"
    if validate_yaml "$file" "$desc"; then
        ((passed++))
    else
        ((failed++))
    fi
done

# Config files (non-YAML)
config_files=(
    "configs/logstash.conf:Logstash pipeline"
    "configs/fluentd.conf:Fluentd configuration"
)

for item in "${config_files[@]}"; do
    IFS=':' read -r file desc <<< "$item"
    if validate_file "$file" "$desc"; then
        ((passed++))
    else
        ((failed++))
    fi
done

echo ""
echo "📊 Validating Grafana configurations..."
echo ""

# Grafana files
grafana_files=(
    "grafana/provisioning/datasources/datasources.yaml:Grafana datasources"
    "grafana/provisioning/dashboards/dashboards.yaml:Grafana dashboard config"
    "grafana/dashboards/nexs-tracer-overview.json:Tracer dashboard"
    "grafana/dashboards/nexs-logger-overview.json:Logger dashboard"
    "grafana/dashboards/infrastructure-health.json:Infrastructure dashboard"
)

for item in "${grafana_files[@]}"; do
    IFS=':' read -r file desc <<< "$item"
    if [[ "$file" == *.yaml ]]; then
        if validate_yaml "$file" "$desc"; then
            ((passed++))
        else
            ((failed++))
        fi
    else
        if validate_file "$file" "$desc"; then
            ((passed++))
        else
            ((failed++))
        fi
    fi
done

echo ""
echo "🗄️  Validating initialization scripts..."
echo ""

# Init scripts
init_files=(
    "init/postgres/init.sql:PostgreSQL initialization"
    "init/mongodb/init.js:MongoDB initialization"
)

for item in "${init_files[@]}"; do
    IFS=':' read -r file desc <<< "$item"
    if validate_file "$file" "$desc"; then
        ((passed++))
    else
        ((failed++))
    fi
done

echo ""
echo "🔧 Validating script permissions..."
echo ""

# Check executable permissions
if [[ -x "manage.sh" ]]; then
    success "manage.sh has executable permissions"
    ((passed++))
else
    warning "manage.sh is not executable (run: chmod +x manage.sh)"
    ((failed++))
fi

echo ""
echo "📋 Validation Summary"
echo "===================="
echo -e "Passed: ${GREEN}$passed${NC}"
echo -e "Failed: ${RED}$failed${NC}"
echo -e "Total:  $((passed + failed))"
echo ""

if [[ $failed -eq 0 ]]; then
    success "All validations passed! 🎉"
    echo ""
    info "Next steps:"
    echo "  1. Install Docker and Docker Compose"
    echo "  2. Run: make infra-up"
    echo "  3. Check status: make infra-status"
    echo "  4. View services: make infra-urls"
    echo ""
    exit 0
else
    error "Some validations failed. Please fix the issues above."
    echo ""
    exit 1
fi
