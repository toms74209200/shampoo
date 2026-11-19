.PHONY: all setup build check lint check-exceptions check-dependencies check-small-test-coverage check-side-effects check-comments check-test-helpers check-test-names check-fixed-values check-functions check-white-box check-scenarios fmt test-small test-medium test-acceptance test-tools help

all: check build

setup: ## Enable the git hooks in .githooks
	git config core.hooksPath .githooks

build: ## Build the binary
	go build -o shampoo .

check: lint check-exceptions check-dependencies test-small check-small-test-coverage check-side-effects check-comments check-test-helpers check-test-names check-fixed-values check-functions check-white-box check-scenarios test-tools ## Run all checks guarding a commit

lint: ## Run linters and format checks
	scripts/lint.sh

check-exceptions: ## Check that no exception other than nolint is used
	scripts/check-exceptions.sh

check-dependencies: ## Check that every module is in the allowlist
	scripts/check-dependencies.sh

check-small-test-coverage: ## Judge small-test coverage outside side-effect packages
	scripts/judge-small-test-coverage.sh out/small.cover

check-side-effects: ## Judge the system calls of small tests
	scripts/judge-side-effects.sh out/traces

check-comments: ## Check that code has no comments other than directives
	scripts/check-comments.sh

check-test-helpers: ## Check that test files declare only Test, Fuzz and Example functions
	scripts/check-test-helpers.sh

check-test-names: ## Check that every test is named when <target> with <given conditions> then returns <expected>
	scripts/check-test-names.sh

check-fixed-values: ## Check that literals in tests appear only in parameterized test tables
	scripts/check-fixed-values.sh

check-functions: ## Check that production code has none of the prohibited functions
	scripts/check-functions.sh

check-white-box: ## Check that white-box tests depend on their target and types only
	scripts/check-white-box.sh

check-scenarios: ## Check that every scenario without a matching @sha carries @skip
	scripts/check-scenarios.sh

fmt: ## Apply formatters
	scripts/golangci-lint.sh fmt
	cd tools && ../scripts/golangci-lint.sh fmt

test-small: ## Run small tests once, recording coverage and system calls
	scripts/test-small.sh

test-medium: ## Run medium tests, recording coverage
	mkdir -p out
	go test -tags=medium -timeout 300s -coverpkg=./... -coverprofile=out/medium.cover ./...

test-acceptance: ## Run acceptance tests against a binary built from the working tree
	uv run --directory tests/acceptance pytest

test-tools: ## Run small tests of the checks in tools
	go -C tools test -tags=small ./...

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "%-12s %s\n", $$1, $$2}'
