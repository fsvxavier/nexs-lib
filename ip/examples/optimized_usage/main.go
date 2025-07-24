package main

import (
	"fmt"
	"net/http"

	"github.com/fsvxavier/nexs-lib/ip"
)

func main() {
	// Exemplo de uso das funções otimizadas por padrão
	fmt.Println("=== Exemplo: Funções IP Otimizadas ===")
	fmt.Println()

	// Simular uma requisição HTTP
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.195, 192.168.1.1, 10.0.0.1")
	req.Header.Set("X-Real-IP", "203.0.113.195")
	req.Header.Set("CF-Connecting-IP", "203.0.113.100")
	req.RemoteAddr = "192.168.1.1:8080"

	// 1. Extrair IP real (agora otimizado por padrão)
	clientIP := ip.GetRealIP(req)
	fmt.Printf("🔍 IP Real do Cliente: %s\n", clientIP)

	// 2. Extrair informações detalhadas (agora otimizado por padrão)
	ipInfo := ip.GetRealIPInfo(req)
	if ipInfo != nil {
		fmt.Printf("📊 Informações do IP:\n")
		fmt.Printf("   - IP: %s\n", ipInfo.IP)
		fmt.Printf("   - Tipo: %s\n", ipInfo.Type.String())
		fmt.Printf("   - IPv4: %v\n", ipInfo.IsIPv4)
		fmt.Printf("   - IPv6: %v\n", ipInfo.IsIPv6)
		fmt.Printf("   - Público: %v\n", ipInfo.IsPublic)
		fmt.Printf("   - Privado: %v\n", ipInfo.IsPrivate)
		fmt.Printf("   - Fonte: %s\n", ipInfo.Source)
	}

	// 3. Extrair cadeia completa de IPs (agora otimizado por padrão)
	ipChain := ip.GetIPChain(req)
	fmt.Printf("\n🔗 Cadeia de IPs:\n")
	for i, chainIP := range ipChain {
		fmt.Printf("   %d. %s\n", i+1, chainIP)
	}

	// 4. Demonstrar cache de parsing
	fmt.Printf("\n🚀 Demonstração de Cache:\n")

	// Primeira chamada - popula o cache
	ipInfo1 := ip.ParseIP("203.0.113.195")
	fmt.Printf("   Primeira chamada: %s (tipo: %s)\n", ipInfo1.IP, ipInfo1.Type.String())

	// Segunda chamada - usa o cache
	ipInfo2 := ip.ParseIP("203.0.113.195")
	fmt.Printf("   Segunda chamada: %s (tipo: %s)\n", ipInfo2.IP, ipInfo2.Type.String())

	// 5. Estatísticas do cache
	cacheSize, maxSize := ip.GetCacheStats()
	fmt.Printf("\n📈 Estatísticas do Cache:\n")
	fmt.Printf("   - Entradas atuais: %d\n", cacheSize)
	fmt.Printf("   - Tamanho máximo: %d\n", maxSize)

	// 6. Frameworks suportados
	frameworks := ip.GetSupportedFrameworks()
	fmt.Printf("\n🔧 Frameworks Suportados:\n")
	for i, framework := range frameworks {
		fmt.Printf("   %d. %s\n", i+1, framework)
	}

	// 7. Teste de performance com diferentes tipos de IP
	fmt.Printf("\n⚡ Teste de Performance:\n")

	testIPs := []string{
		"203.0.113.195", // IP público
		"192.168.1.1",   // IP privado
		"127.0.0.1",     // Loopback
		"2001:db8::1",   // IPv6 público
		"::1",           // IPv6 loopback
		"10.0.0.1",      // IP privado
		"172.16.0.1",    // IP privado
	}

	for _, testIP := range testIPs {
		info := ip.ParseIP(testIP)
		if info != nil && info.IP != nil {
			fmt.Printf("   - %s → %s\n", testIP, info.Type.String())
		}
	}

	fmt.Printf("\n✅ Todas as operações usam otimizações zero-allocation por padrão!\n")
	fmt.Printf("   • Pool de buffers para parsing\n")
	fmt.Printf("   • Cache de resultados\n")
	fmt.Printf("   • Operações de string otimizadas\n")
	fmt.Printf("   • Gerenciamento de memória eficiente\n")
}
