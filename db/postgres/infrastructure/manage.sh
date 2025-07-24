#!/bin/bash

# NEXS-LIB Infrastructure Management Script
# This script helps manage the PostgreSQL infrastructure for testing and examples

set -e

# Configuration
COMPOSE_FILE="infrastructure/docker/docker-compose.yml"
PROJECT_NAME="nexs-lib"

# Check if script is run from correct directory
check_working_directory() {
    if [ ! -f "$COMPOSE_FILE" ]; then
        error "This script must be run from the project root directory."
        echo "Current directory: $(pwd)"
        echo "Expected file: $COMPOSE_FILE"
        echo ""
        echo "Please run this script from the root of the nexs-lib project:"
        echo "  cd /path/to/nexs-lib"
        echo "  ./infrastructure/manage.sh [command]"
        exit 1
    fi
    
    if [ ! -d "db/postgres/examples" ]; then
        error "Examples directory not found. Please ensure you are in the correct project directory."
        exit 1
    fi
}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] $1${NC}"
}

success() {
    echo -e "${GREEN}[SUCCESS] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[WARNING] $1${NC}"
}

error() {
    echo -e "${RED}[ERROR] $1${NC}"
}

# Check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        error "Docker is not running or accessible."
        echo ""
        echo "Common solutions:"
        echo "1. Start Docker service:"
        echo "   - On Linux: sudo systemctl start docker"
        echo "   - On macOS/Windows: Start Docker Desktop"
        echo ""
        echo "2. Add your user to docker group (Linux):"
        echo "   - sudo usermod -aG docker \$USER"
        echo "   - Log out and log back in"
        echo ""
        echo "3. Check if Docker daemon is running:"
        echo "   - sudo systemctl status docker"
        echo ""
        echo "4. Try with sudo (not recommended for production):"
        echo "   - sudo ./infrastructure/manage.sh [command]"
        echo ""
        exit 1
    fi
}

# Check if docker-compose is available
check_docker_compose() {
    if ! command -v docker-compose > /dev/null 2>&1; then
        error "docker-compose is not installed."
        echo ""
        echo "To install docker-compose:"
        echo "  - On Ubuntu/Debian: sudo apt-get install docker-compose"
        echo "  - On CentOS/RHEL: sudo yum install docker-compose"
        echo "  - On macOS: brew install docker-compose"
        echo "  - Or install via pip: pip install docker-compose"
        echo ""
        exit 1
    fi
    
    # Test docker-compose access
    if ! docker-compose version > /dev/null 2>&1; then
        error "docker-compose is installed but not accessible."
        echo ""
        echo "This might be a permission issue. Try:"
        echo "  - Add your user to docker group: sudo usermod -aG docker \$USER"
        echo "  - Log out and log back in"
        echo "  - Or run with sudo: sudo ./infrastructure/manage.sh [command]"
        echo ""
        exit 1
    fi
}

# Check if compose file exists
check_compose_file() {
    if [ ! -f "$COMPOSE_FILE" ]; then
        error "Docker compose file not found: $COMPOSE_FILE"
        echo "Please ensure you are running this script from the project root directory."
        exit 1
    fi
}

# Cleanup existing containers and networks
cleanup_infrastructure() {
    log "Cleaning up existing infrastructure..."
    
    # Stop and remove containers
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" down --remove-orphans 2>/dev/null || true
    
    # Remove network if it exists
    docker network rm nexs-network 2>/dev/null || true
    
    # Clean up any dangling containers
    docker container prune -f 2>/dev/null || true
}

# Check replica status
check_replica_status() {
    local replica_name=$1
    local port=$2
    
    log "Checking $replica_name status..."
    
    # Check if container is running
    if docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" ps $replica_name 2>/dev/null | grep -q "Up"; then
        log "$replica_name container is running"
        
        # Check if PostgreSQL is ready
        if docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" exec -T $replica_name pg_isready -U nexs_user -d nexs_testdb > /dev/null 2>&1; then
            success "$replica_name is ready and responding!"
            return 0
        else
            warn "$replica_name container is running but PostgreSQL is not ready"
            return 1
        fi
    else
        warn "$replica_name container is not running"
        return 1
    fi
}

