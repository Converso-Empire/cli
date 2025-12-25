# Converso CLI Makefile
# Build automation and development tasks

.PHONY: help build build-all clean test lint format install uninstall

# Variables
VERSION := $(shell git describe --tags --always 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GO_VERSION := 1.21
BUILD_DIR := dist

# Default target
help:
	@echo "🚀 Converso CLI Build System"
	@echo "=========================="
	@echo ""
	@echo "Available targets:"
	@echo "  help        - Show this help message"
	@echo "  build       - Build for current platform"
	@echo "  build-all   - Build for all platforms"
	@echo "  clean       - Clean build artifacts"
	@echo "  test        - Run tests"
	@echo "  lint        - Run linter"
	@echo "  format      - Format code"
	@echo "  install     - Install CLI locally"
	@echo "  uninstall   - Uninstall CLI"
	@echo "  setup       - Setup development environment"
	@echo ""

# Build for current platform
build:
	@echo "🔨 Building Converso CLI for $(GOOS)/$(GOARCH)..."
	@go build -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE) -s -w" -o $(BUILD_DIR)/converso ./cmd/converso/
	@echo "✅ Build completed: $(BUILD_DIR)/converso"

# Build for all platforms
build-all:
	@echo "🔨 Building Converso CLI for all platforms..."
	@chmod +x scripts/build.sh
	@./scripts/build.sh

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "✅ Clean completed"

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test -v ./...
	@echo "✅ Tests completed"

# Run linter
lint:
	@echo "🔍 Running linter..."
	@if command -v golangci-lint &> /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "⚠️  golangci-lint not found, installing..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.55.2; \
		golangci-lint run ./...; \
	fi
	@echo "✅ Linting completed"

# Format code
format:
	@echo "✨ Formatting code..."
	@go fmt ./...
	@goimports -w . || echo "⚠️  goimports not found, skipping imports formatting"
	@echo "✅ Formatting completed"

# Install CLI locally
install:
	@echo "📦 Installing Converso CLI..."
	@mkdir -p $(BUILD_DIR)
	@go build -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE) -s -w" -o $(BUILD_DIR)/converso ./cmd/converso/
	@if [ "$(GOOS)" = "windows" ]; then \
		cp $(BUILD_DIR)/converso.exe /usr/local/bin/converso.exe 2>/dev/null || echo "⚠️  Cannot install to /usr/local/bin, please copy manually"; \
	else \
		sudo cp $(BUILD_DIR)/converso /usr/local/bin/converso; \
		sudo chmod +x /usr/local/bin/converso; \
	fi
	@echo "✅ Installation completed"
	@echo "💡 Run 'converso --help' to get started"

# Uninstall CLI
uninstall:
	@echo "🗑️  Uninstalling Converso CLI..."
	@if [ "$(GOOS)" = "windows" ]; then \
		rm -f /usr/local/bin/converso.exe 2>/dev/null || echo "⚠️  Cannot remove from /usr/local/bin"; \
	else \
		sudo rm -f /usr/local/bin/converso; \
	fi
	@echo "✅ Uninstallation completed"

# Setup development environment
setup:
	@echo "🔧 Setting up development environment..."
	@go mod tidy
	@go mod download
	@echo "✅ Development environment setup completed"

# Run development server (for testing)
dev:
	@echo "🚀 Starting Converso CLI in development mode..."
	@go run -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)" ./cmd/converso/ $(ARGS)

# Generate documentation
docs:
	@echo "📚 Generating documentation..."
	@mkdir -p docs/generated
	@go doc -all ./... > docs/generated/api.md
	@echo "✅ Documentation generated"

# Security scan
security:
	@echo "🔒 Running security scan..."
	@go vet ./...
	@go list -json -deps ./... | gojq -r '.Packages[] | select(.GoFiles != null) | .Dir' | xargs -I {} go list -json {} | gojq -r '.GoFiles[]' | xargs -I {} go vet {}
	@echo "✅ Security scan completed"

# Performance benchmark
bench:
	@echo "⚡ Running benchmarks..."
	@go test -bench=. -benchmem ./...
	@echo "✅ Benchmarks completed"

# Coverage report
coverage:
	@echo "📊 Generating coverage report..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Docker build
docker:
	@echo "🐳 Building Docker image..."
	@docker build -t converso/cli:$(VERSION) .
	@echo "✅ Docker image built: converso/cli:$(VERSION)"

# Release preparation
release: clean build-all
	@echo "🎉 Release preparation completed"
	@echo "📁 Artifacts available in: $(BUILD_DIR)"
	@ls -la $(BUILD_DIR)/

# CI/CD pipeline simulation
ci: format lint test security
	@echo "✅ CI pipeline completed successfully"

# Development watch mode
watch:
	@echo "👀 Starting development watch mode..."
	@while true; do \
		inotifywait -r -e modify,create,delete . --exclude '.*\.git.*|.*\.idea.*|.*node_modules.*' && \
		echo "🔄 Changes detected, rebuilding..." && \
		make build; \
	done
