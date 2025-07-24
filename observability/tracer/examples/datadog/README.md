# Exemplo Datadog APM

Este exemplo demonstra como usar a biblioteca de tracer com Datadog APM.

## Pré-requisitos

1. Conta no Datadog
2. API Key do Datadog
3. Datadog Agent rodando (opcional para desenvolvimento)

## Configuração

### 1. Variáveis de Ambiente

```bash
export TRACER_SERVICE_NAME="datadog-example-service"
export TRACER_ENVIRONMENT="development"
export TRACER_EXPORTER_TYPE="datadog"
export DATADOG_API_KEY="your-datadog-api-key"
export TRACER_SAMPLING_RATIO="1.0"
```

### 2. Arquivo .env

```env
TRACER_SERVICE_NAME=datadog-example-service
TRACER_ENVIRONMENT=development
TRACER_EXPORTER_TYPE=datadog
DATADOG_API_KEY=your-datadog-api-key
TRACER_SAMPLING_RATIO=1.0
```

## Executar o Exemplo

```bash
# Na pasta do exemplo
go run main.go
```

## O que o Exemplo Faz

1. **Configuração**: Define configuração para Datadog APM
2. **Inicialização**: Cria e configura o TracerManager
3. **Tracer Global**: Define o tracer como global no OpenTelemetry
4. **Operações de Negócio**: Executa operações com instrumentação
5. **Sub-spans**: Cria spans aninhados para diferentes operações
6. **Atributos**: Adiciona metadados aos spans
7. **Shutdown**: Finaliza graciosamente o tracer

## Estrutura dos Traces

```
business-operation (span raiz)
├── process-item (item 1)
│   └── database-query
├── process-item (item 2)  
│   └── database-query
└── process-item (item 3)
    └── database-query
```

## Visualizar Traces

Após executar o exemplo, os traces estarão disponíveis em:
- **Datadog APM**: https://app.datadoghq.com/apm/traces

## Atributos Incluídos

- `operation.type`: Tipo da operação
- `user.id`: ID do usuário
- `batch.size`: Tamanho do batch
- `item.id`: ID do item processado
- `item.status`: Status do processamento
- `db.system`: Sistema de banco de dados
- `db.operation`: Operação SQL
- `db.table`: Tabela consultada
- `db.rows_affected`: Linhas afetadas

## Troubleshooting

### Datadog Agent não está rodando
Se você não tem o Datadog Agent local, o exemplo ainda funciona mas enviará traces diretamente para a API do Datadog usando a API key.

### API Key inválida
Verifique se a API key está correta e tem permissões para enviar traces.

### Traces não aparecem
- Verifique a API key
- Confirme se o serviço está no ambiente correto
- Aguarde alguns minutos para propagação
