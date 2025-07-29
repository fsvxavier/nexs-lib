# Makefile for Valkey Infrastructure

.PHONY: help start stop status logs clean test integration-test cluster sentinel standalone

# Default target
help:
	@echo "Valkey Infrastructure Commands:"
	@echo ""
	@echo "Start Commands:"
	@echo "  make start           - Start all Valkey configurations"
	@echo "  make standalone      - Start only standalone configuration"
	@echo "  make cluster         - Start only cluster configuration"
	@echo "  make sentinel        - Start only sentinel configuration"
	@echo ""
	@echo "Status Commands:"
	@echo "  make status          - Show status of all services"
	@echo "  make logs            - Show logs from all services"
	@echo "  make check-cluster   - Check cluster status"
	@echo "  make check-sentinel  - Check sentinel status"
	@echo ""
	@echo "Test Commands:"
	@echo "  make test            - Run unit tests"
	@echo "  make integration-test - Run integration tests with Docker"
	@echo ""
	@echo "Stop Commands:"
	@echo "  make stop            - Stop all services"
	@echo "  make clean           - Stop services and remove volumes"
	@echo "  make clean-all       - Stop services, remove volumes and images"

# Start commands
start:
	@./scripts/start.sh all

standalone:
	@./scripts/start.sh standalone

cluster:
	@./scripts/start.sh cluster

sentinel:
	@./scripts/start.sh sentinel

# Status commands
status:
	@docker-compose ps

logs:
	@docker-compose logs --tail=50 -f

check-cluster:
	@./scripts/check-cluster.sh

check-sentinel:
	@./scripts/check-sentinel.sh

# Test commands
test:
	@echo "Running unit tests..."
	@cd .. && go test -v ./...

integration-test:
	@./scripts/run-integration-tests.sh

# Stop commands
stop:
	@./scripts/stop.sh

clean:
	@./scripts/stop.sh --volumes

clean-all:
	@./scripts/stop.sh --volumes --images

# Utility commands
shell-standalone:
	@docker exec -it valkey-standalone valkey-cli -a testpass123

shell-cluster:
	@docker exec -it valkey-cluster-node-1 valkey-cli -c -p 7000 -a clusterpass123

shell-sentinel:
	@docker exec -it valkey-sentinel-master valkey-cli -a sentinelpass123

shell-sentinel-cli:
	@docker exec -it valkey-sentinel-1 valkey-cli -p 26379

# Docker compose shortcuts
up:
	@docker-compose up -d

down:
	@docker-compose down

restart:
	@docker-compose restart

build:
	@docker-compose build

pull:
	@docker-compose pull
