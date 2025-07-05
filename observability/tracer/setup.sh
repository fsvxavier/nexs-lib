#!/bin/bash

# Script para configurar e executar o ambiente de tracing

set -e

echo "🚀 Configurando ambiente de tracing distribuído..."

# Verificar se o Docker está rodando
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker não está rodando. Por favor, inicie o Docker primeiro."
    exit 1
fi

# Verificar se as variáveis de ambiente estão definidas
if [ -z "$DD_API_KEY" ]; then
    echo "⚠️  DD_API_KEY não está definida. O Datadog Agent não funcionará corretamente."
    echo "   Configure com: export DD_API_KEY=your-datadog-api-key"
fi

if [ -z "$NEW_RELIC_LICENSE_KEY" ]; then
    echo "⚠️  NEW_RELIC_LICENSE_KEY não está definida. O New Relic não funcionará."
    echo "   Configure com: export NEW_RELIC_LICENSE_KEY=your-newrelic-license-key"
fi

# Criar diretórios necessários
mkdir -p grafana/provisioning/{dashboards,datasources}

# Criar configuração de datasource do Grafana
cat > grafana/provisioning/datasources/prometheus.yml << EOF
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
  - name: Jaeger
    type: jaeger
    access: proxy
    url: http://jaeger:16686
EOF

# Criar dashboard do Grafana para métricas do tracer
cat > grafana/provisioning/dashboards/dashboard.yml << EOF
apiVersion: 1

providers:
  - name: 'isis-tracer'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    editable: true
    options:
      path: /etc/grafana/provisioning/dashboards
EOF

# Dashboard JSON para o tracer
cat > grafana/provisioning/dashboards/isis-tracer-dashboard.json << 'EOF'
{
  "dashboard": {
    "id": null,
    "title": "ISIS Tracer Metrics",
    "tags": ["isis", "tracer", "golang"],
    "timezone": "browser",
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "panels": [
      {
        "id": 1,
        "title": "Total Spans",
        "type": "stat",
        "targets": [
          {
            "expr": "rate(tracer_spans_total[5m])",
            "legendFormat": "Spans/sec"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 0}
      },
      {
        "id": 2,
        "title": "Span Duration",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(tracer_spans_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          },
          {
            "expr": "histogram_quantile(0.50, rate(tracer_spans_duration_seconds_bucket[5m]))",
            "legendFormat": "50th percentile"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 0}
      },
      {
        "id": 3,
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(tracer_spans_errors_total[5m])",
            "legendFormat": "Errors/sec"
          }
        ],
        "gridPos": {"h": 8, "w": 24, "x": 0, "y": 8}
      }
    ],
    "refresh": "5s",
    "schemaVersion": 16,
    "version": 0
  }
}
EOF

echo "📁 Criando estrutura de diretórios..."

# Subir os serviços
echo "🐳 Iniciando serviços..."
docker-compose up -d

echo "⏳ Aguardando serviços ficarem prontos..."
sleep 10

# Verificar se os serviços estão rodando
echo "🔍 Verificando status dos serviços..."

services=("datadog-agent" "prometheus" "grafana" "jaeger" "otel-collector")
for service in "${services[@]}"; do
    if docker-compose ps $service | grep -q "Up"; then
        echo "✅ $service está rodando"
    else
        echo "❌ $service não está rodando"
    fi
done

echo ""
echo "🎉 Ambiente de tracing configurado com sucesso!"
echo ""
echo "📊 Acesse os serviços:"
echo "   • Grafana:    http://localhost:3000 (admin/admin)"
echo "   • Prometheus: http://localhost:9090"
echo "   • Jaeger:     http://localhost:16686"
echo "   • App Example: http://localhost:8080"
echo ""
echo "🧪 Para testar o tracer:"
echo "   go run ./tracer/examples"
echo ""
echo "🛑 Para parar os serviços:"
echo "   docker-compose down"
echo ""
echo "📖 Leia o README.md para mais informações sobre uso."
