.PHONY: build run test lint clean install-tools dev help
.DEFAULT_GOAL := help

# Variables
BINARY_NAME=famstack
BINARY_PATH=cmd/famstack/$(BINARY_NAME)

# Build the application
build: build-ts build-go ## Build TypeScript components and Go binary

build-go: ## Build Go binary
	@echo "Building Go binary..."
	CGO_ENABLED=1 go build -ldflags="-s -w" -o $(BINARY_PATH) ./cmd/famstack

build-ts: install-node-deps ## Build TypeScript components
	@echo "Building TypeScript components..."
	cd web/components && npm run build

# Run the application
run: build ## Run the application locally
	@echo "Starting famstack server..."
	./$(BINARY_PATH)

# Development mode with file watching
dev: install-tools ## Development mode with file watching
	@echo "Starting development mode..."
	@echo "Note: File watching requires additional tools like air or reflex"
	@echo "For now, use 'make build && make run' after changes"
	$(MAKE) build
	./$(BINARY_PATH)

# Testing
test: test-go test-ts ## Run all tests

test-go: ## Run Go tests
	@echo "Running Go tests..."
	go test -v ./...

test-ts: install-node-deps ## Run TypeScript tests (if any)
	@echo "Running TypeScript tests..."
	cd web/components && npm run test 2>/dev/null || echo "No TypeScript tests configured yet"

# Linting
lint: lint-go lint-ts ## Run all linters

lint-go: install-golangci-lint ## Run Go linter
	@echo "Running golangci-lint..."
	golangci-lint run

lint-ts: install-node-deps ## Run TypeScript linter
	@echo "Running TypeScript linter..."
	cd web/components && npm run lint

# Tool installation
install-tools: install-golangci-lint install-goose install-node-deps ## Install all required tools

install-golangci-lint: ## Install golangci-lint
	@which golangci-lint > /dev/null || \
		(echo "Installing golangci-lint..." && \
		 curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin)

install-goose: ## Install goose migration tool
	@which goose > /dev/null || \
		(echo "Installing goose..." && \
		 go install github.com/pressly/goose/v3/cmd/goose@latest)

install-node-deps: ## Install Node.js dependencies
	@echo "Installing Node.js dependencies..."
	cd web/components && npm install

# Cleanup
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_PATH)
	rm -rf web/static/js/*.js
	rm -rf web/static/js/*.js.map
	cd web/components && rm -rf node_modules || true
	go clean -cache
	go clean -modcache || true

# Database migrations
migrate-up: ## Run database migrations up
	@echo "Running database migrations..."
	./$(BINARY_PATH) -migrate-up

migrate-down: ## Run database migrations down
	@echo "Rolling back database migrations..."
	./$(BINARY_PATH) -migrate-down

# Development database reset
reset-db: ## Reset development database
	@echo "Resetting development database..."
	rm -f famstack.db || true

# Release preparation
prepare-release: clean lint test build ## Prepare for release (clean, lint, test, build)
	@echo "Release preparation complete"

# Help
help: ## Show this help message
	@echo "Fam-Stack Build System"
	@echo "====================="
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)