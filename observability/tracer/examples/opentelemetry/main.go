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
	fmt.Println("🔭 Exemplo OpenTelemetry OTLP")
	fmt.Println("=============================")

	// Configuração para OpenTelemetry genérico
	cfg := interfaces.Config{
		ServiceName:  "otel-example-service",
		Environment:  "development",
		ExporterType: "opentelemetry",
		Endpoint:     "http://otel-collector:4318/v1/traces", // HTTP/protobuf
		// Alternativas:
		// "otel-collector:4317"                           // gRPC
		// "http://jaeger:14268/api/traces"                // Jaeger HTTP
		// "jaeger:14250"                                  // Jaeger gRPC
		SamplingRatio: 1.0, // 100% sampling para desenvolvimento
		Version:       "1.0.0",
		Propagators:   []string{"tracecontext", "b3", "baggage"},
		Headers: map[string]string{
			"Authorization": "Bearer your-token", // Para endpoints autenticados
			"X-Tenant-ID":   "tenant-123",        // Para multi-tenancy
		},
		Insecure: true, // Para desenvolvimento (sem TLS)
		Attributes: map[string]string{
			"team":        "platform",
			"environment": "development",
			"cluster":     "dev-k8s",
			"namespace":   "default",
			"deployment":  "otel-example",
		},
	}

	// Validar configuração
	if err := config.Validate(cfg); err != nil {
		log.Fatalf("❌ Erro na configuração: %v", err)
	}

	// Inicializar TracerManager
	tracerManager := tracer.NewTracerManager()
	ctx := context.Background()

	fmt.Println("📡 Inicializando OpenTelemetry tracer...")
	tracerProvider, err := tracerManager.Init(ctx, cfg)
	if err != nil {
		log.Fatalf("❌ Erro ao inicializar tracer: %v", err)
	}

	// Configurar como tracer global (opcional)
	otel.SetTracerProvider(tracerProvider)
	fmt.Println("✅ OpenTelemetry tracer configurado globalmente")

	// Obter tracer para este serviço
	tr := tracerProvider.Tracer("otel-example")

	// Exemplo de operação com tracing
	runKubernetesWorkflow(ctx, tr)

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
	fmt.Println("   - Jaeger UI: http://jaeger:16686")
	fmt.Println("   - Zipkin UI: http://zipkin:9411")
	fmt.Println("   - Grafana: http://grafana:3000")
}

func runKubernetesWorkflow(ctx context.Context, tracer trace.Tracer) {
	// Criar span principal para workflow Kubernetes
	ctx, span := tracer.Start(ctx, "k8s-deployment-workflow")
	defer span.End()

	deploymentName := "web-app-v2"
	namespace := "production"
	replicas := 3

	span.SetAttributes(
		attribute.String("k8s.deployment.name", deploymentName),
		attribute.String("k8s.namespace", namespace),
		attribute.Int("k8s.replicas.desired", replicas),
		attribute.String("k8s.cluster", "prod-cluster"),
		attribute.String("workflow.type", "deployment"),
	)

	fmt.Println("☸️ Iniciando workflow de deployment Kubernetes...")

	// Validar manifesto
	if !validateManifest(ctx, tracer, deploymentName) {
		span.SetStatus(codes.Error, "Manifesto inválido")
		return
	}

	// Criar recursos
	if !createResources(ctx, tracer, deploymentName, namespace) {
		span.SetStatus(codes.Error, "Falha na criação de recursos")
		return
	}

	// Aguardar pods ficarem ready
	if !waitForPods(ctx, tracer, deploymentName, namespace, replicas) {
		span.SetStatus(codes.Error, "Pods não ficaram ready")
		return
	}

	// Configurar health checks
	configureHealthChecks(ctx, tracer, deploymentName, namespace)

	// Configurar service mesh
	configureServiceMesh(ctx, tracer, deploymentName, namespace)

	// Executar testes de smoke
	if !runSmokeTests(ctx, tracer, deploymentName, namespace) {
		span.SetStatus(codes.Error, "Testes de smoke falharam")
		return
	}

	// Rollout completion
	span.SetAttributes(
		attribute.String("deployment.status", "completed"),
		attribute.Bool("deployment.success", true),
	)

	span.SetStatus(codes.Ok, "Deployment concluído com sucesso")
	fmt.Println("✅ Workflow de deployment concluído")
}

func validateManifest(ctx context.Context, tracer trace.Tracer, deploymentName string) bool {
	ctx, span := tracer.Start(ctx, "k8s.validate-manifest")
	defer span.End()

	span.SetAttributes(
		attribute.String("k8s.resource.type", "deployment"),
		attribute.String("k8s.resource.name", deploymentName),
		attribute.String("validation.tool", "kubeval"),
	)

	// Simular validação do manifesto
	time.Sleep(100 * time.Millisecond)

	// Validar labels, annotations, resources, etc.
	validateLabels(ctx, tracer)
	validateResources(ctx, tracer)
	validateSecurityContext(ctx, tracer)

	fmt.Printf("✅ Manifesto para %s validado\n", deploymentName)
	return true
}

