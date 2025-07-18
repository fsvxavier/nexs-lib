#!/bin/bash

# Test script for all logger examples

echo "=== Testing Logger Examples ==="
echo

# Test Default Provider Example
echo "--- Testing Default Provider Example ---"
cd examples/default-provider
if go run main.go >/dev/null 2>&1; then
    echo "✅ Default provider example: PASSED"
else
    echo "❌ Default provider example: FAILED"
fi

# Test Basic Example
echo "--- Testing Basic Example ---"
cd ../basic
if go run main.go >/dev/null 2>&1; then
    echo "✅ Basic example: PASSED"
else
    echo "❌ Basic example: FAILED"
fi

# Test Advanced Example
echo "--- Testing Advanced Example ---"
cd ../advanced
if go run main.go >/dev/null 2>&1; then
    echo "✅ Advanced example: PASSED"
else
    echo "❌ Advanced example: FAILED"
fi

# Test Multi-Provider Example
echo "--- Testing Multi-Provider Example ---"
cd ../multi-provider
if go run main.go >/dev/null 2>&1; then
    echo "✅ Multi-provider example: PASSED"
else
    echo "❌ Multi-provider example: FAILED"
fi

# Test Benchmark Example
echo "--- Testing Benchmark Example ---"
cd ../benchmark
if go run main.go 2>/dev/null >/dev/null; then
    echo "✅ Benchmark example: PASSED"
else
    echo "❌ Benchmark example: FAILED"
fi

echo
echo "=== All examples tested ==="
