#!/bin/bash

echo "=========================================="
echo "Integration Test Setup Verification"
echo "=========================================="
echo ""

# Check Docker
echo "1. Checking Docker..."
if docker ps > /dev/null 2>&1; then
    echo "   ✅ Docker is running"
else
    echo "   ❌ Docker is not running"
    exit 1
fi

# Check Go
echo ""
echo "2. Checking Go..."
if /usr/local/go/bin/go version > /dev/null 2>&1; then
    echo "   ✅ Go is installed: $(/usr/local/go/bin/go version)"
else
    echo "   ❌ Go is not installed"
    exit 1
fi

# Check dependencies
echo ""
echo "3. Checking dependencies..."
cd /home/tn-99712/IdeaProjects/place-search
if /usr/local/go/bin/go list -m github.com/testcontainers/testcontainers-go > /dev/null 2>&1; then
    echo "   ✅ testcontainers-go is installed"
else
    echo "   ❌ testcontainers-go is not installed"
    exit 1
fi

echo ""
echo "=========================================="
echo "Setup verification complete!"
echo "=========================================="
echo ""
echo "To run integration tests:"
echo "  make test-integration"
echo ""
echo "To run a single test:"
echo "  go test -v -tags=integration -run TestAPI_Search_QueryTooShort ./tests/integration/"
echo ""