func validateLabels(ctx context.Context, tracer trace.Tracer) {
	ctx, span := tracer.Start(ctx, "validation.labels")
	defer span.End()

	span.SetAttributes(
		attribute.String("validation.type", "labels"),
		attribute.Bool("labels.app", true),
		attribute.Bool("labels.version", true),
		attribute.Bool("labels.component", true),
	)

	time.Sleep(20 * time.Millisecond)
	fmt.Println("🏷️ Labels validados")
}

func validateResources(ctx context.Context, tracer trace.Tracer) {
	ctx, span := tracer.Start(ctx, "validation.resources")
	defer span.End()

	span.SetAttributes(
		attribute.String("validation.type", "resources"),
		attribute.String("cpu.request", "100m"),
		attribute.String("cpu.limit", "500m"),
		attribute.String("memory.request", "128Mi"),
		attribute.String("memory.limit", "512Mi"),
	)

	time.Sleep(30 * time.Millisecond)
	fmt.Println("💾 Recursos validados")
}

func validateSecurityContext(ctx context.Context, tracer trace.Tracer) {
	ctx, span := tracer.Start(ctx, "validation.security-context")
	defer span.End()

	span.SetAttributes(
		attribute.String("validation.type", "security"),
		attribute.Bool("run_as_non_root", true),
		attribute.Bool("read_only_root_filesystem", true),
		attribute.Int("run_as_user", 1000),
	)

	time.Sleep(25 * time.Millisecond)
	fmt.Println("🔒 Security context validado")
}

func createResources(ctx context.Context, tracer trace.Tracer, deploymentName, namespace string) bool {
	ctx, span := tracer.Start(ctx, "k8s.create-resources")
	defer span.End()

	span.SetAttributes(
		attribute.String("k8s.deployment.name", deploymentName),
		attribute.String("k8s.namespace", namespace),
	)

	// Criar deployment
	createDeployment(ctx, tracer, deploymentName, namespace)

	// Criar service
	createService(ctx, tracer, deploymentName, namespace)

	// Criar configmap
	createConfigMap(ctx, tracer, deploymentName, namespace)

	// Criar secret
	createSecret(ctx, tracer, deploymentName, namespace)

	fmt.Printf("📦 Recursos criados para %s no namespace %s\n", deploymentName, namespace)
	return true
}

func createDeployment(ctx context.Context, tracer trace.Tracer, name, namespace string) {
	ctx, span := tracer.Start(ctx, "k8s.create-deployment")
	defer span.End()

	span.SetAttributes(
		attribute.String("k8s.resource.type", "deployment"),
		attribute.String("k8s.resource.name", name),
		attribute.String("k8s.namespace", namespace),
		attribute.String("k8s.api_version", "apps/v1"),
	)

	time.Sleep(200 * time.Millisecond)
	fmt.Printf("🚀 Deployment %s criado\n", name)
}

func createService(ctx context.Context, tracer trace.Tracer, name, namespace string) {
	ctx, span := tracer.Start(ctx, "k8s.create-service")
	defer span.End()

	span.SetAttributes(
		attribute.String("k8s.resource.type", "service"),
		attribute.String("k8s.resource.name", name+"-svc"),
		attribute.String("k8s.namespace", namespace),
		attribute.String("service.type", "ClusterIP"),
		attribute.Int("service.port", 80),
	)

	time.Sleep(50 * time.Millisecond)
	fmt.Printf("🌐 Service %s-svc criado\n", name)
}

func createConfigMap(ctx context.Context, tracer trace.Tracer, name, namespace string) {
	ctx, span := tracer.Start(ctx, "k8s.create-configmap")
	defer span.End()

	span.SetAttributes(
		attribute.String("k8s.resource.type", "configmap"),
		attribute.String("k8s.resource.name", name+"-config"),
		attribute.String("k8s.namespace", namespace),
		attribute.Int("config.keys_count", 5),
	)

	time.Sleep(30 * time.Millisecond)
	fmt.Printf("⚙️ ConfigMap %s-config criado\n", name)
}

func createSecret(ctx context.Context, tracer trace.Tracer, name, namespace string) {
	ctx, span := tracer.Start(ctx, "k8s.create-secret")
	defer span.End()

	span.SetAttributes(
		attribute.String("k8s.resource.type", "secret"),
		attribute.String("k8s.resource.name", name+"-secret"),
		attribute.String("k8s.namespace", namespace),
		attribute.String("secret.type", "Opaque"),
	)

	time.Sleep(40 * time.Millisecond)
	fmt.Printf("🔐 Secret %s-secret criado\n", name)
}

func waitForPods(ctx context.Context, tracer trace.Tracer, deploymentName, namespace string, replicas int) bool {
	ctx, span := tracer.Start(ctx, "k8s.wait-for-pods")
	defer span.End()

	span.SetAttributes(
		attribute.String("k8s.deployment.name", deploymentName),
		attribute.String("k8s.namespace", namespace),
		attribute.Int("k8s.replicas.desired", replicas),
	)

	// Simular verificação de pods
	for i := 1; i <= replicas; i++ {
		checkPodStatus(ctx, tracer, fmt.Sprintf("%s-%d", deploymentName, i), namespace)
	}

	span.SetAttributes(
		attribute.Int("k8s.replicas.ready", replicas),
		attribute.Bool("pods.all_ready", true),
	)

	fmt.Printf("✅ Todos os %d pods estão ready\n", replicas)
	return true
}

