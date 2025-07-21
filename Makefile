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
test:
	go test ./...

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