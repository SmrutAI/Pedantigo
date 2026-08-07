# pedantigo — root orchestrator Makefile
# This Makefile holds NO build/test/lint logic itself. It fans out every target
# to each module's own Makefile (pdcore/Makefile, plugins/Makefile, ...).
# Each module owns its own real implementation of these targets.

MODULES := pdcore plugins

.PHONY: help build test test-verbose test-run test-clean-cache test-coverage \
        test-ci test-ci-cov vet fmt lint deps check clean install bench all pre-commit \
        $(MODULES)

# Default target
.DEFAULT_GOAL := help

help: ## Show this help message
	@echo "Pedantigo — orchestrator Makefile"
	@echo ""
	@echo "This Makefile fans out every target to: $(MODULES)"
	@echo "Real logic lives in each module's own Makefile (e.g. pdcore/Makefile)."
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ============================================
# FAN-OUT TARGETS
# Each target below runs the same-named target in every module's Makefile,
# via `make -C <module> <target>`, stopping on the first module that fails.
# ============================================

build: ## Build all modules
	@for m in $(MODULES); do \
		echo "=== build: $$m ==="; \
		$(MAKE) -C $$m build || exit 1; \
	done

test: ## Run tests in all modules
	@for m in $(MODULES); do \
		echo "=== test: $$m ==="; \
		$(MAKE) -C $$m test || exit 1; \
	done

test-verbose: ## Run tests (verbose) in all modules
	@for m in $(MODULES); do \
		echo "=== test-verbose: $$m ==="; \
		$(MAKE) -C $$m test-verbose || exit 1; \
	done

test-run: ## Flexible test runner in all modules (use RUN=TestName PKG=./path TIMEOUT=5m)
	@for m in $(MODULES); do \
		echo "=== test-run: $$m ==="; \
		$(MAKE) -C $$m test-run RUN=$(RUN) PKG=$(PKG) TIMEOUT=$(TIMEOUT) || exit 1; \
	done

test-clean-cache: ## Clean Go test cache in all modules
	@for m in $(MODULES); do \
		$(MAKE) -C $$m test-clean-cache || exit 1; \
	done

test-coverage: ## Run tests with coverage in all modules
	@for m in $(MODULES); do \
		echo "=== test-coverage: $$m ==="; \
		$(MAKE) -C $$m test-coverage || exit 1; \
	done

test-ci: ## CI: run tests with JUnit XML output in all modules
	@for m in $(MODULES); do \
		echo "=== test-ci: $$m ==="; \
		$(MAKE) -C $$m test-ci || exit 1; \
	done

test-ci-cov: ## CI: run tests with coverage + JUnit XML output in all modules
	@for m in $(MODULES); do \
		echo "=== test-ci-cov: $$m ==="; \
		$(MAKE) -C $$m test-ci-cov || exit 1; \
	done

vet: ## Run go vet in all modules
	@for m in $(MODULES); do \
		echo "=== vet: $$m ==="; \
		$(MAKE) -C $$m vet || exit 1; \
	done

fmt: ## Format code in all modules
	@for m in $(MODULES); do \
		echo "=== fmt: $$m ==="; \
		$(MAKE) -C $$m fmt || exit 1; \
	done

lint: ## Run golangci-lint in all modules
	@for m in $(MODULES); do \
		echo "=== lint: $$m ==="; \
		$(MAKE) -C $$m lint || exit 1; \
	done

deps: ## Download and tidy Go dependencies in all modules
	@for m in $(MODULES); do \
		echo "=== deps: $$m ==="; \
		$(MAKE) -C $$m deps || exit 1; \
	done

check: ## Run lint + test in all modules
	@for m in $(MODULES); do \
		echo "=== check: $$m ==="; \
		$(MAKE) -C $$m check || exit 1; \
	done

bench: ## Run benchmarks in all modules
	@for m in $(MODULES); do \
		echo "=== bench: $$m ==="; \
		$(MAKE) -C $$m bench || exit 1; \
	done

clean: ## Clean build artifacts in all modules
	@for m in $(MODULES); do \
		$(MAKE) -C $$m clean || exit 1; \
	done

install: ## Install/update dependencies in all modules
	@for m in $(MODULES); do \
		$(MAKE) -C $$m install || exit 1; \
	done

all: ## Run fmt, vet, and test in all modules
	@for m in $(MODULES); do \
		echo "=== all: $$m ==="; \
		$(MAKE) -C $$m all || exit 1; \
	done
	@echo "All checks passed!"

pre-commit: ## Quick check before commit, in all modules
	@for m in $(MODULES); do \
		echo "=== pre-commit: $$m ==="; \
		$(MAKE) -C $$m pre-commit || exit 1; \
	done
	@echo "Pre-commit checks passed!"
