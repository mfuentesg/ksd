.PHONY: build test clean lint fmt vet benchmark install help

# Build variables
BINARY_NAME=ksd
VERSION?=$(shell git describe --tags --always --dirty)
COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_TIME)"

# Default target
all: fmt vet lint test build

# Build the binary
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...

# Run benchmarks
benchmark:
	go test -bench=. -benchmem ./...

# Clean build artifacts
clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out

# Lint the code
lint:
	golangci-lint run

# Format the code
fmt:
	go fmt ./...

# Vet the code
vet:
	go vet ./...

# Install the binary
install:
	go install $(LDFLAGS) .

# Update dependencies
deps:
	go mod tidy
	go mod download

# Generate coverage report
coverage: test
	go tool cover -html=coverage.out -o coverage.html

# Run security scan
security:
	gosec ./...

# Release (dry-run)
release-dry-run:
	goreleaser release --snapshot --clean

# Release
release:
	goreleaser release --clean

# Check GoReleaser config
release-check:
	goreleaser check

# Install Krew plugin locally (for testing)
krew-install-local: release-dry-run
	kubectl krew install --manifest=dist/krew/ksd.yaml

# Uninstall Krew plugin
krew-uninstall:
	kubectl krew uninstall ksd

# Test Krew plugin
krew-test: krew-install-local
	kubectl ksd --version

# Help
help:
	@echo "Available targets:"
	@echo "  build             - Build the binary"
	@echo "  test              - Run tests"
	@echo "  benchmark         - Run benchmarks"
	@echo "  clean             - Clean build artifacts"
	@echo "  lint              - Run linter"
	@echo "  fmt               - Format code"
	@echo "  vet               - Run go vet"
	@echo "  install           - Install binary"
	@echo "  deps              - Update dependencies"
	@echo "  coverage          - Generate coverage report"
	@echo "  security          - Run security scan"
	@echo "  release-dry-run   - Test release process"
	@echo "  release           - Create release"
	@echo "  release-check     - Check GoReleaser config"
	@echo "  krew-install-local - Install Krew plugin locally"
	@echo "  krew-uninstall    - Uninstall Krew plugin"
	@echo "  krew-test         - Test Krew plugin"
	@echo "  help              - Show this help"