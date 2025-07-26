package main

import (
	"fmt"
	"log"

	"github.com/fsvxavier/nexs-lib/decimal"
	"github.com/fsvxavier/nexs-lib/decimal/interfaces"
)

func main() {
	fmt.Println("=== Decimal Library Enhanced Features Demo ===")
	fmt.Println()

	// 1. Basic decimal operations
	fmt.Println("1. Basic Operations:")
	price, _ := decimal.NewFromString("99.99")
	tax, _ := decimal.NewFromFloat(0.08)

	taxAmount, _ := price.Mul(tax)
	total, _ := price.Add(taxAmount)

	fmt.Printf("   Price: $%s\n", price.String())
	fmt.Printf("   Tax Amount: $%s\n", taxAmount.String())
	fmt.Printf("   Total: $%s\n\n", total.String())

	// 2. Edge cases demonstration
	fmt.Println("2. Edge Cases:")

	// Small numbers
	small1, _ := decimal.NewFromString("0.001")
	small2, _ := decimal.NewFromString("0.002")
	smallSum, _ := small1.Add(small2)
	fmt.Printf("   Small numbers: %s + %s = %s\n", small1.String(), small2.String(), smallSum.String())

	// Large numbers
	large1, _ := decimal.NewFromString("123456789.123456")
	large2, _ := decimal.NewFromString("987654321.654321")
	largeSum, _ := large1.Add(large2)
	fmt.Printf("   Large numbers: %s + %s = %s\n", large1.String(), large2.String(), largeSum.String())

	// Scientific notation
	scientific, _ := decimal.NewFromString("1.5e3")
	fmt.Printf("   Scientific notation: 1.5e3 = %s\n\n", scientific.String())

	// 3. Performance optimized batch operations
	fmt.Println("3. Performance Optimized Batch Operations:")

	// Create sample data
	salesData := []string{
		"1250.00",
		"980.50",
		"1425.75",
		"756.25",
		"2100.00",
		"890.75",
		"1650.50",
	}

	// Convert to decimal slice for batch operations
	salesDecimals := make([]interfaces.Decimal, len(salesData))
	for i, s := range salesData {
		salesDecimals[i] = mustDecimal(s)
	}

	// Traditional approach (separate operations)
	fmt.Println("   Traditional approach:")
	sum, _ := decimal.SumSlice(salesDecimals)
	avg, _ := decimal.AverageSlice(salesDecimals)
	max, _ := decimal.MaxSlice(salesDecimals)
	min, _ := decimal.MinSlice(salesDecimals)

	fmt.Printf("     Sum: $%s\n", sum.String())
	fmt.Printf("     Average: $%s\n", avg.String())
	fmt.Printf("     Max: $%s\n", max.String())
	fmt.Printf("     Min: $%s\n", min.String())

	// Optimized approach (single pass)
	fmt.Println("   Optimized approach (single pass):")
	stats, err := decimal.ProcessBatchSlice(salesDecimals)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("     Sum: $%s\n", stats.Sum.String())
	fmt.Printf("     Average: $%s\n", stats.Average.String())
	fmt.Printf("     Max: $%s\n", stats.Max.String())
	fmt.Printf("     Min: $%s\n", stats.Min.String())
	fmt.Printf("     Count: %d transactions\n\n", stats.Count)

	// 4. Real-world financial example
	fmt.Println("4. Real-world Financial Analysis:")

	monthlyRevenues := []interfaces.Decimal{
		mustDecimal("45000.00"), // January
		mustDecimal("52000.00"), // February
		mustDecimal("38000.00"), // March
		mustDecimal("41000.00"), // April
		mustDecimal("48000.00"), // May
		mustDecimal("55000.00"), // June
		mustDecimal("62000.00"), // July
		mustDecimal("58000.00"), // August
		mustDecimal("47000.00"), // September
		mustDecimal("49000.00"), // October
		mustDecimal("53000.00"), // November
		mustDecimal("61000.00"), // December
	}

	yearStats, err := decimal.ProcessBatchSlice(monthlyRevenues)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("   Annual Revenue Analysis:\n")
	fmt.Printf("     Total Revenue: $%s\n", yearStats.Sum.String())
	fmt.Printf("     Average Monthly: $%s\n", yearStats.Average.String())
	fmt.Printf("     Best Month: $%s\n", yearStats.Max.String())
	fmt.Printf("     Worst Month: $%s\n", yearStats.Min.String())
	fmt.Printf("     Months Analyzed: %d\n\n", yearStats.Count)

	// 5. Provider switching for different use cases
	fmt.Println("5. Provider Management:")
	manager := decimal.NewManager(nil)

	fmt.Printf("   Current provider: %s\n", manager.GetProvider().Name())

	// Switch to shopspring for performance-critical operations
	err = manager.SwitchProvider("shopspring")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Switched to: %s (for performance)\n", manager.GetProvider().Name())

	// Switch back to cockroach for high precision
	err = manager.SwitchProvider("cockroach")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   Switched back to: %s (for precision)\n\n", manager.GetProvider().Name())

	fmt.Println("=== Demo Complete ===")
	fmt.Println("\nKey Improvements Demonstrated:")
	fmt.Println("✅ Enhanced edge case handling")
	fmt.Println("✅ Performance optimized batch operations")
	fmt.Println("✅ Comprehensive GoDoc documentation")
	fmt.Println("✅ Real-world usage examples")
}

// Helper function to create decimals (panics on error for demo purposes)
func mustDecimal(s string) interfaces.Decimal {
	d, err := decimal.NewFromString(s)
	if err != nil {
		panic(err)
	}
	return d
}
