#!/bin/bash

# Simple demonstration of hooks and middleware concepts

echo "=== Conceitos Básicos: Hooks vs Middleware ==="
echo
echo "📖 HOOKS são usados para:"
echo "   • Reagir a eventos específicos (event-driven)"
echo "   • Logging, auditoria, notificações"
echo "   • Side effects que NÃO modificam o erro"
echo "   • Múltiplos hooks podem ser registrados para o mesmo evento"
echo
echo "🔧 MIDDLEWARE é usado para:"
echo "   • Transformar/enriquecer erros (processing pipeline)"
echo "   • Chain of responsibility pattern"
echo "   • Modificar o erro: adicionar metadados, contexto"
echo "   • Ordem importa: executados em sequência"
echo
echo "⚡ PIPELINE DE EXECUÇÃO:"
echo "   1. Erro criado"
echo "   2. Middleware chain executa (transforma erro)"
echo "   3. Hooks before_* executam (side effects)"
echo "   4. Processamento interno"
echo "   5. Hooks after_* executam (side effects)"
echo "   6. Erro final retornado"
echo
echo "🚀 Para executar os exemplos:"
echo "   cd examples/hooks-middleware"
echo "   go run main.go advanced.go"
echo
echo "🌐 Exemplos de i18n (Internacionalização):"
echo "   cd examples/i18n-hook         # Hook de tradução"
echo "   go run main.go"
echo "   cd examples/i18n-middleware   # Middleware de tradução"
echo "   go run main.go"
echo
echo "📚 Para mais detalhes:"
echo "   • README.md - Documentação completa"
echo "   • patterns/ - Exemplos específicos por padrão"
echo "   • main.go - Exemplos básicos"
echo "   • advanced.go - Cenários de produção"
echo "   • i18n-hook/ - Tradução usando hooks"
echo "   • i18n-middleware/ - Tradução usando middlewares"
