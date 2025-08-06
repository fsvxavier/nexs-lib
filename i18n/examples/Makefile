# Makefile for i18n examples
.PHONY: help basic advanced web microservice cli clean setup deps

# Default target
help: ## Show this help message
	@echo "🌍 I18n Examples Makefile"
	@echo "======================="
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'
	@echo ""
	@echo "Examples:"
	@echo "  make basic-json     # Run basic JSON example"
	@echo "  make web-gin        # Run Gin web application"
	@echo "  make microservice   # Run microservice example"
	@echo "  make all-examples   # Run all examples (non-interactive)"

# Setup and dependencies
setup: ## Setup all examples with dependencies
	@echo "🔧 Setting up examples..."
	@cd basic_json && go mod init basic_json 2>/dev/null || true
	@cd basic_yaml && go mod init basic_yaml 2>/dev/null || true
	@cd advanced && go mod init advanced 2>/dev/null || true
	@cd middleware_demo && go mod init middleware_demo 2>/dev/null || true
	@cd performance_demo && go mod init performance_demo 2>/dev/null || true
	@cd web_app_gin && go mod init web_app_gin 2>/dev/null || true
	@cd api_rest_echo && go mod init api_rest_echo 2>/dev/null || true
	@cd microservice && go mod init microservice_i18n 2>/dev/null || true
	@cd cli_tool && go mod init cli_tool 2>/dev/null || true
	@echo "✅ Setup completed"

deps: ## Install dependencies for all examples
	@echo "📦 Installing dependencies..."
	@cd web_app_gin && go get github.com/gin-gonic/gin
	@cd api_rest_echo && go get github.com/labstack/echo/v4
	@echo "✅ Dependencies installed"

# Basic examples
basic-json: ## Run basic JSON example
	@echo "🔸 Running basic JSON example..."
	@cd basic_json && go run main.go

basic-yaml: ## Run basic YAML example
	@echo "🔸 Running basic YAML example..."
	@cd basic_yaml && go run main.go

basic: basic-json basic-yaml ## Run all basic examples

# Advanced examples
advanced: ## Run advanced features example
	@echo "🔸 Running advanced example..."
	@cd advanced && go run main.go

middleware: ## Run middleware demonstration
	@echo "🔸 Running middleware demo..."
	@cd middleware_demo && go run main.go

performance: ## Run performance demonstration
	@echo "🔸 Running performance demo..."
	@cd performance_demo && go run main.go

advanced-all: advanced middleware performance ## Run all advanced examples

# Web applications
web-gin: deps ## Run Gin web application (requires manual stop)
	@echo "🌐 Starting Gin web application on http://localhost:8080"
	@echo "🛑 Press Ctrl+C to stop"
	@cd web_app_gin && go run main.go

web-echo: deps ## Run Echo API server (requires manual stop)
	@echo "🌐 Starting Echo API server on http://localhost:8080"
	@echo "🛑 Press Ctrl+C to stop"
	@cd api_rest_echo && go run main.go

# Microservice
microservice: ## Run i18n microservice (requires manual stop)
	@echo "🚀 Starting i18n microservice on http://localhost:8080"
	@echo "🔗 Health check: http://localhost:8080/health"
	@echo "🔗 API docs: http://localhost:8080/api/docs"
	@echo "🛑 Press Ctrl+C to stop"
	@cd microservice && go run main.go

# CLI tools
cli-interactive: ## Run CLI tool in interactive mode
	@echo "💻 Starting CLI tool in interactive mode..."
	@cd cli_tool && go run main.go -interactive

cli-translate: ## Run CLI tool translate command
	@echo "💻 Running CLI translate command..."
	@cd cli_tool && go run main.go -cmd translate -lang pt

cli-stats: ## Run CLI tool stats command
	@echo "💻 Running CLI stats command..."
	@cd cli_tool && go run main.go -cmd stats -verbose

cli: cli-stats ## Run CLI tool with stats

# Test all examples (non-interactive)
test-examples: ## Test all non-interactive examples
	@echo "🧪 Testing all examples..."
	@$(MAKE) basic-json
	@echo ""
	@$(MAKE) basic-yaml
	@echo ""
	@$(MAKE) advanced
	@echo ""
	@$(MAKE) middleware
	@echo ""
	@$(MAKE) performance
	@echo ""
	@$(MAKE) cli-stats
	@echo ""
	@echo "✅ All examples tested successfully!"

# Run specific groups
all-basic: basic ## Run all basic examples
all-advanced: advanced-all ## Run all advanced examples

# Cleanup
clean: ## Clean up temporary files and modules
	@echo "🧹 Cleaning up..."
	@find . -name "go.mod" -not -path "./basic_json/go.mod" -not -path "./basic_yaml/go.mod" -delete 2>/dev/null || true
	@find . -name "go.sum" -delete 2>/dev/null || true
	@echo "✅ Cleanup completed"

# Development helpers
validate-syntax: ## Validate Go syntax in all examples
	@echo "🔍 Validating Go syntax..."
	@for dir in */; do \
		if [ -f "$$dir/main.go" ]; then \
			echo "Checking $$dir"; \
			cd "$$dir" && go vet ./... && cd ..; \
		fi \
	done
	@echo "✅ Syntax validation completed"

format: ## Format Go code in all examples
	@echo "🎨 Formatting Go code..."
	@find . -name "*.go" -exec go fmt {} \;
	@echo "✅ Code formatting completed"

# Documentation
docs: ## Generate documentation for examples
	@echo "📖 Generating documentation..."
	@echo "See README.md for complete documentation"
	@echo "Available examples:"
	@ls -1 */ | grep -E '/$' | sed 's/\///' | while read dir; do \
		if [ -f "$$dir/main.go" ]; then \
			echo "  - $$dir"; \
		fi \
	done

# Quick demo for presentations
demo: ## Run a quick demo of key features
	@echo "🎬 I18n Library Demo"
	@echo "==================="
	@echo ""
	@echo "1. Basic JSON translation..."
	@sleep 1
	@cd basic_json && timeout 3s go run main.go || true
	@echo ""
	@echo "2. Performance benchmarks..."
	@sleep 1
	@cd performance_demo && timeout 5s go run main.go || true
	@echo ""
	@echo "3. Middleware demonstration..."
	@sleep 1
	@cd middleware_demo && timeout 4s go run main.go || true
	@echo ""
	@echo "✨ Demo completed! See README.md for more examples."

# CI/CD helpers
ci-test: validate-syntax test-examples ## Run CI tests

# Install global tools
install-tools: ## Install development tools
	@echo "🔧 Installing development tools..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ Tools installed"

# Comprehensive test
full-test: clean setup deps ci-test ## Full test cycle
	@echo "🎯 Full test cycle completed successfully!"

# Default target
.DEFAULT_GOAL := help
