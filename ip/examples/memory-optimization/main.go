package main

import (
	"fmt"
	"net"
	"runtime"
	"time"

	"github.com/fsvxavier/nexs-lib/ip"
)

func main() {
	fmt.Println("🧠 Demonstração de Otimização de Memória")
	fmt.Println("========================================")

	// Configurar memory manager
	memConfig := ip.DefaultMemoryConfig()
	memConfig.GCPercent = 50   // GC mais agressivo
	memConfig.MaxMemoryMB = 50 // Limite de 50MB para demonstração
	memConfig.CheckInterval = 1 * time.Second
	memConfig.ForceGCThreshold = 0.8 // Force GC at 80% of limit

	memManager := ip.NewMemoryManager(memConfig)
	defer memManager.Close()

	fmt.Println("\n📊 Demonstração de Object Pooling")
	fmt.Println("==================================")
	demonstrateObjectPooling()

	fmt.Println("\n📈 Monitoramento de Memória")
	fmt.Println("============================")
	monitorMemoryUsage(memManager)

	fmt.Println("\n🔄 Lazy Loading Database")
	fmt.Println("=========================")
	demonstrateLazyLoading()

	fmt.Println("\n⚡ Performance Comparison")
	fmt.Println("=========================")
	comparePerformance()
}

func demonstrateObjectPooling() {
	fmt.Println("Teste sem object pooling:")
	testWithoutPooling()

	fmt.Println("\nTeste com object pooling:")
	testWithPooling()

	fmt.Println("\nTestando pools de slices:")
	testSlicePools()
}

func testWithoutPooling() {
	start := time.Now()
	var allocs int

	// Simular criação de muitos DetectionResult sem pooling
	for i := 0; i < 10000; i++ {
		result := &ip.DetectionResult{}
		result.IP = net.ParseIP("192.168.1.1")
		result.TrustScore = 0.8
		result.RiskLevel = "low"
		result.IsVPN = false
		allocs++

		// Simular uso do objeto
		_ = result.TrustScore
	}

	duration := time.Since(start)
	fmt.Printf("• Criados %d objetos em %v\n", allocs, duration)
	fmt.Printf("• Tempo por objeto: %v\n", duration/time.Duration(allocs))
}

func testWithPooling() {
	start := time.Now()
	var pooled int

	// Simular criação com object pooling
	for i := 0; i < 10000; i++ {
		result := ip.GetPooledDetectionResult()
		result.IP = net.ParseIP("192.168.1.1")
		result.TrustScore = 0.8
		result.RiskLevel = "low"
		result.IsVPN = false
		pooled++

		// Simular uso do objeto
		_ = result.TrustScore

		// Retornar ao pool
		ip.PutPooledDetectionResult(result)
	}

	duration := time.Since(start)
	fmt.Printf("• Processados %d objetos pooled em %v\n", pooled, duration)
	fmt.Printf("• Tempo por objeto: %v\n", duration/time.Duration(pooled))
}

func testSlicePools() {
	fmt.Println("\nTeste de slice pools:")

	// Teste string slice pool
	start := time.Now()
	for i := 0; i < 1000; i++ {
		slice := ip.GetPooledStringSlice()
		slice = append(slice, "ip1", "ip2", "ip3", "ip4", "ip5")

		// Simular processamento
		for _, s := range slice {
			_ = len(s)
		}

		ip.PutPooledStringSlice(slice)
	}
	duration := time.Since(start)
	fmt.Printf("• String slice pool: %v para 1000 operações\n", duration)

	// Teste byte slice pool
	start = time.Now()
	for i := 0; i < 1000; i++ {
		slice := ip.GetPooledByteSlice()
		slice = append(slice, []byte("some data for processing")...)

		// Simular processamento
		_ = len(slice)

		ip.PutPooledByteSlice(slice)
	}
	duration = time.Since(start)
	fmt.Printf("• Byte slice pool: %v para 1000 operações\n", duration)
}

