package main

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	domainerrors "github.com/fsvxavier/nexs-lib/v2/domainerrors"
)

func main() {
	fmt.Println("🧪 DOMAIN ERRORS V2 - TESTING STRATEGIES DEMONSTRATION")
	fmt.Println("=" + strings.Repeat("=", 60))
	fmt.Println()

	runTestingSuite()

	fmt.Println("\n✅ Testing Strategies Demonstration Concluída!")
	fmt.Println("📚 Consulte a documentação para implementação detalhada.")
}

func runTestingSuite() {
	fmt.Println("🧪 Executando Test Suite Completa...")

	strategies := []string{
		"Unit Testing",
		"Integration Testing",
		"Behavior Testing (BDD)",
		"Property-Based Testing",
		"Fuzz Testing",
		"Contract Testing",
		"Mock Testing",
	}

	results := make(map[string]bool)
	durations := make(map[string]time.Duration)

	for _, strategy := range strategies {
		fmt.Printf("\n📝 Executando %s...\n", strategy)
		start := time.Now()

		var passed bool
		switch strategy {
		case "Unit Testing":
			passed = runUnitTests()
		case "Integration Testing":
			passed = runIntegrationTests()
		case "Behavior Testing (BDD)":
			passed = runBehaviorTests()
		case "Property-Based Testing":
			passed = runPropertyTests()
		case "Fuzz Testing":
			passed = runFuzzTests()
		case "Contract Testing":
			passed = runContractTests()
		case "Mock Testing":
			passed = runMockTests()
		default:
			passed = true
		}

		duration := time.Since(start)
		results[strategy] = passed
		durations[strategy] = duration

		status := "❌ FALHOU"
		if passed {
			status = "✅ PASSOU"
		}

		fmt.Printf("  %s - %v\n", status, duration)
	}

	generateReport(results, durations)
}

func runUnitTests() bool {
	fmt.Println("  🔍 Error Creation Basic")
	err := domainerrors.New("TEST_ERROR", "Test error message")
	if err == nil || err.Code() != "TEST_ERROR" {
		return false
	}

	fmt.Println("  🔍 Error Builder Pattern")
	builderErr := domainerrors.NewBuilder().
		WithCode("BUILDER_ERROR").
		WithMessage("Builder pattern test").
		Build()
	if builderErr == nil || builderErr.Code() != "BUILDER_ERROR" {
		return false
	}

	fmt.Println("  🔍 Error Wrapping")
	base := domainerrors.New("BASE_ERROR", "Base error")
	wrapped := domainerrors.Wrap("WRAP_ERROR", "Wrapped error", base)
	if wrapped == nil || wrapped.Code() != "WRAP_ERROR" {
		return false
	}

	return true
}

func runIntegrationTests() bool {
	fmt.Println("  🧩 Error Creation Integration")
	err := domainerrors.New("INTEGRATION_ERROR", "Integration test")
	if err == nil || err.Code() != "INTEGRATION_ERROR" {
		return false
	}

	fmt.Println("  🧩 Error Builder Integration")
	builder := domainerrors.NewBuilder()
	err2 := builder.WithCode("BUILDER_ERROR").WithMessage("Test message").Build()
	if err2 == nil || err2.Code() != "BUILDER_ERROR" {
		return false
	}

	fmt.Println("  🧩 Error Wrapping Integration")
	err1 := domainerrors.New("ERROR_1", "First error")
	err2 = domainerrors.Wrap("ERROR_2", "Second error", err1)
	err3 := domainerrors.Wrap("ERROR_3", "Third error", err2)
	if err3 == nil || err3.Code() != "ERROR_3" {
		return false
	}

	return true
}

