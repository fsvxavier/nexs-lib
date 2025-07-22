# Detecção Avançada - VPN/Proxy e ASN Lookup

Este exemplo demonstra as capacidades avançadas de detecção do módulo IP, incluindo identificação de VPNs, proxies, Tor e informações de ASN.

## 🎯 Funcionalidades Demonstradas

### VPN/Proxy Detection
- Database customizável de IPs de VPN conhecidos
- Heurísticas para detecção de proxy
- Score de confiabilidade do IP (0.0-1.0)
- Identificação de redes Tor

### ASN Lookup
- Identificação de ISP/hosting provider
- Detecção de cloud providers (AWS, Google Cloud, Azure)
- Classificação de tipos de rede (ISP, hosting, government)

### Trust Score & Risk Assessment
- Cálculo automático de trust score baseado em múltiplos fatores
- Classificação de nível de risco (low/medium/high/critical)
- Tempo de detecção para monitoramento de performance

## 🚀 Como Executar

```bash
cd /path/to/nexs-lib/ip/examples/advanced-detection
go run main.go
```

## 📊 Exemplo de Saída

```
🔍 Demonstração de Detecção Avançada - VPN/Proxy e ASN Lookup
============================================================

📊 Análise Detalhada dos IPs:
===============================

🌐 IP: 8.8.8.8
   ├─ Trust Score: 1.00/1.0
   ├─ Risk Level: 🟢 low
   ├─ Detection Time: 2.1ms
   ├─ Características: ✅ Clean IP
   └─ ASN: AS15169 - Google LLC (US, hosting)
      └─ Cloud Provider: Google Cloud

🌐 IP: 52.86.85.143
   ├─ Trust Score: 0.70/1.0
   ├─ Risk Level: 🟡 medium
   ├─ Detection Time: 1.8ms
   ├─ Características: 🏢 Datacenter, ☁️ Cloud Provider
   └─ ASN: AS16509 - Amazon Web Services (US, hosting)
      └─ Cloud Provider: AWS

⚡ Detecção Concorrente:
========================
• Processados 7 IPs em 12.3ms
• Tempo médio por IP: 1.7ms
• Distribuição de risco:
  - 🟢 low: 4 IPs
  - 🟡 medium: 2 IPs
  - 🟠 high: 1 IPs
```

## 🔧 Configuração

O exemplo utiliza configurações otimizadas:

```go
config := ip.DefaultDetectorConfig()
config.CacheEnabled = true        // Cache para melhor performance
config.CacheTimeout = 10 * time.Minute
config.MaxWorkers = 5             // Pool de workers para concorrência
```

## 📈 Performance

- **Detecção Individual**: ~2.4μs por IP
- **Detecção Concorrente**: ~133μs para 5 IPs em paralelo
- **Cache Hit**: ~6% mais rápido que detecção fresh
- **Memory Usage**: Object pooling reduz allocations em ~82%

## 🎛️ Customização

### Database VPN Personalizado

```go
// Formato CSV: ip,name,type,reliability
csvData := `1.2.3.4,ExpressVPN,commercial,0.9
5.6.7.8,ProxyService,proxy,0.6`

reader := strings.NewReader(csvData)
detector.LoadVPNDatabase(reader)
```

### Database ASN Personalizado

```go
// Formato CSV: asn,name,country,type,is_cloud_provider,cloud_provider
csvData := `16509,Amazon Web Services,US,hosting,true,AWS
15169,Google LLC,US,hosting,true,Google Cloud`

reader := strings.NewReader(csvData)
detector.LoadASNDatabase(reader)
```

## 🔒 Casos de Uso

1. **Fraud Prevention**: Detectar conexões suspeitas através de VPNs/proxies
2. **Geo-compliance**: Identificar tentativas de bypass geográfico
3. **Bot Detection**: Detectar tráfego automatizado através de datacenters
4. **Risk Scoring**: Avaliar confiabilidade de conexões em tempo real
5. **Security Monitoring**: Monitorar padrões de acesso suspeitos
