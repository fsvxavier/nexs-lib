# Exemplo de Operações Batch

Este exemplo demonstra como usar operações batch para melhorar significativamente a performance de operações em massa no PostgreSQL.

## Funcionalidades Demonstradas

### 1. Batch Básico
- Criação de batch com múltiplas operações
- Execução de operações INSERT em lote
- Processamento de resultados

### 2. Batch com Transação
- Combinação de batch com transações
- Operações atômicas em lote
- Commit/rollback baseado em resultados

### 3. Comparação de Performance
- Inserções individuais vs. batch
- Métricas de performance detalhadas
- Análise de speedup

### 4. Tratamento de Erros
- Gerenciamento de erros em operações batch
- Continuidade de execução mesmo com falhas
- Relatórios detalhados de sucesso/falha

## Vantagens do Batch

- **Performance**: 5-10x mais rápido que operações individuais
- **Atomicidade**: Todas as operações em uma transação
- **Eficiência**: Menor overhead de rede e comunicação
- **Escalabilidade**: Ideal para operações em massa

## Como Executar

```bash
# Certifique-se de que o PostgreSQL está rodando
cd batch/
go run main.go
```

## Requisitos

- PostgreSQL rodando em `localhost:5432`
- Banco de dados `nexs_testdb`
- Usuário `nexs_user` com senha `nexs_password`

## Exemplo de Saída

```
=== Exemplo de Operações Batch ===

1. Conectando ao banco...
2. Criando tabela de teste...
3. Limpando dados anteriores...

4. Exemplo: Operações batch básicas...
   Adicionando 5 produtos ao batch...
   Batch criado com 5 operações
   Executando batch...
   ✅ Batch concluído em 15ms
   📊 Resultado: 5 sucessos, 0 falhas
   
5. Exemplo: Batch com transação...
   Iniciando transação...
   Executando batch na transação...
   ✅ Transação commitada com sucesso
   
6. Exemplo: Comparação de performance...
   Teste 1: Inserções individuais...
   ⏱️ Inserções individuais: 150ms (100 registros)
   Teste 2: Inserção em batch...
   ⏱️ Inserção em batch: 20ms (100 registros)
   📊 Análise de Performance:
   🚀 Speedup: 7.5x mais rápido
```

## Dicas de Otimização

1. **Tamanho do Batch**: 100-1000 operações por batch
2. **Transações**: Use transações para operações relacionadas
3. **Tratamento de Erros**: Implemente tratamento robusto de erros
4. **Monitoramento**: Monitore métricas de performance
5. **Memória**: Considere o uso de memória para batches grandes