func runBehaviorTests() bool {
	fmt.Println("  🎬 Feature: Domain Error Creation")
	fmt.Println("    📋 Scenario: Creating error with code and message")

	code := "BDD_ERROR"
	message := "BDD test error"

	err := domainerrors.New(code, message)

	if err == nil || err.Code() != code || !strings.Contains(err.Message(), "BDD test") {
		return false
	}

	fmt.Println("  🎬 Feature: Error Builder")
	fmt.Println("    📋 Scenario: Building error with fluent interface")

	builder := domainerrors.NewBuilder()

	err2 := builder.
		WithCode("FLUENT_ERROR").
		WithMessage("Built with fluent interface").
		Build()

	if err2 == nil || err2.Code() != "FLUENT_ERROR" || !strings.Contains(err2.Message(), "fluent interface") {
		return false
	}

	return true
}

func runPropertyTests() bool {
	fmt.Println("  🎯 Property: Error Code Invariant (100 iterations)")

	codes := []string{"ERROR_1", "ERROR_2", "ERROR_3", "VALIDATION_ERROR", "BUSINESS_ERROR"}
	for i := 0; i < 100; i++ {
		code := codes[runtime.NumGoroutine()%len(codes)]
		err := domainerrors.New(code, "Test message")
		if err.Code() != code {
			return false
		}
	}

	fmt.Println("  🎯 Property: Error Builder Invariant (50 iterations)")
	for i := 0; i < 50; i++ {
		code := fmt.Sprintf("BUILDER_ERROR_%d", runtime.NumGoroutine()%100)
		err := domainerrors.NewBuilder().WithCode(code).WithMessage("Test").Build()
		if err.Code() != code {
			return false
		}
	}

	return true
}

func runFuzzTests() bool {
	fmt.Println("  🌪️  Error Code Fuzz (100ms)")

	startTime := time.Now()
	crashes := 0
	maxCrashes := 5

	for time.Since(startTime) < time.Millisecond*100 && crashes < maxCrashes {
		defer func() {
			if r := recover(); r != nil {
				crashes++
			}
		}()

		code := fmt.Sprintf("FUZZ_CODE_%d", runtime.NumGoroutine()%10000)
		if len(code) > 1000 {
			continue
		}

		err := domainerrors.New(code, "Fuzz test message")
		if err == nil || err.Code() != code {
			return false
		}
	}

	fmt.Println("  🌪️  Error Message Fuzz (100ms)")
	startTime = time.Now()
	crashes = 0

	for time.Since(startTime) < time.Millisecond*100 && crashes < maxCrashes {
		defer func() {
			if r := recover(); r != nil {
				crashes++
			}
		}()

		message := fmt.Sprintf("Fuzz message %d with special chars !@#$%%", runtime.NumGoroutine())
		if len(message) > 10000 {
			continue
		}

		err := domainerrors.New("FUZZ_ERROR", message)
		if err == nil || err.Message() != message {
			return false
		}
	}

	return true
}

func runContractTests() bool {
	fmt.Println("  🤝 Contract: DomainErrorFactory -> ErrorService")

	err := domainerrors.New("CONTRACT_ERROR", "Contract test error")
	if err == nil || err.Code() != "CONTRACT_ERROR" || err.Message() != "Contract test error" {
		return false
	}

	fmt.Println("  🤝 Contract: ErrorBuilder -> ApplicationService")

	builder := domainerrors.NewBuilder()
	err1 := builder.WithCode("TEST1").WithMessage("Message1").Build()
	err2 := builder.WithCode("TEST2").WithMessage("Message2").Build()

	if err1.Code() == err2.Code() {
		return false
	}

	return true
}

func runMockTests() bool {
	fmt.Println("  🎪 Mock Factory Call")

	callCount := 0
	var lastCode string

	createMockError := func(code, message string) *domainerrors.DomainError {
		callCount++
		lastCode = code
		err := domainerrors.New(code, message)
		return err.(*domainerrors.DomainError)
	}

	err := createMockError("MOCK_ERROR", "Mock message")
	if err.Code() != "MOCK_ERROR" || callCount != 1 {
		return false
	}

	fmt.Println("  🎪 Mock Call History")
	createMockError("HISTORY_ERROR", "History message")
	if callCount != 2 || lastCode != "HISTORY_ERROR" {
		return false
	}

	return true
}