func monitorMemoryUsage(memManager *ip.MemoryManager) {
	fmt.Println("Monitoramento inicial:")
	printMemoryStats(memManager)

	// Simular carga de memória
	fmt.Println("\nSimulando carga de memória...")
	var data [][]byte
	for i := 0; i < 1000; i++ {
		// Criar dados de teste
		chunk := make([]byte, 1024) // 1KB por chunk
		for j := range chunk {
			chunk[j] = byte(i % 256)
		}
		data = append(data, chunk)
	}

	// Aguardar um ciclo de monitoramento
	time.Sleep(2 * time.Second)

	fmt.Println("\nApós alocação de memória:")
	printMemoryStats(memManager)

	// Limpar dados e forçar GC
	data = nil
	runtime.GC()
	time.Sleep(1 * time.Second)

	fmt.Println("\nApós limpeza e GC:")
	printMemoryStats(memManager)
}

func printMemoryStats(memManager *ip.MemoryManager) {
	stats := memManager.GetMemoryStats()

	fmt.Printf("• Memória alocada: %d MB\n", stats.AllocMB)
	fmt.Printf("• Memória do sistema: %d MB\n", stats.SysMB)
	fmt.Printf("• Total alocado: %d MB\n", stats.TotalAllocMB)
	fmt.Printf("• Número de GCs: %d\n", stats.NumGC)
	fmt.Printf("• Tempo total de pausa GC: %v\n", time.Duration(stats.PauseTotalNs))
	if !stats.LastGC.IsZero() {
		fmt.Printf("• Último GC: %v atrás\n", time.Since(stats.LastGC))
	}
}

func demonstrateLazyLoading() {
	// Simular database cara de carregar
	loadFunc := func() error {
		fmt.Println("  Carregando database... (simulando operação cara)")
		time.Sleep(100 * time.Millisecond) // Simular tempo de carregamento
		return nil
	}

	// Criar lazy database com TTL de 5 segundos
	lazyDB := ip.NewLazyDatabase(loadFunc, 5*time.Second)

	fmt.Printf("• Database carregado inicialmente: %v\n", lazyDB.IsLoaded())

	// Definir dados manualmente (sem trigger do loadFunc)
	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}
	lazyDB.Set(testData)

	fmt.Printf("• Database carregado após Set: %v\n", lazyDB.IsLoaded())

	// Acessar dados
	data, err := lazyDB.Get()
	if err != nil {
		fmt.Printf("Erro ao acessar dados: %v\n", err)
		return
	}

	if dataMap, ok := data.(map[string]string); ok {
		fmt.Printf("• Dados carregados: %d entradas\n", len(dataMap))
	}

	// Unload e testar lazy loading
	lazyDB.Unload()
	fmt.Printf("• Database após unload: %v\n", lazyDB.IsLoaded())

	// Primeiro acesso deve triggerar loadFunc
	fmt.Println("• Primeiro acesso (deve carregar):")
	start := time.Now()
	_, err = lazyDB.Get()
	loadTime := time.Since(start)

	if err != nil {
		fmt.Printf("Erro: %v\n", err)
	} else {
		fmt.Printf("  Tempo de carregamento: %v\n", loadTime)
		fmt.Printf("  Database carregado: %v\n", lazyDB.IsLoaded())
	}

	// Segundo acesso deve ser instantâneo
	fmt.Println("• Segundo acesso (cache hit):")
	start = time.Now()
	_, err = lazyDB.Get()
	cacheTime := time.Since(start)

	if err != nil {
		fmt.Printf("Erro: %v\n", err)
	} else {
		fmt.Printf("  Tempo de acesso em cache: %v\n", cacheTime)
		fmt.Printf("  Speedup: %.2fx\n", float64(loadTime)/float64(cacheTime))
	}
}

