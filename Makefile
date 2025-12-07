.PHONY: help build test test-verbose test-coverage vet fmt lint clean install run bench

# Default target
help:
	@echo "Available targets:"
	@echo "  make build         - Build the project"
	@echo "  make test          - Run all tests"
	@echo "  make test-verbose  - Run tests with verbose output"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make vet           - Run go vet"
	@echo "  make fmt           - Format code with gofmt"
	@echo "  make lint          - Run golangci-lint (requires installation)"
	@echo "  make bench         - Run benchmarks"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make install       - Install dependencies"
	@echo "  make all           - Run fmt, vet, and test"

# Build the project
build:
	@echo "Building..."
	go build -v ./...

# Run all tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with verbose output
test-verbose:
	@echo "Running tests (verbose)..."
	go test -v -count=1 ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Format code
fmt:
	@echo "Formatting code..."
	gofmt -s -w .
	go fmt ./...

# Run linter (requires golangci-lint)
lint:
	@echo "Running golangci-lint..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install with: brew install golangci-lint" && exit 1)
	golangci-lint run ./...

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	go clean
	rm -f coverage.out coverage.html

# Install/update dependencies
install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run all checks (fmt, vet, test)
all: fmt vet test
	@echo "All checks passed!"

# Quick check before commit
pre-commit: fmt vet test
	@echo "Pre-commit checks passed!"