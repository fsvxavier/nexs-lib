package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/fsvxavier/nexs-lib/ip"
)

func main() {
	fmt.Println("🔍 Demonstração de Detecção Avançada - VPN/Proxy e ASN Lookup")
	fmt.Println("============================================================")

	// Configurar detector avançado
	config := ip.DefaultDetectorConfig()
	config.CacheEnabled = true
	config.CacheTimeout = 10 * time.Minute
	config.MaxWorkers = 5

	detector := ip.NewAdvancedDetector(config)
	defer detector.Close()

	// Carregar database de VPN de exemplo
	err := loadSampleVPNDatabase(detector)
	if err != nil {
		log.Printf("Erro ao carregar database VPN: %v", err)
	}

	// Carregar database ASN de exemplo
	err = loadSampleASNDatabase(detector)
	if err != nil {
		log.Printf("Erro ao carregar database ASN: %v", err)
	}

	// IPs de teste
	testIPs := []string{
		"8.8.8.8",      // Google DNS - Clean
		"1.1.1.1",      // Cloudflare DNS - Clean
		"52.86.85.143", // AWS - Datacenter
		"1.2.3.4",      // VPN simulado (carregado no database)
		"5.6.7.8",      // Proxy simulado
		"192.168.1.1",  // IP privado
		"127.0.0.1",    // Loopback
	}

	fmt.Println("\n📊 Análise Detalhada dos IPs:")
	fmt.Println("===============================")

	ctx := context.Background()

	for _, ipStr := range testIPs {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			fmt.Printf("❌ IP inválido: %s\n", ipStr)
			continue
		}

		// Realizar detecção avançada
		result, err := detector.DetectAdvanced(ctx, ip)
		if err != nil {
			fmt.Printf("❌ Erro na detecção de %s: %v\n", ipStr, err)
			continue
		}

		// Exibir resultados
		displayDetectionResult(ipStr, result)
	}

	// Demonstrar estatísticas de cache
	fmt.Println("\n📈 Estatísticas do Cache:")
	fmt.Println("=========================")
	size, hitRate := detector.GetCacheStats()
	fmt.Printf("• Entradas em cache: %d\n", size)
	fmt.Printf("• Hit rate: %.2f%%\n", hitRate*100)

	// Demonstrar detecção concorrente
	fmt.Println("\n⚡ Detecção Concorrente:")
	fmt.Println("========================")
	demonstrateConcurrentDetection(ctx, detector, testIPs)
}

func loadSampleVPNDatabase(detector *ip.AdvancedDetector) error {
	// Database VPN de exemplo em formato CSV
	csvData := `ip,name,type,reliability
1.2.3.4,ExpressVPN,commercial,0.9
5.6.7.8,ProxyService,proxy,0.6
9.10.11.12,TorExit,tor,0.8
10.0.0.1,PrivateVPN,commercial,0.85
100.100.100.100,DatacenterProxy,datacenter,0.7`

	reader := strings.NewReader(csvData)
	return detector.LoadVPNDatabase(reader)
}

func loadSampleASNDatabase(detector *ip.AdvancedDetector) error {
	// Database ASN de exemplo em formato CSV
	csvData := `asn,name,country,type,is_cloud_provider,cloud_provider
16509,Amazon Web Services,US,hosting,true,AWS
15169,Google LLC,US,hosting,true,Google Cloud
8075,Microsoft Corporation,US,hosting,true,Azure
13335,Cloudflare Inc,US,hosting,true,Cloudflare
12345,Example ISP,US,isp,false,
54321,Example Hosting,UK,hosting,false,`

	reader := strings.NewReader(csvData)
	return detector.LoadASNDatabase(reader)
}

func displayDetectionResult(ipStr string, result *ip.DetectionResult) {
	fmt.Printf("\n🌐 IP: %s\n", ipStr)
	fmt.Printf("   ├─ Trust Score: %.2f/1.0\n", result.TrustScore)
	fmt.Printf("   ├─ Risk Level: %s\n", getRiskEmoji(result.RiskLevel)+result.RiskLevel)
	fmt.Printf("   ├─ Detection Time: %v\n", result.DetectionTime)

	// Características detectadas
	var characteristics []string
	if result.IsVPN {
		characteristics = append(characteristics, "🔒 VPN")
	}
	if result.IsProxy {
		characteristics = append(characteristics, "🌐 Proxy")
	}
	if result.IsTor {
		characteristics = append(characteristics, "🧅 Tor")
	}
	if result.IsDatacenter {
		characteristics = append(characteristics, "🏢 Datacenter")
	}
	if result.IsCloudProvider {
		characteristics = append(characteristics, "☁️ Cloud Provider")
	}

	if len(characteristics) > 0 {
		fmt.Printf("   ├─ Características: %s\n", strings.Join(characteristics, ", "))
	} else {
		fmt.Printf("   ├─ Características: ✅ Clean IP\n")
	}

	// Informações do provedor VPN
	if result.VPNProvider != nil {
		fmt.Printf("   ├─ VPN Provider: %s (%s, reliability: %.2f)\n",
			result.VPNProvider.Name,
			result.VPNProvider.Type,
			result.VPNProvider.Reliability)
	}

	// Informações ASN
	if result.ASNInfo != nil {
		fmt.Printf("   └─ ASN: AS%d - %s (%s, %s)\n",
			result.ASNInfo.ASN,
			result.ASNInfo.Name,
			result.ASNInfo.Country,
			result.ASNInfo.Type)
		if result.ASNInfo.IsCloudProvider {
			fmt.Printf("      └─ Cloud Provider: %s\n", result.ASNInfo.CloudProvider)
		}
	} else {
		fmt.Printf("   └─ ASN: Not available\n")
	}
}

func getRiskEmoji(riskLevel string) string {
	switch riskLevel {
	case "low":
		return "🟢 "
	case "medium":
		return "🟡 "
	case "high":
		return "🟠 "
	case "critical":
		return "🔴 "
	default:
		return "⚪ "
	}
}

func demonstrateConcurrentDetection(ctx context.Context, detector *ip.AdvancedDetector, ips []string) {
	processor := ip.NewConcurrentIPProcessor(detector, 3)
	defer processor.Close()

	start := time.Now()
	resultChan := processor.ProcessIPs(ctx, ips)

	var results []ip.IPProcessResult
	for result := range resultChan {
		results = append(results, result)
	}
	totalTime := time.Since(start)

	fmt.Printf("• Processados %d IPs em %v\n", len(results), totalTime)
	fmt.Printf("• Tempo médio por IP: %v\n", totalTime/time.Duration(len(results)))

	// Contar resultados por risk level
	riskCounts := make(map[string]int)
	for _, result := range results {
		if result.Error == nil && result.Detection != nil {
			riskCounts[result.Detection.RiskLevel]++
		}
	}

	fmt.Printf("• Distribuição de risco:\n")
	for risk, count := range riskCounts {
		fmt.Printf("  - %s%s: %d IPs\n", getRiskEmoji(risk), risk, count)
	}
}
