# Guia de Uso da Infraestrutura Valkey

## 📋 Pré-requisitos

- Docker e Docker Compose instalados
- Go 1.19+ para executar os testes
- Portas disponíveis:
  - **6379**: Standalone
  - **7000-7005**: Cluster (6 nós)
  - **6380-6382**: Sentinel (master + 2 slaves)
  - **26379-26381**: Sentinel (3 sentinels)

## 🚀 Início Rápido

### 1. Iniciando Todos os Serviços
```bash
cd cache/valkey/infrastructure
make start
# ou
./scripts/start.sh all
# ou
docker-compose up -d
```

### 2. Iniciando Serviços Específicos
```bash
# Apenas standalone
make standalone
./scripts/start.sh standalone

# Apenas cluster
make cluster
./scripts/start.sh cluster

# Apenas sentinel
make sentinel
./scripts/start.sh sentinel
```

### 3. Verificando Status
```bash
# Status geral
make status

# Status específico do cluster
make check-cluster

# Status específico do sentinel
make check-sentinel

# Logs em tempo real
make logs
```

### 4. Executando Testes
```bash
# Testes unitários
make test

# Testes de integração (inicia/para automaticamente)
make integration-test

# Testes manuais
docker-compose up -d
go test -v -tags=integration ./infrastructure/...
```

### 5. Parando Serviços
```bash
# Parar serviços (manter dados)
make stop

# Parar e limpar dados
make clean

# Parar, limpar dados e imagens
make clean-all
```

## 🔧 Configurações para Testes

### Standalone
```go
config := &config.Config{
    Provider: "valkey-glide",
    Hosts:    []string{"localhost:6379"},
    Password: "testpass123",
    Database: 0,
}
```

### Cluster
```go
config := &config.Config{
    Provider: "valkey-glide",
    Hosts: []string{
        "localhost:7000", "localhost:7001", "localhost:7002",
        "localhost:7003", "localhost:7004", "localhost:7005",
    },
    Password: "clusterpass123",
    Mode:     "cluster",
}
```

### Sentinel
```go
config := &config.Config{
    Provider:   "valkey-glide",
    Hosts:      []string{"localhost:26379", "localhost:26380", "localhost:26381"},
    Password:   "sentinelpass123",
    Mode:       "sentinel",
    MasterName: "mymaster",
}
```

## 🛠️ Shells Interativos

### Conectar no Standalone
```bash
make shell-standalone
# ou
docker exec -it valkey-standalone valkey-cli -a testpass123
```

### Conectar no Cluster
```bash
make shell-cluster
# ou
docker exec -it valkey-cluster-node-1 valkey-cli -c -p 7000 -a clusterpass123
```

### Conectar no Sentinel (Valkey)
```bash
make shell-sentinel
# ou
docker exec -it valkey-sentinel-master valkey-cli -a sentinelpass123
```

### Conectar no Sentinel (CLI)
```bash
make shell-sentinel-cli
# ou
docker exec -it valkey-sentinel-1 valkey-cli -p 26379
```

## 📝 Comandos Úteis

### Testar Operações Básicas

#### Standalone
```bash
# Via shell
docker exec -it valkey-standalone valkey-cli -a testpass123
> set test "hello"
> get test
> del test

# Via curl (se tiver Redis REST API)
curl -X POST "localhost:6379/set/test/hello"
curl -X GET "localhost:6379/get/test"
```

#### Cluster
```bash
# Via shell
docker exec -it valkey-cluster-node-1 valkey-cli -c -p 7000 -a clusterpass123
> set test "cluster-hello"
> get test
> cluster nodes
> cluster info
```

#### Sentinel
```bash
# Conectar no sentinel
docker exec -it valkey-sentinel-1 valkey-cli -p 26379
> sentinel masters
> sentinel slaves mymaster
> sentinel sentinels mymaster

# Conectar no master
docker exec -it valkey-sentinel-master valkey-cli -a sentinelpass123
> set test "sentinel-hello"
> get test
```

## 🐛 Troubleshooting

### Problema: Portas em Uso
```bash
# Verificar portas em uso
sudo netstat -tulpn | grep -E ':(6379|7000|26379)'

# Parar processos conflitantes
sudo kill -9 $(sudo lsof -t -i:6379)
```

### Problema: Cluster Não Forma
```bash
# Verificar logs
docker-compose logs valkey-cluster-setup

# Recriar cluster manualmente
docker exec -it valkey-cluster-node-1 valkey-cli --cluster create \
  127.0.0.1:7000 127.0.0.1:7001 127.0.0.1:7002 \
  127.0.0.1:7003 127.0.0.1:7004 127.0.0.1:7005 \
  --cluster-replicas 1 --cluster-yes
```

### Problema: Sentinel Não Conecta
```bash
# Verificar configuração
docker exec -it valkey-sentinel-1 cat /etc/valkey/sentinel.conf

# Verificar logs
docker-compose logs valkey-sentinel-1

# Testar conectividade
docker exec -it valkey-sentinel-1 valkey-cli -p 26379 ping
```

### Problema: Falha de Autenticação
```bash
# Verificar senhas nos logs
docker-compose logs | grep -i auth

# Verificar configuração
docker exec -it valkey-standalone valkey-cli --help
```

## 📊 Monitoramento

### Recursos do Sistema
```bash
# Uso de CPU/Memória
docker stats $(docker ps --format "{{.Names}}" | grep valkey)

# Espaço em disco
docker system df
```

### Métricas do Valkey
```bash
# Info do standalone
docker exec valkey-standalone valkey-cli -a testpass123 info

# Info do cluster
docker exec valkey-cluster-node-1 valkey-cli -p 7000 -a clusterpass123 info

# Info do sentinel
docker exec valkey-sentinel-master valkey-cli -a sentinelpass123 info
```

## 🔄 Desenvolvimento

### Modificar Configurações
1. Edite os arquivos em `config/`
2. Reinicie os serviços: `make restart`

### Adicionar Novos Testes
1. Crie arquivos `*_test.go` com tag `//go:build integration`
2. Use as configurações em `test_config.go`
3. Execute com `make integration-test`

### Personalizar Ambiente
1. Copie `.env.example` para `.env`
2. Modifique as variáveis conforme necessário
3. Reinicie os serviços
