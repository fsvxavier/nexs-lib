#!/bin/bash

# Script to run all domain error examples
# This script demonstrates all the examples in the domain errors library

set -e

echo "========================================"
echo "Domain Errors - Running All Examples"
echo "========================================"
echo ""

# Get the script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Array of examples to run
examples=(
    "basic"
    "advanced"
    "specific-errors"
    "http-integration"
    "error-recovery"
    "serialization"
)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_colored() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to run an example
run_example() {
    local example_name=$1
    local example_dir="$SCRIPT_DIR/$example_name"
    
    if [ ! -d "$example_dir" ]; then
        print_colored $RED "❌ Example directory not found: $example_dir"
        return 1
    fi
    
    if [ ! -f "$example_dir/main.go" ]; then
        print_colored $RED "❌ main.go not found in: $example_dir"
        return 1
    fi
    
    print_colored $BLUE "🚀 Running example: $example_name"
    print_colored $YELLOW "📂 Directory: $example_dir"
    echo ""
    
    # Change to example directory and run
    cd "$example_dir"
    
    # Special handling for HTTP integration example (it starts a server)
    if [ "$example_name" == "http-integration" ]; then
        print_colored $YELLOW "⚠️  HTTP Integration example starts a server."
        print_colored $YELLOW "   This example is skipped in batch mode."
        print_colored $YELLOW "   To run it manually: cd $example_dir && go run main.go"
        echo ""
        return 0
    fi
    
    # Run the example with timeout
    timeout 30s go run main.go 2>&1 || {
        local exit_code=$?
        if [ $exit_code -eq 124 ]; then
            print_colored $YELLOW "⚠️  Example timed out after 30 seconds"
        else
            print_colored $RED "❌ Example failed with exit code: $exit_code"
            return 1
        fi
    }
    
    print_colored $GREEN "✅ Example completed successfully"
    echo ""
    echo "----------------------------------------"
    echo ""
    
    return 0
}

# Function to check Go installation
check_go() {
    if ! command -v go &> /dev/null; then
        print_colored $RED "❌ Go is not installed or not in PATH"
        print_colored $YELLOW "Please install Go from https://golang.org/dl/"
        exit 1
    fi
    
    local go_version=$(go version)
    print_colored $GREEN "✅ Go is installed: $go_version"
    echo ""
}

# Function to check if we're in the right directory
check_directory() {
    if [ ! -f "../error_types.go" ]; then
        print_colored $RED "❌ This script must be run from the examples directory"
        print_colored $YELLOW "Expected location: nexs-lib/domainerrors/examples/"
        exit 1
    fi
    
    print_colored $GREEN "✅ Running from correct directory"
    echo ""
}

# Function to print summary
print_summary() {
    local total=$1
    local successful=$2
    local failed=$3
    
    echo ""
    echo "========================================"
    echo "Summary"
    echo "========================================"
    print_colored $BLUE "Total examples: $total"
    print_colored $GREEN "Successful: $successful"
    print_colored $RED "Failed: $failed"
    echo ""
    
    if [ $failed -eq 0 ]; then
        print_colored $GREEN "🎉 All examples completed successfully!"
    else
        print_colored $YELLOW "⚠️  Some examples failed. Check output above for details."
    fi
    echo ""
}

# Function to print usage
print_usage() {
    echo "Usage: $0 [OPTIONS] [EXAMPLE_NAME]"
    echo ""
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo "  -l, --list     List all available examples"
    echo "  -v, --verbose  Enable verbose output"
    echo ""
    echo "Examples:"
    echo "  $0                    # Run all examples"
    echo "  $0 basic             # Run only the basic example"
    echo "  $0 --list            # List all examples"
    echo ""
}

# Function to list examples
list_examples() {
    echo "Available examples:"
    echo ""
    for example in "${examples[@]}"; do
        if [ -d "$SCRIPT_DIR/$example" ]; then
            print_colored $GREEN "✅ $example"
        else
            print_colored $RED "❌ $example (directory not found)"
        fi
    done
    echo ""
}

# Parse command line arguments
VERBOSE=false
RUN_SPECIFIC=""

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            print_usage
            exit 0
            ;;
        -l|--list)
            list_examples
            exit 0
            ;;
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -*)
            print_colored $RED "Unknown option: $1"
            print_usage
            exit 1
            ;;
        *)
            RUN_SPECIFIC="$1"
            shift
            ;;
    esac
done

# Main execution
main() {
    print_colored $BLUE "Domain Errors Library - Example Runner"
    echo ""
    
    # Pre-flight checks
    check_go
    check_directory
    
    local total=0
    local successful=0
    local failed=0
    
    # Run specific example if requested
    if [ -n "$RUN_SPECIFIC" ]; then
        if [[ " ${examples[@]} " =~ " $RUN_SPECIFIC " ]]; then
            total=1
            if run_example "$RUN_SPECIFIC"; then
                successful=1
            else
                failed=1
            fi
        else
            print_colored $RED "❌ Unknown example: $RUN_SPECIFIC"
            echo ""
            list_examples
            exit 1
        fi
    else
        # Run all examples
        for example in "${examples[@]}"; do
            total=$((total + 1))
            if run_example "$example"; then
                successful=$((successful + 1))
            else
                failed=$((failed + 1))
            fi
        done
    fi
    
    print_summary $total $successful $failed
    
    # Exit with appropriate code
    if [ $failed -eq 0 ]; then
        exit 0
    else
        exit 1
    fi
}

# Run main function
main "$@"
