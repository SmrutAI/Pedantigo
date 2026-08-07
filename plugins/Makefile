# pedantigo/plugins — Makefile
#
# Dispatches each target to plugins/web/echo/Makefile — currently the only
# real plugin module (plugins/rpc/ and plugins/llm/ are empty .gitkeep
# placeholders with no go.mod, so there is nothing to dispatch to there yet).
# Each target below is a single, direct dispatch — no loop, no fan-out.
#
# When plugins/rpc/ or plugins/llm/ gain a real go.mod, do NOT turn these
# targets into loops. Add new, separately-named targets instead (e.g.
# build-rpc, build-llm), matching the root Makefile's <action>-core /
# <action>-plugins pattern one level down.

.PHONY: help build test test-verbose test-run test-clean-cache test-coverage \
        test-ci test-ci-cov vet fmt lint deps check clean install bench all pre-commit

.DEFAULT_GOAL := help

help: ## Show this help message
	@echo "pedantigo/plugins — dispatches to plugins/web/echo/Makefile"

build: ## Build plugins/web/echo
	@$(MAKE) -C web/echo build

test: ## Run tests in plugins/web/echo
	@$(MAKE) -C web/echo test

test-verbose: ## Run tests (verbose) in plugins/web/echo
	@$(MAKE) -C web/echo test-verbose

test-run: ## Flexible test runner in plugins/web/echo (RUN=TestName PKG=./path TIMEOUT=5m)
	@$(MAKE) -C web/echo test-run RUN=$(RUN) PKG=$(PKG) TIMEOUT=$(TIMEOUT)

test-clean-cache: ## Clean Go test cache in plugins/web/echo
	@$(MAKE) -C web/echo test-clean-cache

test-coverage: ## Run tests with coverage in plugins/web/echo
	@$(MAKE) -C web/echo test-coverage

test-ci: ## CI: run tests with JUnit XML output in plugins/web/echo
	@$(MAKE) -C web/echo test-ci

test-ci-cov: ## CI: run tests with coverage + JUnit XML output in plugins/web/echo
	@$(MAKE) -C web/echo test-ci-cov

vet: ## Run go vet in plugins/web/echo
	@$(MAKE) -C web/echo vet

fmt: ## Format code in plugins/web/echo
	@$(MAKE) -C web/echo fmt

lint: ## Run golangci-lint in plugins/web/echo
	@$(MAKE) -C web/echo lint

deps: ## Download and tidy Go dependencies in plugins/web/echo
	@$(MAKE) -C web/echo deps

check: ## Run lint + test in plugins/web/echo
	@$(MAKE) -C web/echo check

bench: ## Run benchmarks in plugins/web/echo
	@$(MAKE) -C web/echo bench

clean: ## Clean build artifacts in plugins/web/echo
	@$(MAKE) -C web/echo clean

install: ## Install/update dependencies in plugins/web/echo
	@$(MAKE) -C web/echo install

all: ## Run fmt, vet, and test in plugins/web/echo
	@$(MAKE) -C web/echo all

pre-commit: ## Quick check before commit, in plugins/web/echo
	@$(MAKE) -C web/echo pre-commit