func checkPodStatus(ctx context.Context, tracer trace.Tracer, podName, namespace string) {
	ctx, span := tracer.Start(ctx, "k8s.check-pod-status")
	defer span.End()

	span.SetAttributes(
		attribute.String("k8s.pod.name", podName),
		attribute.String("k8s.namespace", namespace),
		attribute.String("pod.phase", "Running"),
		attribute.Bool("pod.ready", true),
	)

	// Simular verificação de readiness
	time.Sleep(300 * time.Millisecond)
	fmt.Printf("✅ Pod %s está ready\n", podName)
}

func configureHealthChecks(ctx context.Context, tracer trace.Tracer, deploymentName, namespace string) {
	ctx, span := tracer.Start(ctx, "k8s.configure-health-checks")
	defer span.End()

	span.SetAttributes(
		attribute.String("k8s.deployment.name", deploymentName),
		attribute.String("k8s.namespace", namespace),
		attribute.String("health_check.liveness", "/health/live"),
		attribute.String("health_check.readiness", "/health/ready"),
		attribute.Int("health_check.initial_delay", 30),
		attribute.Int("health_check.period", 10),
	)

	time.Sleep(100 * time.Millisecond)
	fmt.Printf("🏥 Health checks configurados para %s\n", deploymentName)
}

func configureServiceMesh(ctx context.Context, tracer trace.Tracer, deploymentName, namespace string) {
	ctx, span := tracer.Start(ctx, "service-mesh.configure")
	defer span.End()

	span.SetAttributes(
		attribute.String("k8s.deployment.name", deploymentName),
		attribute.String("k8s.namespace", namespace),
		attribute.String("service_mesh.type", "istio"),
		attribute.Bool("service_mesh.sidecar_injection", true),
		attribute.String("service_mesh.version", "1.15.0"),
	)

	// Configurar virtual service
	configureVirtualService(ctx, tracer, deploymentName, namespace)

	// Configurar destination rule
	configureDestinationRule(ctx, tracer, deploymentName, namespace)

	time.Sleep(150 * time.Millisecond)
	fmt.Printf("🕸️ Service mesh configurado para %s\n", deploymentName)
}

func configureVirtualService(ctx context.Context, tracer trace.Tracer, name, namespace string) {
	ctx, span := tracer.Start(ctx, "istio.configure-virtual-service")
	defer span.End()

	span.SetAttributes(
		attribute.String("istio.resource.type", "VirtualService"),
		attribute.String("istio.resource.name", name+"-vs"),
		attribute.String("k8s.namespace", namespace),
	)

	time.Sleep(50 * time.Millisecond)
	fmt.Printf("🛣️ VirtualService %s-vs configurado\n", name)
}

func configureDestinationRule(ctx context.Context, tracer trace.Tracer, name, namespace string) {
	ctx, span := tracer.Start(ctx, "istio.configure-destination-rule")
	defer span.End()

	span.SetAttributes(
		attribute.String("istio.resource.type", "DestinationRule"),
		attribute.String("istio.resource.name", name+"-dr"),
		attribute.String("k8s.namespace", namespace),
		attribute.String("load_balancer.type", "ROUND_ROBIN"),
	)

	time.Sleep(40 * time.Millisecond)
	fmt.Printf("🎯 DestinationRule %s-dr configurado\n", name)
}

func runSmokeTests(ctx context.Context, tracer trace.Tracer, deploymentName, namespace string) bool {
	ctx, span := tracer.Start(ctx, "testing.smoke-tests")
	defer span.End()

	span.SetAttributes(
		attribute.String("k8s.deployment.name", deploymentName),
		attribute.String("k8s.namespace", namespace),
		attribute.String("test.type", "smoke"),
		attribute.Int("test.count", 3),
	)

	// Executar múltiplos testes
	tests := []string{"health-check", "api-endpoint", "database-connection"}

	for _, test := range tests {
		runSingleTest(ctx, tracer, test, deploymentName)
	}

	span.SetAttributes(
		attribute.Int("test.passed", 3),
		attribute.Int("test.failed", 0),
		attribute.Bool("test.all_passed", true),
	)

	fmt.Printf("🧪 Testes de smoke executados para %s\n", deploymentName)
	return true
}

func runSingleTest(ctx context.Context, tracer trace.Tracer, testName, deploymentName string) {
	ctx, span := tracer.Start(ctx, fmt.Sprintf("test.%s", testName))
	defer span.End()

	span.SetAttributes(
		attribute.String("test.name", testName),
		attribute.String("k8s.deployment.name", deploymentName),
		attribute.String("test.result", "passed"),
	)

	time.Sleep(200 * time.Millisecond)
	fmt.Printf("✅ Teste %s passou\n", testName)
}
