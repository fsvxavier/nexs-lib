#!/bin/bash
# Stop Valkey services

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Change to infrastructure directory
cd "${SCRIPT_DIR}/.."

# Function to display usage
usage() {
    echo "Usage: $0 [--volumes] [--images]"
    echo ""
    echo "Options:"
    echo "  --volumes   Also remove volumes (data will be lost)"
    echo "  --images    Also remove downloaded images"
    echo ""
    echo "Examples:"
    echo "  $0                    # Stop services only"
    echo "  $0 --volumes          # Stop services and remove data"
    echo "  $0 --volumes --images # Stop services, remove data and images"
    exit 1
}

# Parse command line arguments
REMOVE_VOLUMES=false
REMOVE_IMAGES=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --volumes)
            REMOVE_VOLUMES=true
            shift
            ;;
        --images)
            REMOVE_IMAGES=true
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            usage
            ;;
    esac
done

echo -e "${YELLOW}Stopping Valkey services...${NC}"

# Stop services
if $REMOVE_VOLUMES; then
    echo -e "${RED}Warning: This will remove all data volumes!${NC}"
    docker-compose down -v
    echo -e "${GREEN}✓ Services stopped and volumes removed${NC}"
else
    docker-compose down
    echo -e "${GREEN}✓ Services stopped${NC}"
fi

# Remove images if requested
if $REMOVE_IMAGES; then
    echo -e "${YELLOW}Removing Valkey images...${NC}"
    docker images valkey/valkey --format "table {{.Repository}}:{{.Tag}}" | grep -v "REPOSITORY" | xargs -r docker rmi
    echo -e "${GREEN}✓ Images removed${NC}"
fi

# Show remaining containers (if any)
echo ""
echo "Remaining Valkey containers:"
if docker ps -a --filter "name=valkey" --format "table {{.Names}}\t{{.Status}}" | grep -q "valkey"; then
    docker ps -a --filter "name=valkey" --format "table {{.Names}}\t{{.Status}}"
else
    echo "None"
fi

# Show remaining volumes (if any)
echo ""
echo "Remaining Valkey volumes:"
if docker volume ls --filter "name=valkey" --format "table {{.Name}}" | grep -q "valkey"; then
    docker volume ls --filter "name=valkey" --format "table {{.Name}}"
else
    echo "None"
fi

echo ""
echo -e "${GREEN}Cleanup completed!${NC}"
