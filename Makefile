.PHONY: help build build-migrate clean test vet generate genmigrate migrate dev-up dev-down dev-reset

NAME    ?= auto
SCHEMA  ?= pkg/storage/ent
MIGDIR  ?= deploy/migrations
DB      ?= hotpot_dev

ifeq ($(OS),Windows_NT)
  MKDIR_BIN = if not exist bin mkdir bin
  RM_BIN = if exist bin rmdir /s /q bin
  RM_LOOSE = if exist ingest.exe del /q ingest.exe & if exist migrate.exe del /q migrate.exe & if exist genmigrate.exe del /q genmigrate.exe
  BIN_EXT = .exe
else
  MKDIR_BIN = mkdir -p bin
  RM_BIN = rm -rf bin/
  RM_LOOSE = rm -f ingest migrate genmigrate
  BIN_EXT =
endif

help: ## Show this help
ifeq ($(OS),Windows_NT)
	@echo Available targets: help build build-migrate clean test vet generate genmigrate migrate dev-up dev-down dev-reset
else
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
endif

build:
	@$(MKDIR_BIN)
	@echo "Building ingest..."
	@go build -o bin/ingest$(BIN_EXT) ./cmd/ingest
	@echo "Production binaries built in bin/"

build-migrate:
	@$(MKDIR_BIN)
	@echo "Building migrate..."
	@go build -o bin/migrate$(BIN_EXT) ./cmd/migrate
	@echo "migrate built in bin/"

clean: ## Remove built binaries
	@$(RM_BIN)
	@$(RM_LOOSE)
	@echo "Cleaned"

test: ## Run tests
	@go test ./... -v

vet: ## Run go vet
	@go vet ./...

generate: ## Generate ent code
	@cd pkg/storage && go generate
	@echo "Ent code generated"

## ── Migrations ────────────────────────────────────────────

genmigrate: ## Generate migration SQL (NAME=description DB=dbname)
	@$(MKDIR_BIN)
	@echo "Building genmigrate..."
	@go build -o bin/genmigrate$(BIN_EXT) ./tools/genmigrate
	@echo "genmigrate built in bin/"
	@go run ./tools/genmigrate --schema $(SCHEMA) --out $(MIGDIR) --db $(DB) $(NAME)

migrate: ## Apply pending migrations
	@$(MKDIR_BIN)
	@echo "Building migrate..."
	@go build -o bin/migrate$(BIN_EXT) ./cmd/migrate
	@echo "migrate built in bin/"
	@go run ./cmd/migrate

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
