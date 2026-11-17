GOLANGCI_LINT := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0
BASE ?= origin/master
OUT := out

.PHONY: all setup build check lint check-exceptions check-dependencies check-small-test-coverage check-side-effects check-coverage check-comments check-test-helpers check-test-names fmt test-small help

all: check build

setup: ## Enable the git hooks in .githooks
	git config core.hooksPath .githooks

build: ## Build the binary
	go build -o shampoo .

check: lint check-exceptions check-dependencies test-small check-small-test-coverage check-side-effects check-comments check-test-helpers check-test-names ## Run all checks guarding a commit

lint: ## Run linters and format checks
	$(GOLANGCI_LINT) run --build-tags small,medium,large

check-exceptions: ## Check that no exception other than nolint is used
	scripts/check-exceptions.sh

check-dependencies: ## Check that every module is in the allowlist
	scripts/check-dependencies.sh

check-small-test-coverage: test-small ## Judge small-test coverage outside side-effect packages
	scripts/judge-small-test-coverage.sh $(OUT)/small.cover

check-side-effects: test-small ## Judge the system calls of small tests
	scripts/judge-side-effects.sh $(OUT)/traces

check-coverage: test-small ## Judge that coverage does not go down from BASE
	scripts/judge-coverage.sh $(OUT)/small.cover $(BASE)

check-comments: ## Check that code has no comments other than directives
	scripts/check-comments.sh

check-test-helpers: ## Check that test files declare only Test, Fuzz and Example functions
	scripts/check-test-helpers.sh

check-test-names: ## Check that every test is named when <target> with <given conditions> then returns <expected>
	scripts/check-test-names.sh

fmt: ## Apply formatters
	$(GOLANGCI_LINT) fmt

test-small: ## Run small tests once, recording coverage and system calls
	rm -rf $(OUT)/traces && mkdir -p $(OUT)/traces
	CGO_ENABLED=0 TRACE_DIR=$(abspath $(OUT)/traces) go test -tags=small -exec $(abspath scripts/trace-exec.sh) -coverpkg=./... -coverprofile=$(OUT)/small.cover ./...

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "%-12s %s\n", $$1, $$2}'
