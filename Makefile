.PHONY: all build clean ingest normalize detect admin

# Build all services
all: build

# Build all available services
build: ingest

# Individual service targets
ingest:
	@mkdir -p bin
	go build -o bin/ingest ./cmd/ingest

normalize:
	@mkdir -p bin
	go build -o bin/normalize ./cmd/normalize

detect:
	@mkdir -p bin
	go build -o bin/detect ./cmd/detect

admin:
	@mkdir -p bin
	go build -o bin/admin ./cmd/admin

# Clean build artifacts
clean:
	rm -rf bin/

# Development helpers
tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

# Run checks before commit
check: fmt vet
	go build ./...
