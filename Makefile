# Makefile for dscli

# Build variables
APP_NAME := dscli
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go build flags
LDFLAGS := -ldflags "-X github.com/yourenyouyu/dscli/cmd.Version=$(VERSION)"

# Default target
.PHONY: all
all: build

# Build the application
.PHONY: build
build:
	@echo "Building $(APP_NAME)..."
	go build $(LDFLAGS) -o bin/$(APP_NAME) .
	@echo "Build complete: bin/$(APP_NAME)"

# Build for multiple platforms
.PHONY: build-all
build-all: clean
	@echo "Building for multiple platforms..."
	@mkdir -p dist
	# Windows
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(APP_NAME)-windows-amd64.exe .
	GOOS=windows GOARCH=386 go build $(LDFLAGS) -o dist/$(APP_NAME)-windows-386.exe .
	# macOS
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(APP_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(APP_NAME)-darwin-arm64 .
	# Linux
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(APP_NAME)-linux-amd64 .
	GOOS=linux GOARCH=386 go build $(LDFLAGS) -o dist/$(APP_NAME)-linux-386 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(APP_NAME)-linux-arm64 .
	@echo "Multi-platform build complete!"

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -rf bin/ dist/

# Install the application
.PHONY: install
install: build
	@echo "Installing $(APP_NAME)..."
	cp bin/$(APP_NAME) /usr/local/bin/
	@echo "Installation complete!"

# Show version information
.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Commit: $(COMMIT)"

# Development setup
.PHONY: dev-setup
dev-setup: deps
	@echo "Setting up development environment..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development setup complete!"

# Lint the code
.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run

# Format the code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run the application
.PHONY: run
run: build
	./bin/$(APP_NAME)

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build      - Build the application"
	@echo "  build-all  - Build for multiple platforms"
	@echo "  deps       - Install dependencies"
	@echo "  test       - Run tests"
	@echo "  clean      - Clean build artifacts"
	@echo "  install    - Install the application"
	@echo "  version    - Show version information"
	@echo "  dev-setup  - Setup development environment"
	@echo "  lint       - Run linter"
	@echo "  fmt        - Format code"
	@echo "  run        - Build and run the application"
	@echo "  help       - Show this help"