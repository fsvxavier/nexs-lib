# Nexs Observability Infrastructure Makefile

.PHONY: help infra-up infra-down infra-logs infra-status infra-clean infra-reset test-integration

# Default target
help: ## Show this help message
	@echo "Nexs Observability Infrastructure Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Service Groups:"
	@echo "  make infra-up GROUP=tracer     # Start only tracing services"
	@echo "  make infra-up GROUP=logger     # Start only logging services"  
	@echo "  make infra-up GROUP=metrics    # Start only metrics services"
	@echo "  make infra-up GROUP=databases  # Start only databases"

# Infrastructure management
infra-up: ## Start observability infrastructure
	@echo "🚀 Starting Nexs Observability Infrastructure..."
	@cd infrastructure && ./manage.sh up $(GROUP)

infra-down: ## Stop observability infrastructure
	@echo "🛑 Stopping Nexs Observability Infrastructure..."
	@cd infrastructure && ./manage.sh down

infra-restart: ## Restart observability infrastructure
	@echo "🔄 Restarting Nexs Observability Infrastructure..."
	@cd infrastructure && ./manage.sh restart

infra-logs: ## Show infrastructure logs (SERVICE=service_name optional)
	@cd infrastructure && ./manage.sh logs $(SERVICE)

infra-status: ## Show infrastructure status
	@cd infrastructure && ./manage.sh status

infra-clean: ## Clean up infrastructure (removes volumes)
	@echo "🧹 Cleaning up Nexs Observability Infrastructure..."
	@cd infrastructure && ./manage.sh clean

infra-reset: ## Reset infrastructure (clean + restart)
	@echo "🔄 Resetting Nexs Observability Infrastructure..."
	@cd infrastructure && ./manage.sh reset

infra-urls: ## Show service URLs
	@cd infrastructure && ./manage.sh urls

infra-health: ## Check infrastructure health
	@cd infrastructure && ./manage.sh health

# Development workflows
dev-setup: infra-up ## Setup complete development environment
	@echo "✅ Development environment ready!"
	@echo ""
	@make infra-urls

dev-tracer: ## Start only tracer development environment
	@make infra-up GROUP=tracer
	@make infra-up GROUP=databases

dev-logger: ## Start only logger development environment  
	@make infra-up GROUP=logger
	@make infra-up GROUP=databases

dev-metrics: ## Start only metrics development environment
	@make infra-up GROUP=metrics

# Testing with infrastructure
test-integration: infra-up ## Run integration tests with infrastructure
	@echo "🧪 Running integration tests..."
	@sleep 10  # Wait for services to be ready
	@cd tracer && go test -tags=integration -v -timeout=60s ./...
	@cd logger && go test -tags=integration -v -timeout=60s ./...

test-integration-tracer: ## Run tracer integration tests
	@make infra-up GROUP=tracer
	@make infra-up GROUP=databases  
	@echo "🧪 Running tracer integration tests..."
	@sleep 10
	@cd tracer && go test -tags=integration -v -timeout=60s ./...

test-integration-logger: ## Run logger integration tests
	@make infra-up GROUP=logger
	@make infra-up GROUP=databases
	@echo "🧪 Running logger integration tests..."
	@sleep 10
	@cd logger && go test -tags=integration -v -timeout=60s ./...

# Example testing
test-examples: infra-up ## Test all examples against infrastructure
	@echo "🧪 Testing examples against infrastructure..."
	@sleep 15  # Wait for all services
	@$(MAKE) test-tracer-examples
	@$(MAKE) test-logger-examples

test-tracer-examples: ## Test tracer examples
	@echo "🔍 Testing tracer examples..."
	@cd tracer/examples && for dir in */; do \
		echo "Testing $$dir..."; \
		cd $$dir && timeout 30s go run main.go & \
		PID=$$!; sleep 5; \
		curl -f http://localhost:8080/health || echo "❌ $$dir failed"; \
		kill $$PID 2>/dev/null || true; \
		cd ..; \
	done

test-logger-examples: ## Test logger examples (if any)
	@echo "📝 Testing logger examples..."
	@echo "Logger examples testing not yet implemented"

# Monitoring
monitor-traces: ## Monitor traces in Jaeger UI
	@echo "🔍 Opening Jaeger UI for trace monitoring..."
	@open http://localhost:16686 || xdg-open http://localhost:16686 || echo "Open http://localhost:16686 manually"

monitor-metrics: ## Monitor metrics in Grafana
	@echo "📊 Opening Grafana for metrics monitoring..."
	@open http://localhost:3000 || xdg-open http://localhost:3000 || echo "Open http://localhost:3000 manually (admin/nexs123)"

monitor-logs: ## Monitor logs in Kibana
	@echo "📝 Opening Kibana for log monitoring..."
	@open http://localhost:5601 || xdg-open http://localhost:5601 || echo "Open http://localhost:5601 manually"

# Cleanup and maintenance
clean-all: infra-clean ## Clean everything including Go mod cache
	@echo "🧹 Cleaning Go modules..."
	@go clean -modcache
	@go clean -cache

# Quick development commands
quick-start: ## Quick start for development (tracer + databases)
	@make infra-up GROUP=tracer
	@make infra-up GROUP=databases
	@echo "⚡ Quick development environment ready!"

quick-full: ## Quick start full stack
	@make infra-up
	@echo "🚀 Full observability stack ready!"
	@make infra-urls

# Docker specific commands
docker-prune: ## Prune Docker system (careful!)
	@echo "🧹 Pruning Docker system..."
	@docker system prune -f
	@docker volume prune -f

docker-logs: ## Show all Docker logs
	@docker-compose -f infrastructure/docker-compose.yml logs

# Environment validation
validate-env: ## Validate environment setup
	@echo "🔍 Validating environment..."
	@command -v docker >/dev/null 2>&1 || { echo "❌ Docker not found"; exit 1; }
	@command -v docker-compose >/dev/null 2>&1 || { echo "❌ Docker Compose not found"; exit 1; }
	@command -v go >/dev/null 2>&1 || { echo "❌ Go not found"; exit 1; }
	@command -v curl >/dev/null 2>&1 || { echo "❌ curl not found"; exit 1; }
	@echo "✅ Environment validation passed!"

# Performance testing
perf-test: infra-up ## Run performance tests against infrastructure
	@echo "⚡ Running performance tests..."
	@sleep 15
	@cd tracer && go test -bench=. -benchmem ./...

# Default target when no target is specified
.DEFAULT_GOAL := help
