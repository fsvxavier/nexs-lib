#!/bin/bash

# Hooks and Middleware Examples Test Script

echo "=== Testing Hooks and Middleware Examples ==="
echo

# Test main hooks and middleware example
echo "Running main hooks and middleware example..."
cd /mnt/e/go/src/github.com/fsvxavier/nexs-lib/domainerrors/examples/hooks-middleware
go run main.go advanced.go > /tmp/hooks_middleware_output.txt 2>&1

if [ $? -eq 0 ]; then
    echo "✅ Main example executed successfully"
else
    echo "❌ Main example failed"
    cat /tmp/hooks_middleware_output.txt
fi

# Test enrichment pattern
echo
echo "Running enrichment pattern example..."
cd patterns
go run enrichment_pattern.go > /tmp/enrichment_output.txt 2>&1

if [ $? -eq 0 ]; then
    echo "✅ Enrichment pattern executed successfully"
else
    echo "❌ Enrichment pattern failed"
    cat /tmp/enrichment_output.txt
fi

echo
echo "=== All Hooks and Middleware Examples Tested ==="
echo
echo "To run examples manually:"
echo "1. cd examples/hooks-middleware && go run main.go advanced.go"
echo "2. cd examples/hooks-middleware/patterns && go run enrichment_pattern.go"
echo
echo "For more details, check the README files in each directory."