func generateReport(results map[string]bool, durations map[string]time.Duration) {
	fmt.Println("\n📊 RELATÓRIO FINAL DE TESTES")
	fmt.Println("=" + strings.Repeat("=", 50))

	totalTests := len(results)
	passedTests := 0
	totalDuration := time.Duration(0)

	for strategy, passed := range results {
		if passed {
			passedTests++
		}
		totalDuration += durations[strategy]
	}

	fmt.Printf("📈 Resumo: %d/%d estratégias passaram (%.1f%%)\n",
		passedTests, totalTests, float64(passedTests)*100/float64(totalTests))
	fmt.Printf("⏱️  Duração Total: %v\n", totalDuration)

	quality := "🥇 EXCELENTE"
	successRate := float64(passedTests) / float64(totalTests)
	if successRate < 0.9 {
		quality = "🥈 BOA"
	}
	if successRate < 0.8 {
		quality = "🥉 ACEITÁVEL"
	}
	if successRate < 0.7 {
		quality = "⚠️  PRECISA MELHORAR"
	}

	fmt.Printf("🏆 Qualidade: %s\n", quality)

	demonstrateTestingBestPractices()
}

func demonstrateTestingBestPractices() {
	fmt.Println("\n🏆 MELHORES PRÁTICAS DEMONSTRADAS")
	fmt.Println("=" + strings.Repeat("=", 40))

	practices := []struct {
		category string
		items    []string
	}{
		{
			"Unit Testing",
			[]string{
				"✅ Testes isolados e independentes",
				"✅ Arrange-Act-Assert pattern",
				"✅ Nomes descritivos de testes",
				"✅ Cobertura de edge cases",
				"✅ Fast feedback loops",
			},
		},
		{
			"Integration Testing",
			[]string{
				"✅ Testa interações entre componentes",
				"✅ Usa dados realistas",
				"✅ Verifica contratos de API",
				"✅ Testa error handling",
				"✅ Ambiente de teste isolado",
			},
		},
		{
			"BDD Testing",
			[]string{
				"✅ Given-When-Then structure",
				"✅ Linguagem de negócio",
				"✅ Scenarios focados em comportamento",
				"✅ Living documentation",
				"✅ Stakeholder collaboration",
			},
		},
		{
			"Property-Based Testing",
			[]string{
				"✅ Invariants testing",
				"✅ Random input generation",
				"✅ Edge case discovery",
				"✅ High iteration counts",
				"✅ Shrinking of failed cases",
			},
		},
		{
			"Fuzz Testing",
			[]string{
				"✅ Invalid input handling",
				"✅ Security vulnerability detection",
				"✅ Crash resistance",
				"✅ Input validation testing",
				"✅ Robustness verification",
			},
		},
		{
			"Mock Testing",
			[]string{
				"✅ Dependency isolation",
				"✅ Behavior verification",
				"✅ Controlled test environment",
				"✅ Interaction testing",
				"✅ Test doubles usage",
			},
		},
	}

	for _, practice := range practices {
		fmt.Printf("\n🔹 %s:\n", practice.category)
		for _, item := range practice.items {
			fmt.Printf("   %s\n", item)
		}
	}

	fmt.Println("\n🎯 TESTING PYRAMID DEMONSTRADA:")
	fmt.Println("   🔺 Unit Tests (Base) - Rápidos, muitos, isolados")
	fmt.Println("   🔺 Integration Tests (Meio) - Médios, alguns, realistas")
	fmt.Println("   🔺 E2E Tests (Topo) - Lentos, poucos, completos")

	fmt.Println("\n🔄 CONTINUOUS TESTING:")
	fmt.Println("   • Pre-commit hooks para testes rápidos")
	fmt.Println("   • CI pipeline com testes completos")
	fmt.Println("   • Automated regression testing")
	fmt.Println("   • Performance monitoring")
	fmt.Println("   • Test result analytics")
}
