# Valkey infraestructure for Testing

Esta pasta contém a infraestrutura Docker para testes do Valkey em diferentes configurações:

## Configurações Disponíveis

### 1. Standalone
- Instância única do Valkey
- Ideal para testes básicos de funcionalidade
- Porta padrão: 6379

### 2. Cluster
- Cluster com 6 nós (3 masters + 3 replicas)
- Testa funcionalidades de clustering e sharding
- Portas: 7000-7005

### 3. Sentinel
- Configuração de alta disponibilidade
- 1 master + 2 replicas + 3 sentinels
- Testa failover automático

## Como Usar

### Iniciar todos os serviços:
```bash
docker-compose up -d
```

### Iniciar apenas standalone:
```bash
docker-compose up -d valkey-standalone
```

### Iniciar apenas cluster:
```bash
docker-compose up -d valkey-cluster-node-1 valkey-cluster-node-2 valkey-cluster-node-3 valkey-cluster-node-4 valkey-cluster-node-5 valkey-cluster-node-6
```

### Iniciar apenas sentinel:
```bash
docker-compose up -d valkey-sentinel-master valkey-sentinel-slave-1 valkey-sentinel-slave-2 valkey-sentinel-1 valkey-sentinel-2 valkey-sentinel-3
```

### Parar todos os serviços:
```bash
docker-compose down
```

### Limpar volumes (dados):
```bash
docker-compose down -v
```

## Configurações de Teste

### Standalone
- **Host**: localhost
- **Port**: 6379
- **Password**: testpass123

### Cluster
- **Hosts**: localhost:7000, localhost:7001, localhost:7002, localhost:7003, localhost:7004, localhost:7005
- **Password**: clusterpass123

### Sentinel
- **Sentinel Hosts**: localhost:26379, localhost:26380, localhost:26381
- **Master Name**: mymaster
- **Password**: sentinelpass123

## Scripts Utilitários

### Verificar status do cluster:
```bash
./scripts/check-cluster.sh
```

### Verificar status do sentinel:
```bash
./scripts/check-sentinel.sh
```

### Executar testes de integração:
```bash
./scripts/run-integration-tests.sh
```

## Logs

Para ver logs em tempo real:
```bash
# Todos os serviços
docker-compose logs -f

# Apenas standalone
docker-compose logs -f valkey-standalone

# Apenas cluster
docker-compose logs -f valkey-cluster-node-1

# Apenas sentinel
docker-compose logs -f valkey-sentinel-1
```

## Troubleshooting

### Problema: Portas já em uso
```bash
# Verificar quais portas estão em uso
sudo netstat -tulpn | grep :6379
sudo netstat -tulpn | grep :7000

# Parar processos conflitantes se necessário
```

### Problema: Cluster não forma
```bash
# Executar manualmente a formação do cluster
docker exec -it valkey-cluster-node-1 valkey-cli --cluster create \
  127.0.0.1:7000 127.0.0.1:7001 127.0.0.1:7002 \
  127.0.0.1:7003 127.0.0.1:7004 127.0.0.1:7005 \
  --cluster-replicas 1 --cluster-yes
```

### Problema: Sentinel não conecta
```bash
# Verificar configuração do sentinel
docker exec -it valkey-sentinel-1 cat /etc/valkey/sentinel.conf
```
