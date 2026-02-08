.PHONY: help build clean test vet lint generate dev-up dev-down dev-reset

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

## ── Dev Infrastructure ──────────────────────────────────────

dev-up: ## Start dev infrastructure (PostgreSQL, Redis, Temporal, Metabase)
	@docker compose -f deploy/dev/docker-compose.yml up -d
	@echo "Dev infrastructure starting... use 'docker compose -f deploy/dev/docker-compose.yml ps' to check status"

dev-down: ## Stop dev infrastructure
	@docker compose -f deploy/dev/docker-compose.yml down

dev-reset: ## Stop dev infrastructure and destroy all data
	@docker compose -f deploy/dev/docker-compose.yml down -v
	@echo "Dev infrastructure stopped and volumes removed"

.DEFAULT_GOAL := help