# Show startup logs for debugging
show_startup_logs() {
    echo ""
    error "=== STARTUP LOGS FOR DEBUGGING ==="
    echo ""
    
    log "Primary database logs:"
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" logs --tail=20 postgres-primary || true
    
    echo ""
    log "Replica 1 logs:"
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" logs --tail=20 postgres-replica1 || true
    
    echo ""
    log "Container status:"
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" ps || true
    
    echo ""
}

# Start infrastructure
start_infrastructure() {
    log "Starting NEXS-LIB infrastructure..."
    
    check_docker
    check_docker_compose
    check_compose_file
    
    # Cleanup any existing infrastructure
    cleanup_infrastructure
    
    # Create network
    log "Creating Docker network..."
    docker network create nexs-network 2>/dev/null || true
    
    # Start services
    log "Starting Docker services..."
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" up -d
    
    # Check if services started successfully
    if [ $? -ne 0 ]; then
        error "Failed to start services. Check Docker logs for details."
        exit 1
    fi
    
    # Wait for services to be ready
    log "Waiting for services to be ready..."
    
    # Wait for primary database with timeout
    log "Waiting for primary database..."
    local timeout=60
    local counter=0
    
    while [ $counter -lt $timeout ]; do
        if docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" exec -T postgres-primary pg_isready -U nexs_user -d nexs_testdb > /dev/null 2>&1; then
            success "Primary database is ready!"
            break
        fi
        
        counter=$((counter + 1))
        sleep 2
        
        if [ $counter -eq $timeout ]; then
            error "Primary database failed to start within ${timeout} seconds"
            show_startup_logs
            exit 1
        fi
    done
    
    # Wait for replicas with timeout
    log "Waiting for replica databases..."
    sleep 20
    
    # Check replica 1
    log "Checking replica 1 status..."
    local replica_timeout=60
    local replica_counter=0
    
    while [ $replica_counter -lt $replica_timeout ]; do
        if check_replica_status "postgres-replica1" "5433"; then
            break
        fi
        
        replica_counter=$((replica_counter + 1))
        sleep 3
        
        if [ $replica_counter -eq $replica_timeout ]; then
            warn "Replica 1 is not ready after ${replica_timeout} attempts, but continuing..."
            log "Showing replica 1 logs for debugging:"
            docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" logs --tail=15 postgres-replica1 || true
            break
        fi
    done
    
    # Check replica 2
    log "Checking replica 2 status..."
    replica_counter=0
    
    while [ $replica_counter -lt $replica_timeout ]; do
        if check_replica_status "postgres-replica2" "5434"; then
            break
        fi
        
        replica_counter=$((replica_counter + 1))
        sleep 3
        
        if [ $replica_counter -eq $replica_timeout ]; then
            warn "Replica 2 is not ready after ${replica_timeout} attempts, but continuing..."
            log "Showing replica 2 logs for debugging:"
            docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" logs --tail=15 postgres-replica2 || true
            break
        fi
    done
    
    success "Infrastructure started successfully!"
    
    # Show connection information
    echo ""
    echo "Connection Information:"
    echo "======================="
    echo "Primary Database:"
    echo "  Host: localhost"
    echo "  Port: 5432"
    echo "  Database: nexs_testdb"
    echo "  User: nexs_user"
    echo "  Password: nexs_password"
    echo "  Connection String: postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb"
    echo ""
    echo "Replica 1:"
    echo "  Host: localhost"
    echo "  Port: 5433"
    echo "  Connection String: postgres://nexs_user:nexs_password@localhost:5433/nexs_testdb"
    echo ""
    echo "Replica 2:"
    echo "  Host: localhost"
    echo "  Port: 5434"
    echo "  Connection String: postgres://nexs_user:nexs_password@localhost:5434/nexs_testdb"
    echo ""
    echo "PgAdmin: http://localhost:8080"
    echo "  Email: admin@nexs.com"
    echo "  Password: admin123"
    echo ""
    echo "Redis: localhost:6379"
}

