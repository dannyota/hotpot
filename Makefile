.PHONY: help build clean test vet lint migrate-tool ingest-tool

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build all binaries to bin/
	@mkdir -p bin
	@echo "Building ingest..."
	@go build -o bin/ingest ./cmd/ingest
	@echo "Building migrate..."
	@go build -o bin/migrate ./cmd/migrate
	@echo "✅ Binaries built in bin/"

clean: ## Remove built binaries
	@rm -rf bin/
	@rm -f ingest migrate
	@echo "✅ Cleaned"

test: ## Run tests
	@go test ./... -v

vet: ## Run go vet
	@go vet ./...

lint: ## Run golangci-lint (requires golangci-lint installed)
	@golangci-lint run

generate: ## Generate ent code
	@cd pkg/storage && go generate
	@echo "✅ Ent code generated"

.DEFAULT_GOAL := help
