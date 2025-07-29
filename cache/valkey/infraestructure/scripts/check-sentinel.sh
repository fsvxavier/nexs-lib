#!/bin/bash
# Check Valkey Sentinel Status

set -e

echo "=== Checking Valkey Sentinel Status ==="

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

# Function to check sentinel
check_sentinel() {
    local container_name=$1
    local port=$2
    
    if check_container "${container_name}"; then
        local master_info=$(docker exec "${container_name}" valkey-cli -p "${port}" sentinel masters 2>/dev/null | grep -A1 "name" | grep "mymaster" || echo "")
        if [ -n "${master_info}" ]; then
            echo -e "  ${GREEN}✓${NC} Sentinel is monitoring master 'mymaster'"
        else
            echo -e "  ${RED}✗${NC} Sentinel is not monitoring master 'mymaster'"
        fi
        
        local master_addr=$(docker exec "${container_name}" valkey-cli -p "${port}" sentinel get-master-addr-by-name mymaster 2>/dev/null | tr '\n' ':' | sed 's/:$//')
        if [ -n "${master_addr}" ]; then
            echo -e "  ${YELLOW}ℹ${NC} Master address: ${master_addr}"
        fi
    fi
}

# Function to check valkey instance
check_valkey_instance() {
    local container_name=$1
    local port=$2
    local role_expected=$3
    
    if check_container "${container_name}"; then
        local role=$(docker exec "${container_name}" valkey-cli -p "${port}" -a sentinelpass123 info replication 2>/dev/null | grep "role:" | cut -d: -f2 | tr -d '\r')
        if [ "${role}" = "${role_expected}" ]; then
            echo -e "  ${GREEN}✓${NC} Role: ${role}"
        else
            echo -e "  ${RED}✗${NC} Role: ${role} (expected: ${role_expected})"
        fi
        
        if [ "${role}" = "master" ]; then
            local slaves=$(docker exec "${container_name}" valkey-cli -p "${port}" -a sentinelpass123 info replication 2>/dev/null | grep "connected_slaves:" | cut -d: -f2 | tr -d '\r')
            echo -e "  ${YELLOW}ℹ${NC} Connected slaves: ${slaves:-0}"
        elif [ "${role}" = "slave" ]; then
            local master_host=$(docker exec "${container_name}" valkey-cli -p "${port}" -a sentinelpass123 info replication 2>/dev/null | grep "master_host:" | cut -d: -f2 | tr -d '\r')
            local master_port=$(docker exec "${container_name}" valkey-cli -p "${port}" -a sentinelpass123 info replication 2>/dev/null | grep "master_port:" | cut -d: -f2 | tr -d '\r')
            echo -e "  ${YELLOW}ℹ${NC} Master: ${master_host}:${master_port}"
        fi
    fi
}

echo ""
echo "=== Valkey Instances Status ==="
check_valkey_instance "valkey-sentinel-master" "6379" "master"
check_valkey_instance "valkey-sentinel-slave-1" "6379" "slave"
check_valkey_instance "valkey-sentinel-slave-2" "6379" "slave"

echo ""
echo "=== Sentinel Instances Status ==="
check_sentinel "valkey-sentinel-1" "26379"
check_sentinel "valkey-sentinel-2" "26379"
check_sentinel "valkey-sentinel-3" "26379"

echo ""
echo "=== Sentinel Information ==="
if docker ps --format "table {{.Names}}" | grep -q "^valkey-sentinel-1$"; then
    echo "Sentinel masters:"
    docker exec valkey-sentinel-1 valkey-cli -p 26379 sentinel masters 2>/dev/null || echo "Failed to get sentinel masters"
    
    echo ""
    echo "Sentinel slaves:"
    docker exec valkey-sentinel-1 valkey-cli -p 26379 sentinel slaves mymaster 2>/dev/null || echo "Failed to get sentinel slaves"
    
    echo ""
    echo "Sentinel sentinels:"
    docker exec valkey-sentinel-1 valkey-cli -p 26379 sentinel sentinels mymaster 2>/dev/null || echo "Failed to get sentinel sentinels"
else
    echo -e "${RED}Cannot get sentinel information - sentinel 1 is not running${NC}"
fi

echo ""
echo "=== Connection Test ==="
if docker ps --format "table {{.Names}}" | grep -q "^valkey-sentinel-master$"; then
    echo "Testing master connection..."
    docker exec valkey-sentinel-master valkey-cli -p 6379 -a sentinelpass123 set test-key "sentinel-test-value" 2>/dev/null && \
    docker exec valkey-sentinel-master valkey-cli -p 6379 -a sentinelpass123 get test-key 2>/dev/null && \
    echo -e "${GREEN}✓${NC} Master connection test passed" || \
    echo -e "${RED}✗${NC} Master connection test failed"
    
    if docker ps --format "table {{.Names}}" | grep -q "^valkey-sentinel-slave-1$"; then
        echo "Testing slave read..."
        local value=$(docker exec valkey-sentinel-slave-1 valkey-cli -p 6379 -a sentinelpass123 get test-key 2>/dev/null)
        if [ "${value}" = "sentinel-test-value" ]; then
            echo -e "${GREEN}✓${NC} Slave read test passed"
        else
            echo -e "${RED}✗${NC} Slave read test failed (got: ${value})"
        fi
    fi
else
    echo -e "${RED}Cannot test sentinel connection - master is not running${NC}"
fi

echo ""
echo "=== Done ==="