# Stop infrastructure
stop_infrastructure() {
    log "Stopping NEXS-LIB infrastructure..."
    
    check_docker_compose
    check_compose_file
    
    # Stop and remove containers
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" down --remove-orphans
    
    # Remove network
    docker network rm nexs-network 2>/dev/null || true
    
    success "Infrastructure stopped successfully!"
}

# Restart infrastructure
restart_infrastructure() {
    log "Restarting NEXS-LIB infrastructure..."
    
    stop_infrastructure
    start_infrastructure
}

# Show logs
show_logs() {
    local service=${1:-""}
    
    check_docker_compose
    
    if [ -n "$service" ]; then
        docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" logs -f "$service"
    else
        docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" logs -f
    fi
}

# Show status
show_status() {
    log "NEXS-LIB Infrastructure Status:"
    echo ""
    
    check_docker_compose
    check_compose_file
    
    # Check if infrastructure is running
    if docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" ps 2>/dev/null | grep -q "Up"; then
        docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" ps
        echo ""
        success "Infrastructure is running"
    else
        warn "Infrastructure is not running"
        echo ""
        echo "To start the infrastructure:"
        echo "  ./infrastructure/manage.sh start"
        echo ""
        echo "To see what containers exist:"
        docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" ps 2>/dev/null || echo "No containers found"
    fi
}

# Reset database
reset_database() {
    log "Resetting database..."
    
    check_docker_compose
    check_compose_file
    
    # Stop services
    log "Stopping services..."
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" stop
    
    # Remove volumes
    log "Removing volumes..."
    docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" down -v --remove-orphans
    
    # Remove network
    docker network rm nexs-network 2>/dev/null || true
    
    # Start services again
    log "Starting services again..."
    start_infrastructure
    
    success "Database reset successfully!"
}

# Run tests
run_tests() {
    log "Running tests with Docker infrastructure..."
    
    check_docker_compose
    check_compose_file
    
    # Start infrastructure if not running
    if ! docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" ps | grep -q "Up"; then
        log "Infrastructure not running, starting it..."
        start_infrastructure
    else
        log "Infrastructure already running, verifying health..."
        
        # Verify primary database is healthy
        if ! docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" exec -T postgres-primary pg_isready -U nexs_user -d nexs_testdb > /dev/null 2>&1; then
            warn "Primary database not healthy, restarting infrastructure..."
            start_infrastructure
        fi
    fi
    
    # Set environment variables for tests
    export NEXS_DB_HOST=localhost
    export NEXS_DB_PORT=5432
    export NEXS_DB_NAME=nexs_testdb
    export NEXS_DB_USER=nexs_user
    export NEXS_DB_PASSWORD=nexs_password
    export NEXS_DB_REPLICA1_HOST=localhost
    export NEXS_DB_REPLICA1_PORT=5433
    export NEXS_DB_REPLICA2_HOST=localhost
    export NEXS_DB_REPLICA2_PORT=5434
    export NEXS_DB_DSN="postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb"
    export NEXS_DB_REPLICA_DSN="postgres://nexs_user:nexs_password@localhost:5433/nexs_testdb"
    
    # Run tests
    log "Running Go tests..."
    cd db/postgres
    if go test -v -race -timeout 30s ./...; then
        success "Tests completed successfully!"
    else
        error "Tests failed! Check the output above for details."
        exit 1
    fi
}

