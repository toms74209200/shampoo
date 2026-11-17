GOLANGCI_LINT := go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0

.PHONY: all setup build check lint fmt test-small help

all: check build

setup: ## Enable the git hooks in .githooks
	git config core.hooksPath .githooks

build: ## Build the binary
	go build -o shampoo .

check: lint test-small ## Run all checks guarding a commit

lint: ## Run linters and format checks
	$(GOLANGCI_LINT) run --build-tags small,medium,large

fmt: ## Apply formatters
	$(GOLANGCI_LINT) fmt

test-small: ## Run small size tests
	go test -tags=small ./...

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "%-12s %s\n", $$1, $$2}'
