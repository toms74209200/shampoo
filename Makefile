GOLANGCI_LINT := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0

.PHONY: all setup build check lint check-exceptions check-dependencies check-small-test-coverage check-side-effects check-comments check-test-helpers check-test-names check-fixed-values check-functions fmt test-small test-tools help

all: check build

setup: ## Enable the git hooks in .githooks
	git config core.hooksPath .githooks

build: ## Build the binary
	go build -o shampoo .

check: lint check-exceptions check-dependencies test-small check-small-test-coverage check-side-effects check-comments check-test-helpers check-test-names check-fixed-values check-functions test-tools ## Run all checks guarding a commit

lint: ## Run linters and format checks
	$(GOLANGCI_LINT) run --build-tags small,medium,large
	cd tools && $(GOLANGCI_LINT) run --build-tags small,medium,large

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

fmt: ## Apply formatters
	$(GOLANGCI_LINT) fmt
	cd tools && $(GOLANGCI_LINT) fmt

test-small: ## Run small tests once, recording coverage and system calls
	strace -qq -o /dev/null true
	rm -rf out/traces && mkdir -p out/traces
	CGO_ENABLED=0 TRACE_DIR=$(CURDIR)/out/traces go test -tags=small -exec $(CURDIR)/scripts/trace-exec.sh -coverpkg=./... -coverprofile=out/small.cover ./...

test-tools: ## Run small tests of the checks in tools
	go -C tools test -tags=small ./...

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "%-12s %s\n", $$1, $$2}'
