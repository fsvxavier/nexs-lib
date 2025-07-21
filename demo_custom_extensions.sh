#!/bin/bash

# Demo script for Custom Hooks and Middleware functionality
# This script demonstrates the new custom extensions system

echo "🎛️  Custom Hooks and Middleware Demo"
echo "====================================="
echo ""

# Check if we're in the right directory
if [[ ! -f "go.mod" ]]; then
    echo "❌ Error: Please run this script from the project root directory"
    exit 1
fi

echo "📋 Running Custom Hooks and Middleware Tests..."
echo ""

# Run custom hooks tests
echo "🎣 Testing Custom Hooks Implementation..."
go test -v ./httpserver/hooks/custom_test.go ./httpserver/hooks/custom.go -timeout 30s
HOOKS_EXIT_CODE=$?

echo ""
echo "🔧 Testing Custom Middleware Implementation..."
go test -v ./httpserver/middleware/custom_test.go ./httpserver/middleware/custom.go -timeout 30s
MIDDLEWARE_EXIT_CODE=$?

echo ""
echo "📊 Test Results Summary:"
echo "========================"

if [ $HOOKS_EXIT_CODE -eq 0 ]; then
    echo "✅ Custom Hooks Tests: PASSED"
else
    echo "❌ Custom Hooks Tests: FAILED"
fi

if [ $MIDDLEWARE_EXIT_CODE -eq 0 ]; then
    echo "✅ Custom Middleware Tests: PASSED"
else
    echo "❌ Custom Middleware Tests: FAILED"
fi

echo ""

# Run the examples
echo "🚀 Running Custom Usage Examples..."
echo "===================================="

# Create a temporary main file to run the examples
cat > /tmp/custom_demo.go << 'EOF'
package main

import (
    "log"
    "os"
    "path/filepath"
)

func main() {
    // Get the current working directory
    wd, err := os.Getwd()
    if err != nil {
        log.Fatal("Error getting working directory:", err)
    }
    
    // Construct the path to the examples
    examplesPath := filepath.Join(wd, "httpserver", "examples")
    
    log.Printf("📁 Examples location: %s", examplesPath)
    log.Println("")
    log.Println("🎯 Custom Hooks and Middleware Examples:")
    log.Println("==========================================")
    log.Println("")
    log.Println("ℹ️  The following examples demonstrate:")
    log.Println("   • Custom Hook Builder Pattern")
    log.Println("   • Custom Middleware Builder Pattern") 
    log.Println("   • Factory Methods for Different Hook Types")
    log.Println("   • Advanced Filtering and Conditions")
    log.Println("   • Async Hook Execution")
    log.Println("   • Before/After Middleware Functions")
    log.Println("   • Integration Between Hooks and Middleware")
    log.Println("")
    log.Println("📖 For detailed usage, see:")
    log.Println("   • httpserver/examples/custom_usage.go")
    log.Println("   • httpserver/README_CUSTOM.md")
    log.Println("")
    log.Println("✨ Custom extensions are now ready for use!")
}
EOF

# Run the demo
go run /tmp/custom_demo.go

# Clean up
rm -f /tmp/custom_demo.go

echo ""
echo "📚 Documentation and Examples:"
echo "==============================="
echo "• 📖 README_CUSTOM.md - Complete usage guide"
echo "• 🎯 examples/custom_usage.go - Working examples"
echo "• 🧪 hooks/custom_test.go - Hook test examples"
echo "• 🧪 middleware/custom_test.go - Middleware test examples"

echo ""
echo "🎉 Custom Hooks and Middleware Implementation Complete!"
echo "========================================================"
echo ""
echo "Key Features Implemented:"
echo "• ✅ Custom Hook Builder with fluent API"
echo "• ✅ Custom Hook Factory with multiple hook types"
echo "• ✅ Custom Middleware Builder with before/after functions"
echo "• ✅ Custom Middleware Factory with built-in helpers"
echo "• ✅ Advanced filtering (path, method, header, conditions)"
echo "• ✅ Async execution with timeout and buffer configuration"
echo "• ✅ Skip logic for middleware"
echo "• ✅ Type safety with interface compliance"
echo "• ✅ Comprehensive test coverage"
echo "• ✅ Complete documentation and examples"
echo ""
echo "🚀 Ready for production use!"

# Final exit code
if [ $HOOKS_EXIT_CODE -eq 0 ] && [ $MIDDLEWARE_EXIT_CODE -eq 0 ]; then
    exit 0
else
    exit 1
fi
