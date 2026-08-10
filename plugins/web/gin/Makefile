# pedantigo/plugins/web/gin Makefile
# Run `make help` for available targets

# Use bash for PIPESTATUS support (required for test-run target)
SHELL := /bin/bash

.PHONY: help build test test-verbose test-run test-clean-cache test-coverage \
        test-ci test-ci-cov vet fmt lint deps check clean install bench all pre-commit

.DEFAULT_GOAL := help

help: ## Show this help message
	@echo "pedantigo/plugins/web/gin - Gin request plugin for pedantigo"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "\033[1mTest Runner Examples (make test-run):\033[0m"
	@echo "  make test-run                                  # All tests"
	@echo "  make test-run RUN=TestMyFunction               # Single test by name"
	@echo "  make test-run TIMEOUT=5m                       # Custom timeout"

RUN ?=
PKG ?= ./...
TIMEOUT ?= 10m

GO_TEST_FLAGS := -v -race -count=1

test-run: ## Flexible test runner (use RUN=TestName PKG=./path TIMEOUT=5m)
	@echo "═══════════════════════════════════════════════════════"
	@echo "Test Run: $$(date)"
	@echo "═══════════════════════════════════════════════════════"
	@echo "Config:"
	@echo "  RUN:     $(if $(RUN),$(RUN),(all tests))"
	@echo "  PKG:     $(PKG)"
	@echo "  TIMEOUT: $(TIMEOUT)"
	@echo "═══════════════════════════════════════════════════════"
	@echo ""
	@go test $(GO_TEST_FLAGS) \
		$(if $(RUN),-run=$(RUN),) \
		-timeout $(TIMEOUT) \
		$(PKG) 2>&1 | tee /tmp/pedantigo_gin_test_output.log; \
		EXIT_CODE=$${PIPESTATUS[0]}; \
		echo ""; \
		echo "═══════════════════════════════════════════════════════"; \
		echo "Exit code: $$EXIT_CODE"; \
		echo "Output saved to: /tmp/pedantigo_gin_test_output.log"; \
		echo "═══════════════════════════════════════════════════════"; \
		exit $$EXIT_CODE

build: ## Build the project
	@echo "Building..."
	go build -v ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

fmt: ## Format code with goimports + gofmt
	@echo "Formatting code..."
	@which goimports > /dev/null || (echo "goimports not installed. Install with: go install golang.org/x/tools/cmd/goimports@latest" && exit 1)
	goimports -w -local github.com/SmrutAI/pedantigo .
	gofmt -s -w .

GOLANGCI_LINT_VERSION := v2.12.2
GOLANGCI_LINT_BIN := $(shell go env GOPATH)/bin/golangci-lint

$(GOLANGCI_LINT_BIN): ## Install the pinned golangci-lint version if missing or outdated
	@if [ ! -x "$(GOLANGCI_LINT_BIN)" ] || ! "$(GOLANGCI_LINT_BIN)" version 2>/dev/null | grep -q "version $(GOLANGCI_LINT_VERSION:v%=%) "; then \
		echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."; \
		GOTOOLCHAIN=go1.25.12 go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
	fi

lint: $(GOLANGCI_LINT_BIN) ## Run golangci-lint
	@echo "Running golangci-lint..."
	GOTOOLCHAIN=go1.25.12 "$(GOLANGCI_LINT_BIN)" run ./...

deps: ## Download and tidy Go dependencies
	@echo "Tidying dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies up to date"

check: lint test ## Run lint + all tests

test: ## Run all tests (parallel with race detection)
	@echo "Running tests (parallel, race detection enabled)..."
	go test -race -parallel 8 -count=1 ./...

test-verbose: ## Run tests with verbose output
	@echo "Running tests (verbose)..."
	go test -v -race -parallel 8 -count=1 ./...

test-clean-cache: ## Clean Go test cache (forces recompilation)
	go clean -testcache
	@echo "Test cache cleaned"

test-coverage: ## Run tests with coverage report and threshold check
	@echo "Running tests with coverage..."
	go test -race -parallel 8 -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-ci: ## CI: Run tests with JUnit XML output
	@echo "Running tests (CI mode, JUnit XML output)..."
	go test -race -parallel 8 -count=1 -v ./... 2>&1 | go-junit-report -set-exit-code > test-results.xml

test-ci-cov: ## CI: Run tests with coverage + JUnit XML output
	@echo "Running tests with coverage (CI mode)..."
	go test -race -parallel 8 -v -coverprofile=coverage.out -covermode=atomic ./... 2>&1 | go-junit-report -set-exit-code > test-results.xml
	go tool cover -func=coverage.out | grep total | awk '{print $$3}' > coverage.txt

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	go clean
	rm -f coverage.out coverage.html coverage.txt test-results.xml coverage.xml

install: deps ## Install/update dependencies (alias for deps)

all: fmt vet test ## Run fmt, vet, and test
	@echo "All checks passed!"

pre-commit: fmt vet test ## Quick check before commit
	@echo "Pre-commit checks passed!"
