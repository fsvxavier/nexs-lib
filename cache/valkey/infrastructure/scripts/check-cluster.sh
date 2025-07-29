#!/bin/bash
# Check Valkey Cluster Status

set -e

echo "=== Checking Valkey Cluster Status ==="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to check if container is running
check_container() {
    local container_name=$1
    if docker ps --format "table {{.Names}}" | grep -q "^${container_name}$"; then
        echo -e "${GREEN}✓${NC} ${container_name} is running"
        return 0
    else
        echo -e "${RED}✗${NC} ${container_name} is not running"
        return 1
    fi
}

# Function to check cluster node
check_cluster_node() {
    local container_name=$1
    local port=$2
    
    if check_container "${container_name}"; then
        local info=$(docker exec "${container_name}" valkey-cli -p "${port}" -a clusterpass123 cluster info 2>/dev/null | grep cluster_state | cut -d: -f2)
        if [ "${info}" = "ok" ]; then
            echo -e "  ${GREEN}✓${NC} Cluster state: OK"
        else
            echo -e "  ${RED}✗${NC} Cluster state: ${info:-UNKNOWN}"
        fi
        
        local nodes=$(docker exec "${container_name}" valkey-cli -p "${port}" -a clusterpass123 cluster nodes 2>/dev/null | wc -l)
        echo -e "  ${YELLOW}ℹ${NC} Visible nodes: ${nodes}"
    fi
}

echo ""
echo "=== Cluster Nodes Status ==="
check_cluster_node "valkey-cluster-node-1" "7000"
check_cluster_node "valkey-cluster-node-2" "7001"
check_cluster_node "valkey-cluster-node-3" "7002"
check_cluster_node "valkey-cluster-node-4" "7003"
check_cluster_node "valkey-cluster-node-5" "7004"
check_cluster_node "valkey-cluster-node-6" "7005"

echo ""
echo "=== Cluster Information ==="
if docker ps --format "table {{.Names}}" | grep -q "^valkey-cluster-node-1$"; then
    echo "Cluster nodes:"
    docker exec valkey-cluster-node-1 valkey-cli -p 7000 -a clusterpass123 cluster nodes 2>/dev/null || echo "Failed to get cluster nodes"
    
    echo ""
    echo "Cluster slots:"
    docker exec valkey-cluster-node-1 valkey-cli -p 7000 -a clusterpass123 cluster slots 2>/dev/null || echo "Failed to get cluster slots"
else
    echo -e "${RED}Cannot get cluster information - node 1 is not running${NC}"
fi

echo ""
echo "=== Connection Test ==="
if docker ps --format "table {{.Names}}" | grep -q "^valkey-cluster-node-1$"; then
    echo "Testing cluster connection..."
    docker exec valkey-cluster-node-1 valkey-cli -c -p 7000 -a clusterpass123 set test-key "cluster-test-value" 2>/dev/null && \
    docker exec valkey-cluster-node-1 valkey-cli -c -p 7000 -a clusterpass123 get test-key 2>/dev/null && \
    echo -e "${GREEN}✓${NC} Cluster connection test passed" || \
    echo -e "${RED}✗${NC} Cluster connection test failed"
else
    echo -e "${RED}Cannot test cluster connection - nodes are not running${NC}"
fi

echo ""
echo "=== Done ==="
