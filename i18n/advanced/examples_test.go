package advanced

import (
	"fmt"

	"golang.org/x/text/language"
)

func Example_numberFormatting() {
	// Basic number formatting
	locale := language.English
	number := 1234567.89
	formatted := formatNumber(number, locale)
	fmt.Printf("English: %s\n", formatted)

	// Different locale (French)
	locale = language.French
	formatted = formatNumber(number, locale)
	fmt.Printf("French: %s\n", formatted)
	// Output:
	// English: 1,234,567.89
	// French: 1 234 567,89
}

func Example_currencyFormatting() {
	// US Dollars
	locale := language.English
	amount := 1234.56
	formatted := formatCurrency(amount, "USD", locale)
	fmt.Printf("USD: %s\n", formatted)

	// Euros
	locale = language.French
	formatted = formatCurrency(amount, "EUR", locale)
	fmt.Printf("EUR: %s\n", formatted)
	// Output:
	// USD: $1,234.56
	// EUR: 1 234,56 €
}

func Example_pluralRules() {
	// English plural rules
	locale := language.English
	numbers := []float64{0, 1, 2, 5}

	fmt.Println("English plural rules:")
	for _, n := range numbers {
		form := getPluralForm(n, locale)
		fmt.Printf("%.0f items -> %s\n", n, form)
	}

	// Arabic plural rules (more complex)
	fmt.Println("\nArabic plural rules:")
	locale = language.Arabic
	for _, n := range numbers {
		form := getPluralForm(n, locale)
		fmt.Printf("%.0f items -> %s\n", n, form)
	}
	// Output:
	// English plural rules:
	// 0 items -> other
	// 1 items -> one
	// 2 items -> other
	// 5 items -> other
	//
	// Arabic plural rules:
	// 0 items -> zero
	// 1 items -> one
	// 2 items -> two
	// 5 items -> few
}
