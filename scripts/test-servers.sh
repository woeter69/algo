#!/bin/bash

# Test script to verify the Go server builds and tests pass

echo "🧪 Testing AlumniGo Go Server..."
echo "=================================="

# Test Go build
echo "1️⃣  Testing Go server build..."
cd app/src
if go build -o /tmp/test-algo-server .; then
    echo "✅ Go server builds successfully"
    rm -f /tmp/test-algo-server
else
    echo "❌ Go server build failed"
    exit 1
fi

# Run tests
echo "2️⃣  Running Go tests..."
if go test -v ./...; then
    echo "✅ All tests pass"
else
    echo "❌ Some tests failed"
    exit 1
fi

# Verify dependencies
echo "3️⃣  Checking Go module dependencies..."
if go mod verify; then
    echo "✅ Go dependencies verified"
else
    echo "❌ Go dependencies have issues"
    exit 1
fi

# Run vet
echo "4️⃣  Running go vet..."
if go vet ./...; then
    echo "✅ go vet passed"
else
    echo "⚠️  go vet reported issues (non-fatal)"
fi

cd ../..

echo ""
echo "✅ All checks passed!"
echo "🚀 Ready to run: ./scripts/start-all.sh"
