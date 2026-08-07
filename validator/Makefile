# pedantigo Makefile
# Run `make help` for available targets

# Use bash for PIPESTATUS support (required for test-run target)
SHELL := /bin/bash

.PHONY: help build test test-verbose test-run test-clean-cache test-coverage \
        test-ci test-ci-cov vet fmt lint deps check clean install bench

# Default target
.DEFAULT_GOAL := help

# ============================================
# HELP
# ============================================

help: ## Show this help message
	@echo "Pedantigo - Pydantic-inspired validation library for Go"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "\033[1mTest Runner Examples (make test-run):\033[0m"
	@echo "  make test-run                                  # All tests"
	@echo "  make test-run RUN=TestMyFunction               # Single test by name"
	@echo "  make test-run RUN=TestUnmarshal                # Tests matching pattern"
	@echo "  make test-run PKG=./validator/internal/constraints/...   # Tests in specific package"
	@echo "  make test-run TIMEOUT=5m                       # Custom timeout"

# ============================================
# FLEXIBLE TEST RUNNER
# ============================================
# Usage Examples:
#   make test-run                                  # All tests
#   make test-run RUN=TestMyFunction               # Single test by name
#   make test-run PKG=./internal/constraints/...   # Tests in specific package
#   make test-run TIMEOUT=5m                       # Custom timeout

# Test runner defaults
RUN ?=
PKG ?= ./...
TIMEOUT ?= 10m

# Common go test flags
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
		$(PKG) 2>&1 | tee /tmp/pedantigo_test_output.log; \
		EXIT_CODE=$${PIPESTATUS[0]}; \
		echo ""; \
		echo "═══════════════════════════════════════════════════════"; \
		echo "Exit code: $$EXIT_CODE"; \
		echo "Output saved to: /tmp/pedantigo_test_output.log"; \
		echo "═══════════════════════════════════════════════════════"; \
		exit $$EXIT_CODE

# ============================================
# BUILD & QUALITY
# ============================================

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

# Pinned to match the exact version CI installs (see .github/workflows/ci.yml) —
# invoked by direct path, not via PATH resolution, so a stale golangci-lint
# elsewhere on PATH (e.g. an older Homebrew install) can never silently shadow
# it and let a local `make lint` pass on issues CI would catch.
GOLANGCI_LINT_VERSION := v2.12.2
GOLANGCI_LINT_BIN := $(shell go env GOPATH)/bin/golangci-lint

$(GOLANGCI_LINT_BIN): ## Install the pinned golangci-lint version if missing or outdated
	@if [ ! -x "$(GOLANGCI_LINT_BIN)" ] || ! "$(GOLANGCI_LINT_BIN)" version 2>/dev/null | grep -q "version $(GOLANGCI_LINT_VERSION:v%=%) "; then \
		echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."; \
		GOTOOLCHAIN=go1.25.12 go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
	fi

lint: $(GOLANGCI_LINT_BIN) ## Run golangci-lint
	@echo "Running golangci-lint..."
	# golangci-lint's binary is built with go1.25 and doesn't yet support go1.26
	# (tracked upstream: golangci/golangci-lint#6272). Force its internal `go`
	# subprocess calls to use go1.25 so package analysis doesn't panic, while
	# the rest of the toolchain (build, vet, test) stays on go1.26.
	GOTOOLCHAIN=go1.25.12 "$(GOLANGCI_LINT_BIN)" run ./...

deps: ## Download and tidy Go dependencies
	@echo "Tidying dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies up to date"

check: lint test ## Run lint + all tests

# ============================================
# TESTING
# ============================================

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
	@echo "Checking coverage threshold..."
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	THRESHOLD=85.0; \
	echo "Current coverage: $${COVERAGE}%"; \
	echo "Target coverage: $${THRESHOLD}%"; \
	if awk -v cov="$$COVERAGE" -v thresh="$$THRESHOLD" 'BEGIN {exit !(cov >= thresh)}'; then \
		echo "Coverage check passed"; \
	else \
		echo "Coverage below target: $${COVERAGE}% < $${THRESHOLD}%"; \
		exit 1; \
	fi

# ============================================
# CI TARGETS
# ============================================

test-ci: ## CI: Run tests with JUnit XML output
	@echo "Running tests (CI mode, JUnit XML output)..."
	go test -race -parallel 8 -count=1 -v ./... 2>&1 | go-junit-report -set-exit-code > test-results.xml

test-ci-cov: ## CI: Run tests with coverage + JUnit XML output
	@echo "Running tests with coverage (CI mode)..."
	go test -race -parallel 8 -v -coverprofile=coverage.out -covermode=atomic ./... 2>&1 | go-junit-report -set-exit-code > test-results.xml
	go tool cover -func=coverage.out | grep total | awk '{print $$3}' > coverage.txt

# ============================================
# BENCHMARKS & MISC
# ============================================

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

pre-commit: fmt vet test-coverage ## Quick check before commit
	@echo "Pre-commit checks passed!"
