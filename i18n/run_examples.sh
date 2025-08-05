#!/bin/bash

# Set the working directory to the script location
cd "$(dirname "$0")"

echo "Running i18n module examples..."

# Run all examples
for example in examples/*/; do
    if [ -d "$example" ]; then
        echo "Running example: ${example%/}"
        cd "$example"
        go run main.go
        if [ $? -ne 0 ]; then
            echo "Warning: Example ${example%/} failed to run"
            # Continue with other examples instead of exiting
        fi
        cd - > /dev/null
    fi
done

echo "Running i18n module tests..."
# Run all tests with race detection
go test -v -race ./...
