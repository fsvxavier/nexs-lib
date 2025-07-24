# Exemplo Provider Logrus

Este exemplo demonstra o uso completo do provider Logrus para o sistema de logging da nexs-lib, incluindo integração com loggers Logrus existentes e uso de hooks.

## Características demonstradas

- **Provider Básico**: Criação com configuração padrão
- **Provider Configurado**: Configuração personalizada com campos globais
- **Logs Formatados**: Uso de templates com printf-style
- **Logs com Códigos**: Sistema de códigos de evento para tracking
- **Enriquecimento**: Logger com campos contextuais adicionais
- **Integração**: Uso de logger Logrus existente
- **Hooks Personalizados**: Extensão com hooks nativos do Logrus
- **Múltiplos Formatos**: JSON, Text e Console
- **Buffer System**: Logging com buffer para alta performance
- **Níveis de Log**: Demonstração de todos os níveis
- **Clonagem**: Criação de instâncias independentes

## Como executar

```bash
go run main.go
```

## Características únicas do Provider Logrus

### Compatibilidade com Logrus Existente
```go
// Integra um logger Logrus já configurado
existingLogrus := logrus.New()
existingLogrus.SetLevel(logrus.WarnLevel)
provider := logrusProvider.NewProviderWithLogger(existingLogrus)
```

### Hooks Nativos do Logrus
```go
// Adiciona hooks personalizados do Logrus
hook := &CustomHook{}
provider.AddHook(hook)
```

### Acesso ao Logger Subjacente
```go
// Acessa o logger Logrus para configurações avançadas
logrusLogger := provider.GetLogrusLogger()
logrusLogger.AddHook(someHook)
```

## Vantagens da Integração

1. **Migração Facilitada**: Permite migrar gradualmente de Logrus puro
2. **Compatibilidade Total**: Mantém todos os hooks e configurações existentes
3. **Interface Unificada**: Beneficia-se da interface padronizada da nexs-lib
4. **Performance**: Mantém a performance nativa do Logrus
5. **Extensibilidade**: Suporte completo aos hooks do Logrus

## Casos de Uso Ideais

- **Migração de sistemas legados** que já usam Logrus
- **Aplicações que precisam de hooks específicos** do Logrus
- **Sistemas que requerem compatibilidade** com bibliotecas que dependem do Logrus
- **Desenvolvimento gradual** onde se quer manter funcionalidades existentes

## Configurações Disponíveis

- Todos os formatos suportados: JSON, Text, Console
- Buffer configurável para alta performance
- Níveis de log dinâmicos
- Campos globais e contextuais
- Hooks before/after para processamento customizado

## Saída Esperada

O exemplo produzirá logs em diferentes formatos demonstrando:
- Logs estruturados com campos
- Integração com logger existente
- Funcionamento de hooks personalizados
- Diferentes níveis e formatos de saída
- Buffer system em ação
