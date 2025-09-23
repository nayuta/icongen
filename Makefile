.PHONY: help build test test-short test-race test-cover bench clean install lint fmt vet deps check release-test

# Default target
help: ## Show this help message
	@echo "IconGen Development Commands"
	@echo "============================"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Build targets
build: ## Build the icongen binary
	@echo "🔨 Building icongen..."
	go build -o icongen

build-all: ## Build for all platforms
	@echo "🌍 Building for all platforms..."
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o dist/icongen-linux-amd64
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o dist/icongen-linux-arm64
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o dist/icongen-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o dist/icongen-darwin-arm64
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/icongen-windows-amd64.exe
	GOOS=windows GOARCH=arm64 go build -ldflags="-s -w" -o dist/icongen-windows-arm64.exe
	@echo "✅ Built binaries in dist/"

# Test targets
test: ## Run all tests
	@echo "🧪 Running all tests..."
	go test -v

test-short: ## Run short tests (skip slow ones)
	@echo "⚡ Running short tests..."
	go test -v -short

test-race: ## Run tests with race detection
	@echo "🏁 Running tests with race detection..."
	go test -v -race

test-cover: ## Run tests with coverage
	@echo "📊 Running tests with coverage..."
	go test -v -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "📈 Coverage report generated: coverage.html"

test-cover-func: ## Show test coverage by function
	@echo "📊 Generating function coverage report..."
	go test -coverprofile=coverage.out
	go tool cover -func=coverage.out

# Benchmark targets
bench: ## Run all benchmarks
	@echo "⚡ Running benchmarks..."
	go test -bench=. -benchmem

bench-cpu: ## Run CPU benchmarks with profiling
	@echo "🔥 Running CPU benchmarks with profiling..."
	go test -bench=. -benchmem -cpuprofile=cpu.prof
	@echo "🔍 CPU profile saved to cpu.prof"

bench-mem: ## Run memory benchmarks with profiling
	@echo "🧠 Running memory benchmarks with profiling..."
	go test -bench=. -benchmem -memprofile=mem.prof
	@echo "🔍 Memory profile saved to mem.prof"

# Quality targets
fmt: ## Format Go code
	@echo "✨ Formatting code..."
	go fmt ./...

vet: ## Run go vet
	@echo "🔍 Running go vet..."
	go vet ./...

lint: ## Run golangci-lint (requires golangci-lint installed)
	@echo "🔍 Running linter..."
	@which golangci-lint > /dev/null || (echo "❌ golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run

check: fmt vet ## Run all code quality checks
	@echo "✅ All quality checks passed"

# Dependency management
deps: ## Download and tidy dependencies
	@echo "📦 Managing dependencies..."
	go mod download
	go mod tidy

deps-update: ## Update all dependencies
	@echo "🔄 Updating dependencies..."
	go get -u ./...
	go mod tidy

# Development targets
install: ## Install icongen to $GOPATH/bin
	@echo "📥 Installing icongen..."
	go install

run: build ## Build and run icongen with sample args
	@echo "🚀 Running icongen..."
	@echo "Usage: make run ARGS='--help'"
	@echo "       make run ARGS='sample.png'"
	./icongen $(ARGS)

# Release targets
release-test: test-race test-cover check ## Run all tests for release validation
	@echo "🚀 Release validation complete!"

# Example targets
example: build ## Run icongen with example (requires sample image)
	@echo "📷 Running example..."
	@if [ ! -f "example.png" ]; then \
		echo "⚠️  example.png not found. Create a sample PNG file first."; \
		echo "   You can use any PNG image as example.png"; \
		exit 1; \
	fi
	./icongen --clean --trim-percent=75 --radius-percent=25 example.png examples/
	@echo "✅ Example icons generated in examples/"

# Clean targets
clean: ## Clean build artifacts and test files
	@echo "🧹 Cleaning up..."
	rm -f icongen
	rm -f coverage.out coverage.html
	rm -f cpu.prof mem.prof
	rm -rf dist/
	rm -rf examples/
	go clean -testcache
	@echo "✅ Cleaned up"

clean-all: clean ## Clean everything including Go module cache
	@echo "🧹 Deep cleaning..."
	go clean -modcache
	@echo "✅ Deep cleaned"

# Docker targets (optional)
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t icongen:latest .

docker-test: ## Run tests in Docker
	@echo "🐳 Running tests in Docker..."
	docker run --rm icongen:latest make test

# Performance targets
profile-cpu: bench-cpu ## Analyze CPU profile
	@echo "🔍 Analyzing CPU profile..."
	go tool pprof cpu.prof

profile-mem: bench-mem ## Analyze memory profile
	@echo "🔍 Analyzing memory profile..."
	go tool pprof mem.prof

# Development server (for documentation)
docs: ## Generate and serve documentation
	@echo "📚 Generating documentation..."
	@which godoc > /dev/null || go install golang.org/x/tools/cmd/godoc@latest
	@echo "🌐 Documentation server starting at http://localhost:6060/"
	@echo "📖 View package docs at: http://localhost:6060/pkg/github.com/nayuta/icongen/"
	godoc -http=:6060

# Git hooks
git-hooks: ## Install git hooks for development
	@echo "🪝 Installing git hooks..."
	@mkdir -p .git/hooks
	@echo '#!/bin/sh\nmake check' > .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "✅ Git hooks installed (pre-commit: make check)"

# Show project info
info: ## Show project information
	@echo "IconGen Project Information"
	@echo "=========================="
	@echo "Go version:    $$(go version)"
	@echo "Module:        $$(go list -m)"
	@echo "Dependencies:  $$(go list -m all | wc -l) modules"
	@echo "Test files:    $$(find . -name '*_test.go' | wc -l) files"
	@echo "Lines of code: $$(find . -name '*.go' -not -path './vendor/*' | xargs wc -l | tail -1)"
	@echo ""
	@echo "Quick commands:"
	@echo "  make test      - Run all tests"
	@echo "  make bench     - Run benchmarks"
	@echo "  make build     - Build binary"
	@echo "  make check     - Code quality checks"
	@echo ""

# Quick development workflow
dev: deps fmt vet test build ## Complete development workflow
	@echo "🎉 Development workflow complete!"

# CI simulation
ci: clean deps fmt vet test-race test-cover build-all ## Simulate CI pipeline
	@echo "🚀 CI pipeline simulation complete!"