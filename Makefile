# Makefile for go-sqlite-regexp

.PHONY: all build test test-race test-cover clean examples help

# Default target
all: test build

# Build the package
build:
	@echo "Building package..."
	@CGO_ENABLED=1 go build -v ./...

# Run tests
test:
	@echo "Running tests..."
	@CGO_ENABLED=1 go test -v ./...

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	@CGO_ENABLED=1 go test -race -v ./...

# Run tests with coverage
test-cover:
	@echo "Running tests with coverage..."
	@CGO_ENABLED=1 go test -cover -v ./...

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	@CGO_ENABLED=1 go test -bench=. -v ./...

# Build examples
examples:
	@echo "Building examples..."
	@cd examples && CGO_ENABLED=1 go build -o example example.go

# Run examples
run-examples: examples
	@echo "Running examples..."
	@cd examples && ./example

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@go clean ./...
	@rm -f examples/example

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download

# Generate documentation
docs:
	@echo "Generating documentation..."
	@go doc -all .

# Check for security vulnerabilities
security:
	@echo "Checking for security vulnerabilities..."
	@govulncheck ./...

# Full CI pipeline
ci: fmt tidy test-race test-cover build examples

# Help
help:
	@echo "Available targets:"
	@echo "  all        - Run tests and build (default)"
	@echo "  build      - Build the package"
	@echo "  test       - Run tests"
	@echo "  test-race  - Run tests with race detection"
	@echo "  test-cover - Run tests with coverage"
	@echo "  bench      - Run benchmarks"
	@echo "  examples   - Build examples"
	@echo "  run-examples - Build and run examples"
	@echo "  clean      - Clean build artifacts"
	@echo "  fmt        - Format code"
	@echo "  lint       - Lint code"
	@echo "  tidy       - Tidy dependencies"
	@echo "  deps       - Install dependencies"
	@echo "  docs       - Generate documentation"
	@echo "  security   - Check for security vulnerabilities"
	@echo "  ci         - Run full CI pipeline"
	@echo "  help       - Show this help"

