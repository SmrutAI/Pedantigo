# pedantigo/plugins — Makefile
#
# Dispatches each target to the concrete web plugin modules that actually exist:
# plugins/web/echo/Makefile and plugins/web/gin/Makefile. Each target below is
# an explicit pair of direct dispatches — no loop, no wildcard fan-out.
#
# When plugins/rpc/ or plugins/llm/ gain a real go.mod, do NOT turn these
# targets into loops. Add new, separately-named targets instead (e.g.
# build-rpc, build-llm), matching the root Makefile's <action>-core /
# <action>-plugins pattern one level down.

.PHONY: help build test test-verbose test-run test-clean-cache test-coverage \
        test-ci test-ci-cov vet fmt lint deps check clean install bench all pre-commit

.DEFAULT_GOAL := help

help: ## Show this help message
	@echo "pedantigo/plugins — dispatches to plugins/web/{echo,gin}/Makefile"

build: ## Build plugins/web/{echo,gin}
	@$(MAKE) -C web/echo build
	@$(MAKE) -C web/gin build

test: ## Run tests in plugins/web/{echo,gin}
	@$(MAKE) -C web/echo test
	@$(MAKE) -C web/gin test

test-verbose: ## Run tests (verbose) in plugins/web/{echo,gin}
	@$(MAKE) -C web/echo test-verbose
	@$(MAKE) -C web/gin test-verbose

test-run: ## Flexible test runner in plugins/web/{echo,gin} (RUN=TestName PKG=./path TIMEOUT=5m)
	@$(MAKE) -C web/echo test-run RUN=$(RUN) PKG=$(PKG) TIMEOUT=$(TIMEOUT)
	@$(MAKE) -C web/gin test-run RUN=$(RUN) PKG=$(PKG) TIMEOUT=$(TIMEOUT)

test-clean-cache: ## Clean Go test cache in plugins/web/{echo,gin}
	@$(MAKE) -C web/echo test-clean-cache
	@$(MAKE) -C web/gin test-clean-cache

test-coverage: ## Run tests with coverage in plugins/web/{echo,gin}
	@$(MAKE) -C web/echo test-coverage
	@$(MAKE) -C web/gin test-coverage

test-ci: ## CI: run tests with JUnit XML output in plugins/web/{echo,gin}
	@$(MAKE) -C web/echo test-ci
	@$(MAKE) -C web/gin test-ci

test-ci-cov: ## CI: run tests with coverage + JUnit XML output in plugins/web/{echo,gin}
	@$(MAKE) -C web/echo test-ci-cov
	@$(MAKE) -C web/gin test-ci-cov

vet: ## Run go vet in plugins/web/{echo,gin}
	@$(MAKE) -C web/echo vet
	@$(MAKE) -C web/gin vet

fmt: ## Format code in plugins/web/{echo,gin}
	@$(MAKE) -C web/echo fmt
	@$(MAKE) -C web/gin fmt

lint: ## Run golangci-lint in plugins/web/{echo,gin}
	@$(MAKE) -C web/echo lint
	@$(MAKE) -C web/gin lint

deps: ## Download and tidy Go dependencies in plugins/web/{echo,gin}
	@$(MAKE) -C web/echo deps
	@$(MAKE) -C web/gin deps

check: ## Run lint + test in plugins/web/{echo,gin}
	@$(MAKE) -C web/echo check
	@$(MAKE) -C web/gin check

bench: ## Run benchmarks in plugins/web/{echo,gin}
	@$(MAKE) -C web/echo bench
	@$(MAKE) -C web/gin bench

clean: ## Clean build artifacts in plugins/web/{echo,gin}
	@$(MAKE) -C web/echo clean
	@$(MAKE) -C web/gin clean

install: ## Install/update dependencies in plugins/web/{echo,gin}
	@$(MAKE) -C web/echo install
	@$(MAKE) -C web/gin install

all: ## Run fmt, vet, and test in plugins/web/{echo,gin}
	@$(MAKE) -C web/echo all
	@$(MAKE) -C web/gin all

pre-commit: ## Quick check before commit, in plugins/web/{echo,gin}
	@$(MAKE) -C web/echo pre-commit
	@$(MAKE) -C web/gin pre-commit
