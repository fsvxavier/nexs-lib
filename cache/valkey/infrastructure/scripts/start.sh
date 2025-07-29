#!/bin/bash
# Start specific Valkey configuration

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to display usage
usage() {
    echo "Usage: $0 [standalone|cluster|sentinel|all]"
    echo ""
    echo "Options:"
    echo "  standalone  Start only standalone Valkey instance"
    echo "  cluster     Start only cluster configuration (6 nodes)"
    echo "  sentinel    Start only sentinel configuration (1 master + 2 slaves + 3 sentinels)"
    echo "  all         Start all configurations (default)"
    echo ""
    echo "Examples:"
    echo "  $0 standalone"
    echo "  $0 cluster"
    echo "  $0 sentinel"
    echo "  $0 all"
    exit 1
}

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Change to infrastructure directory
cd "${SCRIPT_DIR}/.."

# Parse command line arguments
CONFIG=${1:-all}

case $CONFIG in
    standalone)
        echo -e "${GREEN}Starting Valkey Standalone...${NC}"
        docker-compose up -d valkey-standalone
        echo ""
        echo "Waiting for standalone to be ready..."
        sleep 5
        if docker exec valkey-standalone valkey-cli -a testpass123 ping 2>/dev/null | grep -q "PONG"; then
            echo -e "${GREEN}✓ Standalone is ready!${NC}"
            echo "Connection: localhost:6379"
            echo "Password: testpass123"
        else
            echo -e "${RED}✗ Standalone failed to start${NC}"
            exit 1
        fi
        ;;
    
    cluster)
        echo -e "${GREEN}Starting Valkey Cluster...${NC}"
        docker-compose up -d valkey-cluster-node-1 valkey-cluster-node-2 valkey-cluster-node-3 valkey-cluster-node-4 valkey-cluster-node-5 valkey-cluster-node-6
        echo ""
        echo "Waiting for cluster nodes to be ready..."
        sleep 10
        
        echo "Creating cluster..."
        docker-compose up -d valkey-cluster-setup
        sleep 15
        
        if docker exec valkey-cluster-node-1 valkey-cli -p 7000 -a clusterpass123 cluster info 2>/dev/null | grep -q "cluster_state:ok"; then
            echo -e "${GREEN}✓ Cluster is ready!${NC}"
            echo "Connections: localhost:7000-7005"
            echo "Password: clusterpass123"
        else
            echo -e "${YELLOW}! Cluster may need more time to form${NC}"
            echo "Check status with: ./scripts/check-cluster.sh"
        fi
        ;;
    
    sentinel)
        echo -e "${GREEN}Starting Valkey Sentinel...${NC}"
        docker-compose up -d valkey-sentinel-master valkey-sentinel-slave-1 valkey-sentinel-slave-2 valkey-sentinel-1 valkey-sentinel-2 valkey-sentinel-3
        echo ""
        echo "Waiting for sentinel to be ready..."
        sleep 15
        
        if docker exec valkey-sentinel-1 valkey-cli -p 26379 sentinel masters 2>/dev/null | grep -q "mymaster"; then
            echo -e "${GREEN}✓ Sentinel is ready!${NC}"
            echo "Sentinel connections: localhost:26379-26381"
            echo "Master name: mymaster"
            echo "Password: sentinelpass123"
        else
            echo -e "${RED}✗ Sentinel failed to start${NC}"
            exit 1
        fi
        ;;
    
    all)
        echo -e "${GREEN}Starting All Valkey Configurations...${NC}"
        docker-compose up -d
        echo ""
        echo "Waiting for services to be ready..."
        sleep 20
        
        # Check standalone
        if docker exec valkey-standalone valkey-cli -a testpass123 ping 2>/dev/null | grep -q "PONG"; then
            echo -e "${GREEN}✓ Standalone is ready${NC}"
        else
            echo -e "${RED}✗ Standalone failed${NC}"
        fi
        
        # Check cluster
        if docker exec valkey-cluster-node-1 valkey-cli -p 7000 -a clusterpass123 cluster info 2>/dev/null | grep -q "cluster_state:ok"; then
            echo -e "${GREEN}✓ Cluster is ready${NC}"
        else
            echo -e "${YELLOW}! Cluster may need more time${NC}"
        fi
        
        # Check sentinel
        if docker exec valkey-sentinel-1 valkey-cli -p 26379 sentinel masters 2>/dev/null | grep -q "mymaster"; then
            echo -e "${GREEN}✓ Sentinel is ready${NC}"
        else
            echo -e "${YELLOW}! Sentinel may need more time${NC}"
        fi
        ;;
    
    *)
        echo -e "${RED}Invalid configuration: $CONFIG${NC}"
        usage
        ;;
esac

echo ""
echo -e "${GREEN}Configuration '$CONFIG' started successfully!${NC}"
echo ""
echo "Available commands:"
echo "  ./scripts/check-cluster.sh     - Check cluster status"
echo "  ./scripts/check-sentinel.sh    - Check sentinel status"
echo "  docker-compose logs -f         - View logs"
echo "  docker-compose down            - Stop all services"
