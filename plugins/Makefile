# pedantigo/plugins — Makefile
#
# Placeholder. No plugin sub-module has real build/test logic wired in yet
# (plugins/web/echo/ is created in a later task; plugins/rpc/ and plugins/llm/
# are empty .gitkeep placeholders). Every target below is a deliberate no-op
# so that the root orchestrator Makefile's fan-out (`make -C plugins <target>`)
# never fails while this directory has no real modules to build.
#
# As each plugin module gains a real go.mod, wire its build/test/lint here
# (mirror the fan-out pattern used in the root Makefile).

.PHONY: help build test test-verbose test-run test-clean-cache test-coverage \
        test-ci test-ci-cov vet fmt lint deps check clean install bench all pre-commit

.DEFAULT_GOAL := help

help: ## Show this help message
	@echo "pedantigo/plugins — no real modules wired in yet (no-op targets)"

build: ## (no-op) Build all plugin modules
	@true

test: ## (no-op) Run tests in all plugin modules
	@true

test-verbose: ## (no-op) Run tests (verbose) in all plugin modules
	@true

test-run: ## (no-op) Flexible test runner in all plugin modules
	@true

test-clean-cache: ## (no-op) Clean Go test cache in all plugin modules
	@true

test-coverage: ## (no-op) Run tests with coverage in all plugin modules
	@true

test-ci: ## (no-op) CI: run tests with JUnit XML output in all plugin modules
	@true

test-ci-cov: ## (no-op) CI: run tests with coverage + JUnit XML output in all plugin modules
	@true

vet: ## (no-op) Run go vet in all plugin modules
	@true

fmt: ## (no-op) Format code in all plugin modules
	@true

lint: ## (no-op) Run golangci-lint in all plugin modules
	@true

deps: ## (no-op) Download and tidy Go dependencies in all plugin modules
	@true

check: ## (no-op) Run lint + test in all plugin modules
	@true

bench: ## (no-op) Run benchmarks in all plugin modules
	@true

clean: ## (no-op) Clean build artifacts in all plugin modules
	@true

install: ## (no-op) Install/update dependencies in all plugin modules
	@true

all: ## (no-op) Run fmt, vet, and test in all plugin modules
	@true

pre-commit: ## (no-op) Quick check before commit, in all plugin modules
	@true
