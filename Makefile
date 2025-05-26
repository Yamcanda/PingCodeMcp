# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=server
BINARY_UNIX=$(BINARY_NAME)_unix

# Build the application
build:
	$(GOBUILD) -o bin/$(BINARY_NAME) cmd/server/main.go

# Build for Linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o bin/$(BINARY_UNIX) cmd/server/main.go

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)
	rm -f bin/$(BINARY_UNIX)

# Run tests
test:
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) verify

# Tidy dependencies
tidy:
	$(GOMOD) tidy

# Run the application
run:
	$(GOCMD) run cmd/server/main.go

# Format code
fmt:
	$(GOCMD) fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Install development tools
install-tools:
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint

# Create directories
dirs:
	mkdir -p bin
	mkdir -p logs

# Development setup
dev-setup: dirs deps install-tools

# Docker build
docker-build:
	docker build -t pingcode-mcp-server .

# Docker run
docker-run:
	docker run -p 8080:8080 pingcode-mcp-server

.PHONY: build build-linux clean test test-coverage deps tidy run fmt lint install-tools dirs dev-setup docker-build docker-run 