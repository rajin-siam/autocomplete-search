.PHONY: test test-verbose test-coverage test-coverage-html test-unit test-integration test-all clean

# Detect Go binary
GO := $(shell which go || echo /usr/local/go/bin/go)

# Run all tests
test:
	$(GO) test ./...

# Run tests with verbose output
test-verbose:
	$(GO) test -v ./...

# Run tests with coverage
test-coverage:
	$(GO) test -cover ./...

# Generate HTML coverage report
test-coverage-html:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run only unit tests (service layer)
test-unit:
	$(GO) test -v -cover ./internal/application/

# Run only integration tests
test-integration:
	$(GO) test -v -tags=integration ./tests/integration/...

# Run all tests (unit + integration)
test-all:
	@echo "=========================================="
	@echo "Running Unit Tests..."
	@echo "=========================================="
	$(GO) test -v ./...
	@echo ""
	@echo "=========================================="
	@echo "Running Integration Tests..."
	@echo "=========================================="
	$(GO) test -v -tags=integration ./tests/integration/...
	@echo ""
	@echo "=========================================="
	@echo "All tests passed! ✅"
	@echo "=========================================="

# Run integration tests with coverage
test-integration-coverage:
	$(GO) test -v -tags=integration -cover ./tests/integration/...

# Run tests with race detection
test-race:
	$(GO) test -race ./...

# Run specific test
# Usage: make test-run TEST=TestSearchService_Search_ValidQuery
test-run:
	$(GO) test -v -run $(TEST) ./internal/application/

# Clean test artifacts
clean:
	rm -f coverage.out coverage.html

# Run tests and show coverage per function
test-coverage-func:
	$(GO) test -coverprofile=coverage.out ./internal/application/
	$(GO) tool cover -func=coverage.out

# Watch mode (requires entr: brew install entr or apt-get install entr)
test-watch:
	find . -name "*.go" | entr -c $(GO) test -v ./internal/application/
