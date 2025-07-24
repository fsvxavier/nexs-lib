# Exemplo de Copy Operations

Este exemplo demonstra as poderosas operações COPY do PostgreSQL para transferência eficiente de dados em massa.

## Funcionalidades Demonstradas

### 1. COPY FROM Básico
- Importação de dados de uma fonte
- Interface CopyFromSource customizada
- Inserção eficiente em massa

### 2. COPY TO Básico
- Exportação de dados para um destino
- Interface CopyToWriter customizada
- Streaming de resultados

### 3. COPY FROM com Dados Grandes
- Processamento de grandes volumes
- Métricas de performance
- Estatísticas automáticas

### 4. Comparação de Performance
- COPY vs INSERT individual
- Análise de speedup
- Métricas de throughput

### 5. Tratamento de Erros
- Validação de dados
- Erros de schema
- Recuperação robusta

## Conceitos Fundamentais

### COPY Operations
- **Eficiência**: 10-100x mais rápido que INSERTs individuais
- **Streaming**: Processa dados maiores que a memória disponível
- **Transacional**: Operações atômicas com rollback automático
- **Flexível**: Suporta múltiplos formatos e transformações

### Interfaces
- **CopyFromSource**: Define fonte de dados para importação
- **CopyToWriter**: Define destino para exportação
- **Streaming**: Processamento sob demanda

## Como Executar

```bash
# Certifique-se de que o PostgreSQL está rodando
cd copy/
go run main.go
```

## Exemplo de Saída

```
=== Exemplo de Copy Operations ===

1. Conectando ao banco...
2. Criando tabela de teste...
   ✅ Tabela criada com sucesso

3. Exemplo: COPY FROM básico...
   Preparando 5 registros para COPY FROM...
   Executando COPY FROM...
   ✅ COPY FROM concluído em 2ms
   📊 Linhas inseridas: 5
   📊 Total de registros na tabela: 5

4. Exemplo: COPY TO básico...
   Executando COPY TO...
   ✅ COPY TO concluído em 1ms
   📊 Linhas exportadas: 5
   Dados exportados:
     1. Ana Costa (ana@email.com) - TI - $5500.00
     2. Carlos Lima (carlos@email.com) - Marketing - $4800.00

5. Exemplo: COPY FROM com dados grandes...
   Gerando 1000 registros para teste de bulk...
   Executando COPY FROM em massa...
   ✅ COPY FROM em massa concluído em 25ms
   📊 Linhas inseridas: 1000
   📈 Taxa de inserção: 40000 linhas/segundo
   📊 Total de registros na tabela: 1005

6. Exemplo: Comparação de performance...
   Teste 1: INSERT individual (500 registros)...
   ⏱️ INSERT individual: 450ms
   Teste 2: COPY FROM (500 registros)...
   ⏱️ COPY FROM: 15ms
   📊 Análise de Performance:
   - INSERT individual: 450ms (500 registros)
   - COPY FROM: 15ms (500 registros)
   - Speedup: 30.00x mais rápido
   - INSERT rate: 1111 registros/segundo
   - COPY rate: 33333 registros/segundo
```

## Implementação das Interfaces

### CopyFromSource
```go
type TestCopyFromSource struct {
    data  [][]interface{}
    index int
}

func (s *TestCopyFromSource) Next() bool {
    return s.index < len(s.data)
}

func (s *TestCopyFromSource) Values() ([]interface{}, error) {
    if s.index >= len(s.data) {
        return nil, fmt.Errorf("no more data")
    }
    values := s.data[s.index]
    s.index++
    return values, nil
}

func (s *TestCopyFromSource) Err() error {
    return nil
}
```

### CopyToWriter
```go
type TestCopyToWriter struct {
    rows [][]interface{}
}

func (w *TestCopyToWriter) Write(row []interface{}) error {
    w.rows = append(w.rows, row)
    return nil
}

func (w *TestCopyToWriter) Close() error {
    return nil
}
```

## Casos de Uso

### 1. Importação de CSV
```go
// Importar dados de CSV
copySource := &CSVCopyFromSource{
    reader: csv.NewReader(file),
}

rowsAffected, err := conn.CopyFrom(ctx, "table_name", 
    []string{"col1", "col2", "col3"}, copySource)
```

### 2. Backup de Dados
```go
// Exportar dados para backup
backupWriter := &BackupWriter{
    file: backupFile,
}

err := conn.CopyTo(ctx, backupWriter, 
    "SELECT * FROM important_table")
```

### 3. ETL Operations
```go
// Extract, Transform, Load
transformSource := &TransformCopyFromSource{
    source: originalData,
    transformer: dataTransformer,
}

rowsAffected, err := conn.CopyFrom(ctx, "target_table", 
    columns, transformSource)
```

### 4. Data Migration
```go
// Migração entre sistemas
migrationWriter := &MigrationWriter{
    targetDB: targetConnection,
    batchSize: 1000,
}

err := conn.CopyTo(ctx, migrationWriter,
    "SELECT * FROM legacy_table")
```

## Vantagens das COPY Operations

### Performance
- **Velocidade**: 10-100x mais rápido que INSERTs individuais
- **Throughput**: Milhões de registros por segundo
- **Eficiência**: Menor overhead de CPU e memória

### Escalabilidade
- **Streaming**: Processa datasets maiores que a RAM
- **Batching**: Processamento em lotes otimizados
- **Paralelismo**: Múltiplas operações simultâneas

### Confiabilidade
- **Transacional**: Rollback automático em caso de erro
- **Validação**: Verificação de tipos e constraints
- **Atomicidade**: Todas as operações ou nenhuma

## Considerações de Performance

### Otimização
- Use transações para operações grandes
- Desabilite índices durante importação massiva
- Configure adequadamente shared_buffers e work_mem
- Use formato binário para máxima performance

### Monitoramento
- Monitore I/O e CPU durante operações
- Acompanhe locks e bloqueios
- Verifique estatísticas de buffer hits

### Troubleshooting
- Analise logs de erro detalhados
- Verifique constraints e triggers
- Monitore espaço em disco disponível

## Formatos Suportados

### CSV (Comma-Separated Values)
```go
// Configuração CSV
copyOptions := postgres.CopyOptions{
    Format: "csv",
    Header: true,
    Delimiter: ",",
}
```

### TSV (Tab-Separated Values)
```go
// Configuração TSV
copyOptions := postgres.CopyOptions{
    Format: "text",
    Delimiter: "\t",
}
```

### Binary
```go
// Configuração binária (máxima performance)
copyOptions := postgres.CopyOptions{
    Format: "binary",
}
```

## Tratamento de Erros

### Validação de Dados
- Verificação de tipos automática
- Validação de constraints
- Relatórios de erros detalhados

### Recuperação
- Rollback automático em transações
- Logs de erro para debugging
- Continuação após erros não-fatais

## Requisitos

- PostgreSQL rodando em `localhost:5432`
- Banco de dados `nexs_testdb`
- Usuário `nexs_user` com senha `nexs_password`
- Permissões de CREATE TABLE e INSERT

## Integração com Aplicações

### Processamento de Arquivos
```go
// Processar arquivo CSV grande
file, err := os.Open("data.csv")
defer file.Close()

csvSource := &CSVCopyFromSource{reader: csv.NewReader(file)}
rowsAffected, err := conn.CopyFrom(ctx, "table", columns, csvSource)
```

### APIs de Streaming
```go
// Processar dados de API
apiSource := &APICopyFromSource{
    client: httpClient,
    endpoint: "/api/data",
}

rowsAffected, err := conn.CopyFrom(ctx, "api_data", columns, apiSource)
```
