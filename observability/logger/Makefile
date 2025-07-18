# Logger Module Makefile

.PHONY: test examples clean benchmark default basic advanced multi-provider

# Test all examples
test:
	@echo "Testing all logger examples..."
	@bash test_examples.sh

# Run default provider example
default:
	@echo "Running default provider example..."
	@cd examples/default-provider && go run main.go

# Run basic example
basic:
	@echo "Running basic example..."
	@cd examples/basic && go run main.go

# Run advanced example
advanced:
	@echo "Running advanced example..."
	@cd examples/advanced && go run main.go

# Run multi-provider example
multi-provider:
	@echo "Running multi-provider example..."
	@cd examples/multi-provider && go run main.go

# Run benchmark example
benchmark:
	@echo "Running benchmark example..."
	@cd examples/benchmark && go run main.go

# Run all examples
examples: default basic advanced multi-provider benchmark

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@go clean ./...

# Run tests
test-go:
	@echo "Running Go tests..."
	@go test ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Vet code
vet:
	@echo "Vetting code..."
	@go vet ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod tidy

# Full check (format, vet, test)
check: fmt vet test-go test

# Help
help:
	@echo "Available commands:"
	@echo "  test         - Test all examples"
	@echo "  default      - Run default provider example"
	@echo "  basic        - Run basic example"
	@echo "  advanced     - Run advanced example"
	@echo "  multi-provider - Run multi-provider example"
	@echo "  benchmark    - Run benchmark example"
	@echo "  examples     - Run all examples"
	@echo "  clean        - Clean build artifacts"
	@echo "  test-go      - Run Go tests"
	@echo "  fmt          - Format code"
	@echo "  vet          - Vet code"
	@echo "  deps         - Install dependencies"
	@echo "  check        - Full check (format, vet, test)"
	@echo "  help         - Show this help"
