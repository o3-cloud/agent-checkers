APP := bin/agent-checkers
COVERAGE_FILE := coverage.out
COVERAGE_THRESHOLD := 80
GO_PACKAGES := ./...

.PHONY: help install install-tools fmt fmt-check vet lint test coverage vuln build docker-build compose-config tidy check clean

help: ## Show available commands
	@awk 'BEGIN {FS = ":.*## "}; /^[a-zA-Z_-]+:.*## / {printf "%-16s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: install-tools ## Install development tools and git hooks
	pre-commit install
	pre-commit install --hook-type commit-msg

install-tools: ## Install pinned Go development tools
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	go install golang.org/x/vuln/cmd/govulncheck@v1.1.4

fmt: ## Format Go files (spec: go.md:2-3)
	gofmt -w .
	goimports -w .

fmt-check: ## Check Go formatting without modifying files
	@test -z "$$(gofmt -l .)" || { gofmt -l .; exit 1; }
	@test -z "$$(goimports -l .)" || { goimports -l .; exit 1; }

vet: ## Run go vet (spec: go.md:15)
	go vet $(GO_PACKAGES)

lint: ## Run golangci-lint (spec: go.md:16)
	golangci-lint run $(GO_PACKAGES)

test: ## Run tests with race detection (spec: go.md:12)
	go test -race $(GO_PACKAGES)

coverage: ## Enforce application line coverage (spec: unit-testing.md:15)
	go test -race -coverpkg=./internal/... -coverprofile=$(COVERAGE_FILE) ./internal/...
	@coverage=$$(go tool cover -func=$(COVERAGE_FILE) | awk '/^total:/ {gsub("%", "", $$3); print $$3}'); \
	awk -v coverage="$$coverage" -v threshold="$(COVERAGE_THRESHOLD)" 'BEGIN { \
		printf "Total coverage: %.1f%% (minimum: %.1f%%)\n", coverage, threshold; \
		exit !(coverage >= threshold) \
	}'

vuln: ## Scan dependencies for known vulnerabilities
	govulncheck $(GO_PACKAGES)

build: ## Build the service binary
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(APP) ./src

docker-build: ## Build the production container with BuildKit
	DOCKER_BUILDKIT=1 docker build --tag agent-checkers:local .

compose-config: ## Validate Compose configuration
	docker compose config --quiet

tidy: ## Tidy and verify Go modules (spec: go.md:17)
	go mod tidy
	git diff --exit-code -- go.mod go.sum

test-bdd: ## Run BDD acceptance tests (spec: spec.md Gherkin)
	go test -v ./tests/... -run "^$$" -timeout 120s

check: fmt-check vet lint test coverage vuln ## Run all local quality gates

clean: ## Remove generated artifacts
	rm -rf bin $(COVERAGE_FILE) coverage.html
