#!/bin/bash

# Nexs infraestructure Simple Validation
echo "🔍 Validating Nexs Observability infraestructure..."
echo ""

# Check main files
echo "📁 Checking main files..."
files=(
    "docker-compose.yml"
    "manage.sh"
    "Makefile"
    "README.md"
    "NEXT_STEPS.md"
)

for file in "${files[@]}"; do
    if [[ -f "$file" ]]; then
        echo "✅ $file"
    else
        echo "❌ $file (missing)"
    fi
done

echo ""
echo "⚙️ Checking configuration files..."
configs=(
    "configs/otel-collector-config.yaml"
    "configs/tempo.yaml"
    "configs/prometheus.yml"
    "configs/logstash.conf"
    "configs/fluentd.conf"
)

for config in "${configs[@]}"; do
    if [[ -f "$config" ]]; then
        echo "✅ $config"
    else
        echo "❌ $config (missing)"
    fi
done

echo ""
echo "📊 Checking Grafana files..."
grafana_files=(
    "grafana/provisioning/datasources/datasources.yaml"
    "grafana/provisioning/dashboards/dashboards.yaml"
    "grafana/dashboards/nexs-tracer-overview.json"
    "grafana/dashboards/nexs-logger-overview.json"
    "grafana/dashboards/infraestructure-health.json"
)

for gfile in "${grafana_files[@]}"; do
    if [[ -f "$gfile" ]]; then
        echo "✅ $gfile"
    else
        echo "❌ $gfile (missing)"
    fi
done

echo ""
echo "🗄️ Checking initialization scripts..."
init_files=(
    "init/postgres/init.sql"
    "init/mongodb/init.js"
)

for init in "${init_files[@]}"; do
    if [[ -f "$init" ]]; then
        echo "✅ $init"
    else
        echo "❌ $init (missing)"
    fi
done

echo ""
echo "🔧 Checking permissions..."
if [[ -x "manage.sh" ]]; then
    echo "✅ manage.sh is executable"
else
    echo "❌ manage.sh needs +x permission"
fi

echo ""
echo "📋 Structure Overview:"
ls -la
echo ""
echo "✅ Validation complete! Check for any ❌ items above."
