# pedantigo/plugins — Makefile
#
# Dispatches each target to the concrete web plugin modules that actually exist:
# plugins/web/pedantigoecho/v2/Makefile and plugins/web/pedantigogin/v2/Makefile.
# Each target below is an explicit pair of direct dispatches — no loop, no
# wildcard fan-out.
#
# When plugins/rpc/ or plugins/llm/ gain a real go.mod, do NOT turn these
# targets into loops. Add new, separately-named targets instead (e.g.
# build-rpc, build-llm), matching the root Makefile's <action>-core /
# <action>-plugins pattern one level down.

.PHONY: help build test test-verbose test-run test-clean-cache test-coverage \
        test-ci test-ci-cov vet fmt lint deps check clean install bench all pre-commit

.DEFAULT_GOAL := help

help: ## Show this help message
	@echo "pedantigo/plugins — dispatches to plugins/web/{pedantigoecho,pedantigogin}/v2/Makefile"

build: ## Build plugins/web/{pedantigoecho,pedantigogin}/v2
	@$(MAKE) -C web/pedantigoecho/v2 build
	@$(MAKE) -C web/pedantigogin/v2 build

test: ## Run tests in plugins/web/{pedantigoecho,pedantigogin}/v2
	@$(MAKE) -C web/pedantigoecho/v2 test
	@$(MAKE) -C web/pedantigogin/v2 test

test-verbose: ## Run tests (verbose) in plugins/web/{pedantigoecho,pedantigogin}/v2
	@$(MAKE) -C web/pedantigoecho/v2 test-verbose
	@$(MAKE) -C web/pedantigogin/v2 test-verbose

test-run: ## Flexible test runner in plugins/web/{pedantigoecho,pedantigogin}/v2 (RUN=TestName PKG=./path TIMEOUT=5m)
	@$(MAKE) -C web/pedantigoecho/v2 test-run RUN=$(RUN) PKG=$(PKG) TIMEOUT=$(TIMEOUT)
	@$(MAKE) -C web/pedantigogin/v2 test-run RUN=$(RUN) PKG=$(PKG) TIMEOUT=$(TIMEOUT)

test-clean-cache: ## Clean Go test cache in plugins/web/{pedantigoecho,pedantigogin}/v2
	@$(MAKE) -C web/pedantigoecho/v2 test-clean-cache
	@$(MAKE) -C web/pedantigogin/v2 test-clean-cache

test-coverage: ## Run tests with coverage in plugins/web/{pedantigoecho,pedantigogin}/v2
	@$(MAKE) -C web/pedantigoecho/v2 test-coverage
	@$(MAKE) -C web/pedantigogin/v2 test-coverage

test-ci: ## CI: run tests with JUnit XML output in plugins/web/{pedantigoecho,pedantigogin}/v2
	@$(MAKE) -C web/pedantigoecho/v2 test-ci
	@$(MAKE) -C web/pedantigogin/v2 test-ci

test-ci-cov: ## CI: run tests with coverage + JUnit XML output in plugins/web/{pedantigoecho,pedantigogin}/v2
	@$(MAKE) -C web/pedantigoecho/v2 test-ci-cov
	@$(MAKE) -C web/pedantigogin/v2 test-ci-cov

vet: ## Run go vet in plugins/web/{pedantigoecho,pedantigogin}/v2
	@$(MAKE) -C web/pedantigoecho/v2 vet
	@$(MAKE) -C web/pedantigogin/v2 vet

fmt: ## Format code in plugins/web/{pedantigoecho,pedantigogin}/v2
	@$(MAKE) -C web/pedantigoecho/v2 fmt
	@$(MAKE) -C web/pedantigogin/v2 fmt

lint: ## Run golangci-lint in plugins/web/{pedantigoecho,pedantigogin}/v2
	@$(MAKE) -C web/pedantigoecho/v2 lint
	@$(MAKE) -C web/pedantigogin/v2 lint

deps: ## Download and tidy Go dependencies in plugins/web/{pedantigoecho,pedantigogin}/v2
	@$(MAKE) -C web/pedantigoecho/v2 deps
	@$(MAKE) -C web/pedantigogin/v2 deps

check: ## Run lint + test in plugins/web/{pedantigoecho,pedantigogin}/v2
	@$(MAKE) -C web/pedantigoecho/v2 check
	@$(MAKE) -C web/pedantigogin/v2 check

bench: ## Run benchmarks in plugins/web/{pedantigoecho,pedantigogin}/v2
	@$(MAKE) -C web/pedantigoecho/v2 bench
	@$(MAKE) -C web/pedantigogin/v2 bench

clean: ## Clean build artifacts in plugins/web/{pedantigoecho,pedantigogin}/v2
	@$(MAKE) -C web/pedantigoecho/v2 clean
	@$(MAKE) -C web/pedantigogin/v2 clean

install: ## Install/update dependencies in plugins/web/{pedantigoecho,pedantigogin}/v2
	@$(MAKE) -C web/pedantigoecho/v2 install
	@$(MAKE) -C web/pedantigogin/v2 install

all: ## Run fmt, vet, and test in plugins/web/{pedantigoecho,pedantigogin}/v2
	@$(MAKE) -C web/pedantigoecho/v2 all
	@$(MAKE) -C web/pedantigogin/v2 all

pre-commit: ## Quick check before commit, in plugins/web/{pedantigoecho,pedantigogin}/v2
	@$(MAKE) -C web/pedantigoecho/v2 pre-commit
	@$(MAKE) -C web/pedantigogin/v2 pre-commit
