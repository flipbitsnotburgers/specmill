.PHONY: build test clean fmt lint test-verbose test-coverage help

# Default target
.DEFAULT_GOAL := help

# Build the server binary
build:
	go build -o specmill-server .

# Run unit tests
test:
	go test ./...

# Run tests with verbose output
test-verbose:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -cover ./...
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format Go code
fmt:
	go fmt ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -f specmill-server
	rm -f coverage.out coverage.html
	go clean

# Download dependencies
deps:
	go mod download
	go mod tidy

# Run the server with Petstore example
run-example: build
	./specmill-server -spec examples/petstore.yaml

# Quick test with Petstore
test-petstore: build
	@echo "Testing MCP server with Petstore spec..."
	@echo '{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"2025-06-18"},"id":1}' | ./specmill-server -spec examples/petstore.yaml 2>/dev/null | head -1
	@echo '{"jsonrpc":"2.0","method":"tools/list","params":{},"id":2}' | ./specmill-server -spec examples/petstore.yaml 2>/dev/null | head -2 | tail -1 | jq '.result.tools | length' | xargs echo "Number of tools generated:"

# Help target
help:
	@echo "Specmill - OpenAPI to MCP Server Generator"
	@echo ""
	@echo "Available targets:"
	@echo "  build          - Build the server binary"
	@echo "  test           - Run unit tests"
	@echo "  test-verbose   - Run tests with detailed output"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  fmt            - Format Go code"
	@echo "  lint           - Run linter (requires golangci-lint)"
	@echo "  clean          - Remove build artifacts"
	@echo "  deps           - Download and tidy dependencies"
	@echo "  run-example    - Run server with Petstore example"
	@echo "  test-petstore  - Quick test with Petstore spec"
	@echo "  help           - Show this help message"