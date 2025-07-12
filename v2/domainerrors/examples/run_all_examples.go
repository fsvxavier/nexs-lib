package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	fmt.Println("🚀 Running all Domain Errors v2 Examples")
	fmt.Println("=========================================")

	examples := []string{
		"basic",
		"builder-pattern",
		"error-stacking",
	}

	for _, example := range examples {
		fmt.Printf("\n📁 Running example: %s\n", example)
		fmt.Println(strings.Repeat("-", 40))

		examplePath := filepath.Join(".", example)
		if _, err := os.Stat(examplePath); os.IsNotExist(err) {
			fmt.Printf("❌ Example directory not found: %s\n", examplePath)
			continue
		}

		cmd := exec.Command("go", "run", "main.go")
		cmd.Dir = examplePath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("❌ Error running example %s: %v\n", example, err)
		} else {
			fmt.Printf("✅ Example %s completed successfully\n", example)
		}
	}

	fmt.Println("\n🎉 All examples execution completed!")
}
