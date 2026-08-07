# pedantigo — root orchestrator Makefile
# This Makefile holds NO build/test/lint logic itself. Every action has exactly
# two targets: <action>-core (dispatches to pdcore/Makefile) and
# <action>-plugins (dispatches to plugins/Makefile). Each target is a single,
# direct dispatch — no loop, no fan-out, no bare composite target that runs
# both. Real logic lives in pdcore/Makefile and plugins/Makefile.

.PHONY: help \
        build-core build-plugins \
        test-core test-plugins \
        test-verbose-core test-verbose-plugins \
        test-run-core test-run-plugins \
        test-clean-cache-core test-clean-cache-plugins \
        test-coverage-core test-coverage-plugins \
        test-ci-core test-ci-plugins \
        test-ci-cov-core test-ci-cov-plugins \
        vet-core vet-plugins \
        fmt-core fmt-plugins \
        lint-core lint-plugins \
        deps-core deps-plugins \
        check-core check-plugins \
        bench-core bench-plugins \
        clean-core clean-plugins \
        install-core install-plugins \
        all-core all-plugins

# Default target
.DEFAULT_GOAL := help

# Test runner passthrough (used by test-run-core / test-run-plugins)
RUN ?=
PKG ?= ./...
TIMEOUT ?= 10m

help: ## Show this help message
	@echo "Pedantigo — orchestrator Makefile"
	@echo ""
	@echo "Every action has exactly two targets: <action>-core, <action>-plugins."
	@echo "Each is a single direct dispatch to pdcore/Makefile or plugins/Makefile."
	@echo "No loops, no fan-out, no combined target."
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ============================================
# BUILD
# ============================================

build-core: ## Build pdcore
	@$(MAKE) -C pdcore build

build-plugins: ## Build plugins
	@$(MAKE) -C plugins build

# ============================================
# TEST
# ============================================

test-core: ## Run tests in pdcore
	@$(MAKE) -C pdcore test

test-plugins: ## Run tests in plugins
	@$(MAKE) -C plugins test

test-verbose-core: ## Run tests (verbose) in pdcore
	@$(MAKE) -C pdcore test-verbose

test-verbose-plugins: ## Run tests (verbose) in plugins
	@$(MAKE) -C plugins test-verbose

test-run-core: ## Flexible test runner in pdcore (RUN=TestName PKG=./path TIMEOUT=5m)
	@$(MAKE) -C pdcore test-run RUN=$(RUN) PKG=$(PKG) TIMEOUT=$(TIMEOUT)

test-run-plugins: ## Flexible test runner in plugins (RUN=TestName PKG=./path TIMEOUT=5m)
	@$(MAKE) -C plugins test-run RUN=$(RUN) PKG=$(PKG) TIMEOUT=$(TIMEOUT)

test-clean-cache-core: ## Clean Go test cache in pdcore
	@$(MAKE) -C pdcore test-clean-cache

test-clean-cache-plugins: ## Clean Go test cache in plugins
	@$(MAKE) -C plugins test-clean-cache

test-coverage-core: ## Run tests with coverage in pdcore
	@$(MAKE) -C pdcore test-coverage

test-coverage-plugins: ## Run tests with coverage in plugins
	@$(MAKE) -C plugins test-coverage

# ============================================
# CI TARGETS
# ============================================

test-ci-core: ## CI: run tests with JUnit XML output in pdcore
	@$(MAKE) -C pdcore test-ci

test-ci-plugins: ## CI: run tests with JUnit XML output in plugins
	@$(MAKE) -C plugins test-ci

test-ci-cov-core: ## CI: run tests with coverage + JUnit XML output in pdcore
	@$(MAKE) -C pdcore test-ci-cov

test-ci-cov-plugins: ## CI: run tests with coverage + JUnit XML output in plugins
	@$(MAKE) -C plugins test-ci-cov

# ============================================
# QUALITY
# ============================================

vet-core: ## Run go vet in pdcore
	@$(MAKE) -C pdcore vet

vet-plugins: ## Run go vet in plugins
	@$(MAKE) -C plugins vet

fmt-core: ## Format code in pdcore
	@$(MAKE) -C pdcore fmt

fmt-plugins: ## Format code in plugins
	@$(MAKE) -C plugins fmt

lint-core: ## Run golangci-lint in pdcore
	@$(MAKE) -C pdcore lint

lint-plugins: ## Run golangci-lint in plugins
	@$(MAKE) -C plugins lint

deps-core: ## Download and tidy Go dependencies in pdcore
	@$(MAKE) -C pdcore deps

deps-plugins: ## Download and tidy Go dependencies in plugins
	@$(MAKE) -C plugins deps

check-core: ## Run lint + test in pdcore
	@$(MAKE) -C pdcore check

check-plugins: ## Run lint + test in plugins
	@$(MAKE) -C plugins check

# ============================================
# BENCHMARKS & MISC
# ============================================

bench-core: ## Run benchmarks in pdcore
	@$(MAKE) -C pdcore bench

bench-plugins: ## Run benchmarks in plugins
	@$(MAKE) -C plugins bench

clean-core: ## Clean build artifacts in pdcore
	@$(MAKE) -C pdcore clean

clean-plugins: ## Clean build artifacts in plugins
	@$(MAKE) -C plugins clean

install-core: ## Install/update dependencies in pdcore
	@$(MAKE) -C pdcore install

install-plugins: ## Install/update dependencies in plugins
	@$(MAKE) -C plugins install

all-core: ## Run fmt, vet, and test in pdcore
	@$(MAKE) -C pdcore all

all-plugins: ## Run fmt, vet, and test in plugins
	@$(MAKE) -C plugins all