# Run specific example
run_example() {
    local example=${1:-"basic"}
    
    log "Running $example example with Docker infrastructure..."
    
    check_docker_compose
    check_compose_file
    
    # Start infrastructure if not running
    if ! docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" ps | grep -q "Up"; then
        log "Infrastructure not running, starting it..."
        start_infrastructure
    else
        log "Infrastructure already running, verifying health..."
        
        # Verify primary database is healthy
        if ! docker-compose -f "$COMPOSE_FILE" -p "$PROJECT_NAME" exec -T postgres-primary pg_isready -U nexs_user -d nexs_testdb > /dev/null 2>&1; then
            warn "Primary database not healthy, restarting infrastructure..."
            start_infrastructure
        fi
    fi
    
    # Set environment variables for examples
    export NEXS_DB_HOST=localhost
    export NEXS_DB_PORT=5432
    export NEXS_DB_NAME=nexs_testdb
    export NEXS_DB_USER=nexs_user
    export NEXS_DB_PASSWORD=nexs_password
    export NEXS_DB_REPLICA1_HOST=localhost
    export NEXS_DB_REPLICA1_PORT=5433
    export NEXS_DB_REPLICA2_HOST=localhost
    export NEXS_DB_REPLICA2_PORT=5434
    
    # Set DSN for different examples
    export NEXS_DB_DSN="postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb"
    export NEXS_DB_REPLICA_DSN="postgres://nexs_user:nexs_password@localhost:5433/nexs_testdb"
    
    # Validate example type
    case $example in
        basic|replicas|advanced|pool)
            ;;
        *)
            error "Invalid example type: $example"
            echo "Available examples: basic, replicas, advanced, pool"
            exit 1
            ;;
    esac
    
    # Run example
    if [ -d "db/postgres/examples/$example" ]; then
        log "Running example: $example"
        cd "db/postgres/examples/$example"
        if go run main.go; then
            success "Example '$example' completed successfully!"
        else
            error "Example '$example' failed! Check the output above for details."
            exit 1
        fi
    else
        error "Example '$example' not found!"
        echo "Available examples directory: db/postgres/examples/"
        ls -la db/postgres/examples/ 2>/dev/null || echo "Examples directory not found!"
        exit 1
    fi
}

# Show help
show_help() {
    echo "NEXS-LIB Infrastructure Management Script"
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  start              Start the infrastructure"
    echo "  stop               Stop the infrastructure"
    echo "  restart            Restart the infrastructure"
    echo "  status             Show infrastructure status"
    echo "  logs [service]     Show logs (optionally for specific service)"
    echo "  reset              Reset database (WARNING: This will delete all data)"
    echo "  test               Run tests with Docker infrastructure"
    echo "  example [name]     Run specific example (basic, replicas, advanced, pool)"
    echo "  help               Show this help"
    echo ""
    echo "Examples:"
    echo "  $0 start                    # Start infrastructure"
    echo "  $0 logs postgres-primary    # Show primary database logs"
    echo "  $0 test                     # Run tests"
    echo "  $0 example replicas         # Run replica example"
    echo "  $0 example advanced         # Run advanced example"
    echo "  $0 example pool             # Run pool example"
    echo ""
    echo "Services:"
    echo "  postgres-primary   Primary PostgreSQL database (port 5432)"
    echo "  postgres-replica1  PostgreSQL replica 1 (port 5433)"
    echo "  postgres-replica2  PostgreSQL replica 2 (port 5434)"
    echo "  redis              Redis cache (port 6379)"
    echo "  pgadmin            PgAdmin web interface (port 8080)"
    echo ""
    echo "Prerequisites:"
    echo "  - Docker must be running"
    echo "  - docker-compose must be installed"
    echo "  - Go 1.19+ must be installed"
    echo ""
    echo "Connection Information:"
    echo "  Primary: postgres://nexs_user:nexs_password@localhost:5432/nexs_testdb"
    echo "  Replica 1: postgres://nexs_user:nexs_password@localhost:5433/nexs_testdb"
    echo "  Replica 2: postgres://nexs_user:nexs_password@localhost:5434/nexs_testdb"
    echo "  PgAdmin: http://localhost:8080 (admin@nexs.com / admin123)"
    echo ""
    echo "Troubleshooting:"
    echo "  - If services fail to start, check: $0 logs"
    echo "  - If database connection fails, try: $0 reset"
    echo "  - If Docker is not running, start it first"
}

# Main script
main() {
    # Check if running from correct directory
    check_working_directory
    
    case "${1:-help}" in
        start)
            start_infrastructure
            ;;
        stop)
            stop_infrastructure
            ;;
        restart)
            restart_infrastructure
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs "$2"
            ;;
        reset)
            reset_database
            ;;
        test)
            run_tests
            ;;
        example)
            run_example "$2"
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            error "Unknown command: $1"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
