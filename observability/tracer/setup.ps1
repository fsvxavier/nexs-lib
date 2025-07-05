# Script PowerShell para configurar o ambiente de tracing no Windows

param(
    [switch]$Help,
    [switch]$Stop
)

if ($Help) {
    Write-Host "🚀 Script de configuração do ambiente de tracing distribuído" -ForegroundColor Green
    Write-Host ""
    Write-Host "Uso:"
    Write-Host "  .\setup.ps1        - Configura e inicia o ambiente"
    Write-Host "  .\setup.ps1 -Stop  - Para todos os serviços"
    Write-Host "  .\setup.ps1 -Help  - Mostra esta ajuda"
    Write-Host ""
    Write-Host "Variáveis de ambiente necessárias:"
    Write-Host "  DD_API_KEY              - Chave da API do Datadog"
    Write-Host "  NEW_RELIC_LICENSE_KEY   - Chave de licença do New Relic"
    Write-Host "  GRAFANA_PASSWORD        - Senha do admin do Grafana (padrão: admin)"
    exit 0
}

if ($Stop) {
    Write-Host "🛑 Parando serviços..." -ForegroundColor Yellow
    docker-compose down
    Write-Host "✅ Serviços parados com sucesso!" -ForegroundColor Green
    exit 0
}

Write-Host "🚀 Configurando ambiente de tracing distribuído..." -ForegroundColor Green

# Verificar se o Docker está rodando
try {
    docker info | Out-Null
    Write-Host "✅ Docker está rodando" -ForegroundColor Green
} catch {
    Write-Host "❌ Docker não está rodando. Por favor, inicie o Docker Desktop primeiro." -ForegroundColor Red
    exit 1
}

# Verificar variáveis de ambiente
if (-not $env:DD_API_KEY) {
    Write-Host "⚠️  DD_API_KEY não está definida. O Datadog Agent não funcionará corretamente." -ForegroundColor Yellow
    Write-Host "   Configure com: `$env:DD_API_KEY='your-datadog-api-key'" -ForegroundColor Yellow
}

if (-not $env:NEW_RELIC_LICENSE_KEY) {
    Write-Host "⚠️  NEW_RELIC_LICENSE_KEY não está definida. O New Relic não funcionará." -ForegroundColor Yellow
    Write-Host "   Configure com: `$env:NEW_RELIC_LICENSE_KEY='your-newrelic-license-key'" -ForegroundColor Yellow
}

# Criar diretórios necessários
Write-Host "📁 Criando estrutura de diretórios..." -ForegroundColor Cyan
New-Item -ItemType Directory -Force -Path "grafana\provisioning\dashboards" | Out-Null
New-Item -ItemType Directory -Force -Path "grafana\provisioning\datasources" | Out-Null

# Criar configuração de datasource do Grafana
$datasourceConfig = @"
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
"@

$datasourceConfig | Out-File -FilePath "grafana\provisioning\datasources\prometheus.yml" -Encoding UTF8

# Criar dashboard do Grafana
$dashboardConfig = @"
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
"@

$dashboardConfig | Out-File -FilePath "grafana\provisioning\dashboards\dashboard.yml" -Encoding UTF8

# Dashboard JSON
$dashboardJson = @"
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
"@

$dashboardJson | Out-File -FilePath "grafana\provisioning\dashboards\isis-tracer-dashboard.json" -Encoding UTF8

# Subir os serviços
Write-Host "🐳 Iniciando serviços..." -ForegroundColor Cyan
docker-compose up -d

Write-Host "⏳ Aguardando serviços ficarem prontos..." -ForegroundColor Yellow
Start-Sleep -Seconds 15

# Verificar status dos serviços
Write-Host "🔍 Verificando status dos serviços..." -ForegroundColor Cyan

$services = @("datadog-agent", "prometheus", "grafana", "jaeger", "otel-collector")
foreach ($service in $services) {
    $status = docker-compose ps $service
    if ($status -match "Up") {
        Write-Host "✅ $service está rodando" -ForegroundColor Green
    } else {
        Write-Host "❌ $service não está rodando" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "🎉 Ambiente de tracing configurado com sucesso!" -ForegroundColor Green
Write-Host ""
Write-Host "📊 Acesse os serviços:" -ForegroundColor Cyan
Write-Host "   • Grafana:     http://localhost:3000 (admin/admin)" -ForegroundColor White
Write-Host "   • Prometheus:  http://localhost:9090" -ForegroundColor White
Write-Host "   • Jaeger:      http://localhost:16686" -ForegroundColor White
Write-Host "   • App Example: http://localhost:8080" -ForegroundColor White
Write-Host ""
Write-Host "🧪 Para testar o tracer:" -ForegroundColor Cyan
Write-Host "   go run .\tracer\examples" -ForegroundColor White
Write-Host ""
Write-Host "🛑 Para parar os serviços:" -ForegroundColor Cyan
Write-Host "   .\setup.ps1 -Stop" -ForegroundColor White
Write-Host ""
Write-Host "📖 Leia o README.md para mais informações sobre uso." -ForegroundColor Cyan
