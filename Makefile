.PHONY: help build test test-verbose test-coverage test-ci test-ci-cov vet fmt lint clean install run bench

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

# Run all tests (parallel with race detection)
test:
	@echo "Running tests (parallel, race detection enabled)..."
	go test -race -parallel 8 -count=1 ./...

# Run tests with verbose output
test-verbose:
	@echo "Running tests (verbose)..."
	go test -v -race -parallel 8 -count=1 ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -race -parallel 8 -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"
	@echo "Checking coverage threshold..."
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	THRESHOLD=85.0; \
	echo "Current coverage: $${COVERAGE}%"; \
	echo "Target coverage: $${THRESHOLD}%"; \
	if awk -v cov="$$COVERAGE" -v thresh="$$THRESHOLD" 'BEGIN {exit !(cov >= thresh)}'; then \
		echo "✓ Coverage check passed"; \
	else \
		echo "⚠ Coverage below target: $${COVERAGE}% < $${THRESHOLD}%"; \
		exit 1; \
	fi

# CI: Run tests with JUnit XML output
test-ci:
	@echo "Running tests (CI mode, JUnit XML output)..."
	go test -race -parallel 8 -count=1 -v ./... 2>&1 | go-junit-report -set-exit-code > test-results.xml

# CI: Run tests with coverage + JUnit XML output (single run)
test-ci-cov:
	@echo "Running tests with coverage (CI mode)..."
	go test -race -parallel 8 -v -coverprofile=coverage.out -covermode=atomic ./... 2>&1 | go-junit-report -set-exit-code > test-results.xml
	go tool cover -func=coverage.out | grep total | awk '{print $$3}' > coverage.txt

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Format code (uses goimports for import grouping)
fix_fmt:
	@echo "Formatting code..."
	@which goimports > /dev/null || (echo "goimports not installed. Install with: go install golang.org/x/tools/cmd/goimports@latest" && exit 1)
	goimports -w -local github.com/SmrutAI/pedantigo .
	gofmt -s -w .

# Format code (alias for compatibility)
fmt: fix_fmt

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
	rm -f coverage.out coverage.html coverage.txt test-results.xml

# Install/update dependencies
install:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run all checks (fmt, vet, test)
all: fmt vet test
	@echo "All checks passed!"

# Quick check before commit
pre-commit: fmt vet test-coverage
	@echo "Pre-commit checks passed!"