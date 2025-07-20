package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/fsvxavier/nexs-lib/observability/tracer"
	"github.com/fsvxavier/nexs-lib/observability/tracer/config"
	"github.com/fsvxavier/nexs-lib/observability/tracer/interfaces"
)

func main() {
	fmt.Println("🐕 Exemplo Datadog APM")
	fmt.Println("=====================")

	// Configuração para Datadog APM
	cfg := interfaces.Config{
		ServiceName:   "datadog-example-service",
		Environment:   "development",
		ExporterType:  "datadog",
		APIKey:        "your-datadog-api-key", // Configure em .env ou variável de ambiente
		SamplingRatio: 1.0,                    // 100% sampling para desenvolvimento
		Version:       "1.0.0",
		Propagators:   []string{"tracecontext", "b3", "datadog"},
		Attributes: map[string]string{
			"team":        "platform",
			"environment": "development",
		},
	}

	// Validar configuração
	if err := config.Validate(cfg); err != nil {
		log.Fatalf("❌ Erro na configuração: %v", err)
	}

	// Inicializar TracerManager
	tracerManager := tracer.NewTracerManager()
	ctx := context.Background()

	fmt.Println("📡 Inicializando Datadog tracer...")
	tracerProvider, err := tracerManager.Init(ctx, cfg)
	if err != nil {
		log.Fatalf("❌ Erro ao inicializar tracer: %v", err)
	}

	// Configurar como tracer global (opcional)
	otel.SetTracerProvider(tracerProvider)
	fmt.Println("✅ Datadog tracer configurado globalmente")

	// Obter tracer para este serviço
	tr := tracerProvider.Tracer("datadog-example")

	// Exemplo de operação com tracing
	runBusinessOperation(ctx, tr)

	// Aguardar um pouco para envio dos dados
	fmt.Println("⏳ Aguardando envio de traces...")
	time.Sleep(2 * time.Second)

	// Shutdown graceful
	fmt.Println("🔄 Fazendo shutdown do tracer...")
	if err := tracerManager.Shutdown(ctx); err != nil {
		log.Printf("⚠️ Erro no shutdown: %v", err)
	}

	fmt.Println("✅ Exemplo concluído!")
	fmt.Println("\n📊 Verifique os traces em:")
	fmt.Println("   https://app.datadoghq.com/apm/traces")
}

func runBusinessOperation(ctx context.Context, tracer trace.Tracer) {
	// Criar span principal
	ctx, span := tracer.Start(ctx, "business-operation")
	defer span.End()

	span.SetAttributes(
		attribute.String("operation.type", "example"),
		attribute.String("user.id", "user-123"),
		attribute.Int("batch.size", 10),
	)

	fmt.Println("🚀 Executando operação de negócio...")

	// Simular sub-operações
	for i := 0; i < 3; i++ {
		processItem(ctx, tracer, i+1)
	}

	// Simular possível erro
	if false { // Mude para true para testar erro
		span.RecordError(fmt.Errorf("erro simulado na operação"))
		span.SetStatus(codes.Error, "Operação falhou")
	} else {
		span.SetStatus(codes.Ok, "Operação concluída com sucesso")
	}

	fmt.Println("✅ Operação de negócio concluída")
}

func processItem(ctx context.Context, tracer trace.Tracer, itemID int) {
	ctx, span := tracer.Start(ctx, "process-item")
	defer span.End()

	span.SetAttributes(
		attribute.Int("item.id", itemID),
		attribute.String("item.status", "processing"),
	)

	// Simular processamento
	time.Sleep(50 * time.Millisecond)

	// Simular consulta ao banco
	queryDatabase(ctx, tracer, itemID)

	span.SetAttributes(
		attribute.String("item.status", "completed"),
	)

	fmt.Printf("📦 Item %d processado\n", itemID)
}

func queryDatabase(ctx context.Context, tracer trace.Tracer, itemID int) {
	ctx, span := tracer.Start(ctx, "database-query")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "items"),
		attribute.Int("db.rows_affected", 1),
	)

	// Simular consulta
	time.Sleep(20 * time.Millisecond)

	fmt.Printf("🗃️ Consulta ao banco para item %d\n", itemID)
}
