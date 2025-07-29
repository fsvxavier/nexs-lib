#!/bin/bash

# Nexs Observability infraestructure Management Script

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
COMPOSE_FILE="$SCRIPT_DIR/docker-compose.yml"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# Help function
show_help() {
    cat << EOF
Nexs Observability infraestructure Management

USAGE:
    ./manage.sh COMMAND [OPTIONS]

COMMANDS:
    up              Start all services
    down            Stop all services
    restart         Restart all services
    logs [service]  Show logs (optionally for specific service)
    status          Show service status
    clean           Clean up volumes and networks
    reset           Reset everything (clean + restart)
    
SERVICES:
    tracer          Start only tracing services (jaeger, tempo, otel-collector)
    logger          Start only logging services (elasticsearch, kibana, logstash)
    metrics         Start only metrics services (prometheus, grafana)
    databases       Start only databases (postgres, mongodb, redis)
    
EXAMPLES:
    ./manage.sh up                    # Start all services
    ./manage.sh up tracer            # Start only tracing services
    ./manage.sh logs jaeger          # Show Jaeger logs
    ./manage.sh status               # Show all services status
    ./manage.sh clean                # Clean up everything

EOF
}

# Check if Docker and Docker Compose are available
check_dependencies() {
    if ! command -v docker &> /dev/null; then
        error "Docker is not installed or not in PATH"
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        error "Docker Compose is not installed or not in PATH"
    fi
}

# Start services
start_services() {
    local service_group=${1:-"all"}
    
    log "Starting Nexs Observability infraestructure..."
    
    case $service_group in
        "all")
            docker-compose -f "$COMPOSE_FILE" up -d
            ;;
        "tracer")
            docker-compose -f "$COMPOSE_FILE" up -d jaeger tempo otel-collector
            ;;
        "logger")
            docker-compose -f "$COMPOSE_FILE" up -d elasticsearch kibana logstash fluentd
            ;;
        "metrics")
            docker-compose -f "$COMPOSE_FILE" up -d prometheus grafana
            ;;
        "databases")
            docker-compose -f "$COMPOSE_FILE" up -d postgres mongodb redis rabbitmq
            ;;
        *)
            docker-compose -f "$COMPOSE_FILE" up -d "$service_group"
            ;;
    esac
    
    success "Services started successfully!"
    show_urls
}

# Stop services
stop_services() {
    log "Stopping Nexs Observability infraestructure..."
    docker-compose -f "$COMPOSE_FILE" down
    success "Services stopped successfully!"
}

# Restart services
restart_services() {
    log "Restarting Nexs Observability infraestructure..."
    docker-compose -f "$COMPOSE_FILE" restart
    success "Services restarted successfully!"
}

# Show logs
show_logs() {
    local service=${1:-""}
    
    if [[ -z "$service" ]]; then
        docker-compose -f "$COMPOSE_FILE" logs -f
    else
        docker-compose -f "$COMPOSE_FILE" logs -f "$service"
    fi
}

# Show service status
show_status() {
    log "Nexs Observability infraestructure Status:"
    docker-compose -f "$COMPOSE_FILE" ps
}

# Clean up
clean_up() {
    warning "This will remove all volumes and networks. Are you sure? (y/N)"
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        log "Cleaning up Nexs Observability infraestructure..."
        docker-compose -f "$COMPOSE_FILE" down -v --remove-orphans
        docker network prune -f
        docker volume prune -f
        success "Cleanup completed!"
    else
        log "Cleanup cancelled."
    fi
}

# Reset everything
reset_all() {
    log "Resetting Nexs Observability infraestructure..."
    clean_up
    start_services
}

# Show URLs
show_urls() {
    cat << EOF

${GREEN}Nexs Observability Services URLs:${NC}

${BLUE}🔍 Tracing:${NC}
  - Jaeger UI:           http://localhost:16686
  - Tempo:               http://localhost:3200
  - OTEL Collector:      http://localhost:13133/status (health)

${BLUE}📊 Metrics:${NC}
  - Grafana:             http://localhost:3000 (admin/nexs123)
  - Prometheus:          http://localhost:9090

${BLUE}📝 Logging:${NC}
  - Kibana:              http://localhost:5601
  - Elasticsearch:       http://localhost:9200

${BLUE}💾 Databases:${NC}
  - PostgreSQL:          localhost:5432 (nexs/nexs123/nexs_test)
  - MongoDB:             localhost:27017 (nexs/nexs123/nexs_test)
  - Redis:               localhost:6379 (password: nexs123)
  - RabbitMQ:            http://localhost:15672 (nexs/nexs123)

${BLUE}🚀 Development Endpoints:${NC}
  - OTLP HTTP:           http://localhost:4318/v1/traces
  - OTLP gRPC:           localhost:4317
  - Logstash TCP:        localhost:5000
  - Fluentd Forward:     localhost:24224

EOF
}

# Health check
health_check() {
    log "Performing health check..."
    
    services=(
        "jaeger:http://localhost:16686"
        "grafana:http://localhost:3000"
        "prometheus:http://localhost:9090"
        "kibana:http://localhost:5601"
        "elasticsearch:http://localhost:9200"
        "otel-collector:http://localhost:13133/status"
    )
    
    for service in "${services[@]}"; do
        name=$(echo "$service" | cut -d: -f1)
        url=$(echo "$service" | cut -d: -f2-)
        
        if curl -s -f "$url" > /dev/null 2>&1; then
            success "$name is healthy"
        else
            warning "$name is not responding"
        fi
    done
}

# Main script logic
main() {
    check_dependencies
    
    case ${1:-""} in
        "up")
            start_services "$2"
            ;;
        "down")
            stop_services
            ;;
        "restart")
            restart_services
            ;;
        "logs")
            show_logs "$2"
            ;;
        "status")
            show_status
            ;;
        "clean")
            clean_up
            ;;
        "reset")
            reset_all
            ;;
        "urls")
            show_urls
            ;;
        "health")
            health_check
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        "")
            show_help
            ;;
        *)
            error "Unknown command: $1. Use 'help' for available commands."
            ;;
    esac
}

# Run main function with all arguments
main "$@"
