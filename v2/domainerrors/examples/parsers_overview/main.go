// Package main demonstra o uso básico do sistema de parsers expandido
package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("=== Sistema de Parsers Expandido - Visão Geral ===")
	fmt.Println()

	demonstrateParserTypes()
	demonstrateRegistryFeatures()
	demonstratePluginArchitecture()
	demonstrateFactoryPattern()
}

func demonstrateParserTypes() {
	fmt.Println("--- Tipos de Parsers Implementados ---")
	fmt.Println()

	parsers := map[string]string{
		"grpc_http_parsers.go": `
• GRPCErrorParser: Parseia erros de gRPC com mapeamento de codes.Code
• HTTPErrorParser: Parseia erros HTTP com análise de status codes
• Suporte a timeout, connection errors, status codes específicos`,

		"nosql_cloud_parsers.go": `
• RedisErrorParser: Parseia erros Redis (timeout, wrong type, auth, etc.)
• MongoDBErrorParser: Parseia erros MongoDB com códigos específicos
• AWSErrorParser: Parseia erros AWS com detecção de throttling`,

		"postgresql_pgx_parsers.go": `
• PostgreSQLErrorParser: Parser aprimorado para PostgreSQL
• PGXErrorParser: Parser específico para driver PGX
• Suporte a constraint violations, connection issues, syntax errors`,
	}

	for file, description := range parsers {
		fmt.Printf("📁 %s%s\n\n", file, description)
	}
}

func demonstrateRegistryFeatures() {
	fmt.Println("--- Funcionalidades do Registry Distribuído ---")
	fmt.Println()

	features := []string{
		"🔄 Sistema de prioridade para parsing de erros",
		"📊 Métricas de performance e contadores",
		"🏥 Health checks para parsers individuais",
		"⚙️ Configuração dinâmica de parsers",
		"🔌 Suporte a plugins customizados",
		"🔍 Logging e observabilidade integrados",
		"⚡ Parsing concorrente e thread-safe",
		"📈 Sistema de fallback para parsers",
	}

	for _, feature := range features {
		fmt.Println(feature)
	}
	fmt.Println()
}

func demonstratePluginArchitecture() {
	fmt.Println("--- Arquitetura de Plugins ---")
	fmt.Println()

	fmt.Printf(`
🔌 ParserPlugin Interface:
   • Name() - Nome único do plugin
   • Version() - Versão do plugin
   • Description() - Descrição funcional
   • CreateParser() - Factory de parsers
   • ValidateConfig() - Validação de configuração
   • DefaultConfig() - Configuração padrão

🔧 GenericDatabasePlugin:
   • Parser genérico para bancos de dados
   • Configuração via patterns e códigos
   • Suporte a regex e keywords
   • Validação automática de configuração

🏭 CustomParserFactory:
   • Factory para tipos customizados
   • Registro dinâmico de parsers
   • Suporte a regex_matcher e keyword_matcher
   • Configuração flexível via JSON
`)
	fmt.Println()
}

func demonstrateFactoryPattern() {
	fmt.Println("--- Padrão Factory e Configuração ---")
	fmt.Println()

	fmt.Printf(`
⚙️ Configuração Dinâmica:
   • Padrões regex personalizados
   • Keywords customizadas
   • Códigos de erro específicos
   • Severidade e categoria configuráveis

🔧 Exemplo de Configuração:
{
  "patterns": ["timeout", "connection", "auth"],
  "error_codes": {
    "timeout": "CUSTOM_TIMEOUT",
    "connection": "CUSTOM_CONNECTION_ERROR"
  },
  "default_severity": "medium",
  "category": "infrastructure"
}

📊 Tipos de Error Suportados:
   • ErrorTypeCloud - Erros de serviços cloud
   • ErrorTypeBusiness - Erros de negócio
   • ErrorTypeTechnical - Erros técnicos
   • ErrorTypeInfrastructure - Erros de infraestrutura
`)
	fmt.Println()
}

func init() {
	// Configuração de log para o exemplo
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
