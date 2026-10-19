.PHONY: dev build run test test-integration test-e2e vet fmt swagger check check-go check-swagger check-frontend clean install-air deps

# swag version pinned in go.mod, so generated docs are reproducible
SWAG_VERSION = $(shell go list -m -f '{{.Version}}' github.com/swaggo/swag)

# Development with live reload
dev:
	air

# Build the application
build:
	go build -o bin/server ./cmd/server

# Run the application
run:
	go run ./cmd/server

# Unit tests (no services needed)
test:
	go test ./...

# Unit and PostgreSQL integration tests; each test gets a throwaway database
# on the server from the DB_* settings (start it with `docker compose up -d --wait`)
test-integration:
	go test -tags integration ./...

# Browser smoke tests against the local stack (needs Google Chrome)
test-e2e:
	cd frontend && npm run test:e2e

# Static analysis
vet:
	go vet ./...

# Format Go code
fmt:
	go fmt ./...

# Regenerate Swagger docs in docs/
swagger:
	go run github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION) init -g cmd/server/main.go -o docs

# Run every check CI runs (needs PostgreSQL running and `npm ci` in frontend/)
check: check-go check-frontend

check-go: check-swagger
	@test -z "$$(gofmt -l cmd internal pkg docs)" || { echo "Not gofmt-formatted:"; gofmt -l cmd internal pkg docs; exit 1; }
	go vet -tags integration ./...
	go test -tags integration ./...

# Fail when the committed Swagger files are stale
check-swagger:
	@tmp=$$(mktemp -d) && \
	go run github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION) init --quiet -g cmd/server/main.go -o $$tmp --packageName docs && \
	for f in docs.go swagger.json swagger.yaml; do \
		cmp -s $$tmp/$$f docs/$$f || { echo "docs/$$f is stale; run make swagger"; rm -rf $$tmp; exit 1; }; \
	done; \
	rm -rf $$tmp

check-frontend:
	cd frontend && npm run lint && npm run build

# Clean build artifacts
clean:
	rm -rf bin/ tmp/ build-errors.log

# Install Air (if not already installed)
install-air:
	go install github.com/air-verse/air@latest

# Install dependencies
deps:
	go mod download
	go mod tidy
