.PHONY: dev build run test clean

# Development with live reload
dev:
	air

# Build the application
build:
	go build -o bin/server.exe cmd/server/main.go

# Run the application
run:
	go run cmd/server/main.go

# Run tests
test: test-unit test-integration test-e2e ## Run all tests

test-unit: ## Run unit tests
	go test -v ./internal/tests/unit/...

test-integration: ## Run integration tests
	go test -v ./internal/tests/integration/...

test-e2e: ## Run end-to-end tests
	go test -v ./internal/tests/e2e/...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./internal/tests/...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf tmp/
	rm -rf bin/
	rm -f build-errors.log
	rm -f tmp/main.exe

# Install Air (if not already installed)
install-air:
	go install github.com/air-verse/air@latest

# Install dependencies
deps:
	go mod download
	go mod tidy 