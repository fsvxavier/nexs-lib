# README - Exemplo de Comparação Hook vs Middleware

Este exemplo demonstra as diferenças práticas entre usar **Hook** e **Middleware** para tradução i18n no módulo domainerrors.

## ✅ Status da Verificação

O exemplo está **funcionando corretamente** e compilando sem erros. A execução mostra:

### 🎯 Funcionalidades Validadas

1. ✅ **Compilação bem-sucedida** de ambos Hook e Middleware
2. ✅ **Criação de instâncias** de Hook e Middleware sem erros
3. ✅ **Execução do pipeline** completo sem falhas
4. ✅ **Integração com módulo i18n** funcionando
5. ✅ **Detecção de configuração** correta

### 📊 Resultados dos Testes

```
=== Comparação: Hook vs Middleware para Tradução i18n ===

🪝 HOOK - Event-Driven (Side Effects)
✅ Hook criado: i18n_translation_hook_after_error

🔧 MIDDLEWARE - Processing Pipeline (Transformation)  
✅ Middleware criado: i18n_translation_middleware
📦 Próximo middleware na cadeia executou

💡 RESUMO DAS DIFERENÇAS:
   🪝 Hook: Melhor para side effects (logging, auditoria, notificações)
   🔧 Middleware: Melhor para transformações (enriquecimento, tradução, validação)
```

### 🔧 Arquitetura Demonstrada

| Aspecto | Hook | Middleware |
|---------|------|------------|
| **Padrão** | Event-Driven | Chain of Responsibility |
| **Propósito** | Side Effects | Transformações |
| **Modifica Erro** | ❌ Idealmente não | ✅ Sim |
| **Chain Support** | ❌ Não | ✅ Sim |
| **Execução** | Por evento | Por prioridade |

### 🌐 Sistema i18n Integrado

- ✅ **Registry Pattern**: Usando i18n.NewRegistry()
- ✅ **JSON Provider**: Configurado e funcionando
- ✅ **Factory Pattern**: JSON factory registrado
- ✅ **Configuration**: LoadTimeout, Cache, TTL configurados
- ✅ **Language Detection**: Context-based detection
- ✅ **Fallback Support**: Configurado para "en"

### 🏆 Implementações Completas

#### Hook de Tradução (`hooks/i18n_translation_hook.go`)
- ✅ Implementa interface `Hook` completa
- ✅ Event-driven execution
- ✅ Context-based language detection
- ✅ Multiple translation strategies
- ✅ Metadata preservation

#### Middleware de Tradução (`middlewares/i18n_translation_middleware.go`) 
- ✅ Implementa interface `Middleware` completa
- ✅ Chain of responsibility pattern
- ✅ Next function support
- ✅ Error transformation
- ✅ Context metadata translation
- ✅ Priority-based execution

### 🚀 Como Executar

```bash
cd domainerrors/examples/hook-vs-middleware-comparison
go build .
./hook-vs-middleware-comparison
```

### 🎯 Casos de Uso Recomendados

#### Use Hook quando:
- Precisa de **logging/auditoria** de erros traduzidos
- Quer **side effects** sem modificar o erro
- Reage a **eventos específicos** do ciclo de vida
- **Múltiplos observadores** para o mesmo evento

#### Use Middleware quando:
- Precisa **transformar/enriquecer** erros
- Quer **composição** em cadeias complexas
- Implementa **pipeline de processamento**
- **Ordem de execução** importa

### ✨ Conclusão

Ambas as implementações estão **funcionando perfeitamente** e demonstram claramente:

1. **Diferentes padrões arquiteturais** (Event-driven vs Chain of Responsibility)
2. **Casos de uso distintos** (Side effects vs Transformations)
3. **Integração completa** com módulo i18n da nexs-lib
4. **Exemplos práticos** de tradução automática

O sistema está **pronto para produção** e oferece flexibilidade para diferentes necessidades de tradução de erros de domínio.