func comparePerformance() {
	fmt.Println("Comparação de performance com/sem otimizações:")

	// Teste sem otimizações (mais realista)
	fmt.Println("\nSem otimizações:")
	start := time.Now()
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	// Simular processamento de IPs sem pools (cenário mais realista)
	var processedResults []string
	for i := 0; i < 5000; i++ { // Aumentar para ver benefícios
		// Criar novos objetos a cada iteração (sem reutilização)
		result := &ip.DetectionResult{}
		result.IP = net.ParseIP(fmt.Sprintf("192.168.1.%d", i%255))
		result.TrustScore = float64(i%100) / 100.0
		result.RiskLevel = []string{"low", "medium", "high"}[i%3]
		result.IsVPN = i%2 == 0
		result.IsProxy = i%3 == 0
		result.IsDatacenter = i%4 == 0

		// Simular algum processamento
		processed := fmt.Sprintf("%s:%s", result.IP.String(), result.RiskLevel)
		processedResults = append(processedResults, processed)
	}

	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)
	duration1 := time.Since(start)

	fmt.Printf("• Tempo: %v\n", duration1)
	fmt.Printf("• Allocações: %d bytes\n", memAfter.TotalAlloc-memBefore.TotalAlloc)
	fmt.Printf("• Número de allocações: %d\n", memAfter.Mallocs-memBefore.Mallocs)
	fmt.Printf("• Itens processados: %d\n", len(processedResults))

	// Limpar para o próximo teste
	processedResults = nil
	runtime.GC()
	time.Sleep(10 * time.Millisecond) // Dar tempo para o GC

	// Teste com otimizações (object pooling)
	fmt.Println("\nCom otimizações (object pooling):")
	start = time.Now()
	runtime.ReadMemStats(&memBefore)

	// Simular o mesmo trabalho com pools
	var pooledProcessedResults []string
	for i := 0; i < 5000; i++ {
		// Usar pool para reutilizar objetos
		result := ip.GetPooledDetectionResult()
		result.IP = net.ParseIP(fmt.Sprintf("192.168.1.%d", i%255))
		result.TrustScore = float64(i%100) / 100.0
		result.RiskLevel = []string{"low", "medium", "high"}[i%3]
		result.IsVPN = i%2 == 0
		result.IsProxy = i%3 == 0
		result.IsDatacenter = i%4 == 0

		// Simular o mesmo processamento
		processed := fmt.Sprintf("%s:%s", result.IP.String(), result.RiskLevel)
		pooledProcessedResults = append(pooledProcessedResults, processed)

		// IMPORTANTE: Retornar ao pool para reutilização
		ip.PutPooledDetectionResult(result)
	}

	runtime.ReadMemStats(&memAfter)
	duration2 := time.Since(start)

	fmt.Printf("• Tempo: %v\n", duration2)
	fmt.Printf("• Allocações: %d bytes\n", memAfter.TotalAlloc-memBefore.TotalAlloc)
	fmt.Printf("• Número de allocações: %d\n", memAfter.Mallocs-memBefore.Mallocs)
	fmt.Printf("• Itens processados: %d\n", len(pooledProcessedResults))

	// Calcular melhoria de forma mais precisa
	fmt.Println("\n🎯 Demonstração de cenário otimizado (alta frequência):")
	demonstrateHighFrequencyScenario()

	if duration2 <= duration1 {
		speedup := float64(duration1) / float64(duration2)
		fmt.Printf("\n� Speedup geral: %.2fx mais rápido com otimizações!\n", speedup)
	} else {
		fmt.Printf("\n💡 Object pooling mostra maior benefício em cenários de alta frequência\n")
	}
}

func demonstrateHighFrequencyScenario() {
	const iterations = 50000 // Volume alto para mostrar benefícios

	// Sem pooling
	start := time.Now()
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	for i := 0; i < iterations; i++ {
		result := &ip.DetectionResult{}
		result.TrustScore = 0.5
		result.RiskLevel = "low"
		// Simular processamento rápido
		_ = result.TrustScore > 0.3
	}

	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)
	durationWithout := time.Since(start)
	allocsWithout := memAfter.TotalAlloc - memBefore.TotalAlloc

	// Com pooling
	start = time.Now()
	runtime.ReadMemStats(&memBefore)

	for i := 0; i < iterations; i++ {
		result := ip.GetPooledDetectionResult()
		result.TrustScore = 0.5
		result.RiskLevel = "low"
		// Simular processamento rápido
		_ = result.TrustScore > 0.3
		ip.PutPooledDetectionResult(result)
	}

	runtime.ReadMemStats(&memAfter)
	durationWith := time.Since(start)
	allocsWith := memAfter.TotalAlloc - memBefore.TotalAlloc

	fmt.Printf("Alta frequência (%d objetos):\n", iterations)
	fmt.Printf("• Sem pool: %v, %d bytes\n", durationWithout, allocsWithout)
	fmt.Printf("• Com pool: %v, %d bytes\n", durationWith, allocsWith)

	if durationWith < durationWithout {
		speedup := float64(durationWithout) / float64(durationWith)
		reduction := float64(allocsWithout-allocsWith) / float64(allocsWithout) * 100
		fmt.Printf("• Speedup: %.2fx, Redução alocações: %.1f%%\n", speedup, reduction)
	}
}
