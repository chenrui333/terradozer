PKG_LIST              := $(shell go list ./...)
GOLANGCI_LINT_VERSION := v2.10.1

.PHONY: setup
setup: ## Install build, test, and lint dependencies
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./bin ${GOLANGCI_LINT_VERSION}
	go install github.com/golang/mock/mockgen@v1.4.0
	go install github.com/hashicorp/terraform@v0.12.28
	curl -sSfL https://raw.githubusercontent.com/jckuester/go-acc/master/install.sh | sh -s v0.2.1

.PHONY: lint
lint: ## Run some static code analysis
	./bin/golangci-lint run

.PHONY: go-mod-tidy
go-mod-tidy: ## Clean go.mod
	@go mod tidy -v
	@git diff HEAD
	@git diff-index --quiet HEAD

.PHONY: fmt
fmt: ## Run gofmt on goimports all files
	gofmt -w -l -s .
	goimports -w -l .

PHONY: generate
generate: ## Run go generate
	go generate

.PHONY: test
test: ## Run unit tests
	go clean -testcache ${PKG_LIST}
	go test -v -p 1 -short -race ${PKG_LIST}

.PHONY: test-all
test-all: ## Run tests (including acceptance and integration tests)
	go clean -testcache ${PKG_LIST}
	./bin/go-acc ${PKG_LIST} -- -v -p 1 -race -failfast -timeout 20m

.PHONY: build
build: ## Build binary
	go build

.PHONY: build
ci: generate build test-all lint # Run all the tests and code checks
