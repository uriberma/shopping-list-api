# Shopping List API Makefile
# This Makefile provides convenient commands for development and deployment

# Variables
APP_NAME := shopping-list-api
BINARY_NAME := main
DOCKER_IMAGE := $(APP_NAME):latest
POSTGRES_CONTAINER := shopping_list_postgres

# Go related variables
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt

# Build flags
BUILD_FLAGS := -a -installsuffix cgo
LDFLAGS := -w -s

.PHONY: help build run test test-fast test-verbose test-watch clean deps fmt lint check docker-build docker-run docker-stop db-start db-stop db-reset all

# Default target
all: deps lint test build

# Help target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development commands
build: ## Build the application binary
	@echo "Building $(APP_NAME)..."
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/server
	@echo "Build completed: $(BINARY_NAME)"

run: ## Run the application locally (requires PostgreSQL to be running)
	@echo "Starting $(APP_NAME)..."
	$(GOCMD) run ./cmd/server/main.go

# Testing commands
test: ## Run all tests with coverage report
	@echo "Running tests with coverage..."
	$(GOTEST) -coverprofile=coverage.out $$(go list ./... | grep -v "/migrations")
	$(GOCMD) tool cover -func=coverage.out
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "✅ Tests completed. Coverage report: coverage.html"

test-fast: ## Run tests without coverage (faster)
	@echo "Running tests (fast mode)..."
	$(GOTEST) ./...
	@echo "✅ Tests completed."

test-verbose: ## Run tests with verbose output and coverage
	@echo "Running tests with verbose output..."
	$(GOTEST) -v -coverprofile=coverage.out $$(go list ./... | grep -v "/migrations")
	$(GOCMD) tool cover -func=coverage.out
	@echo "✅ Verbose tests completed."

test-watch: ## Run tests in watch mode (requires entr)
	@echo "Running tests in watch mode..."
	@if command -v entr >/dev/null 2>&1; then \
		find . -name '*.go' | entr -c make test-fast; \
	else \
		echo "entr not installed. Install with: brew install entr (macOS) or apt-get install entr (Ubuntu)"; \
	fi

# Code quality commands
fmt: ## Format and fix all Go code formatting issues
	@echo "Formatting and fixing code..."
	gofmt -w .
	@echo "Code formatting completed."

lint: ## Check and fix all code quality issues (formatting, vetting, linting)
	@echo "Running comprehensive code quality checks and fixes..."
	@echo "1. Fixing formatting..."
	gofmt -w .
	@echo "2. Running go vet..."
	$(GOCMD) vet ./...
	@echo "3. Installing/running linter with auto-fix..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --fix --timeout=5m; \
	elif [ -f "$$(go env GOPATH)/bin/golangci-lint" ]; then \
		$$(go env GOPATH)/bin/golangci-lint run --fix --timeout=5m; \
	else \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.55.2; \
		$$(go env GOPATH)/bin/golangci-lint run --fix --timeout=5m; \
	fi
	@echo "All code quality issues fixed!"

check: ## Check code quality without fixing (for CI/validation)
	@echo "Checking code quality (no fixes)..."
	@echo "1. Checking formatting..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "❌ The following files are not formatted correctly:"; \
		gofmt -l .; \
		echo "Run 'make fmt' to fix formatting issues."; \
		exit 1; \
	else \
		echo "✅ All files are properly formatted."; \
	fi
	@echo "2. Running go vet..."
	$(GOCMD) vet ./...
	@echo "3. Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	elif [ -f "$$(go env GOPATH)/bin/golangci-lint" ]; then \
		$$(go env GOPATH)/bin/golangci-lint run --timeout=5m; \
	else \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.55.2; \
		$$(go env GOPATH)/bin/golangci-lint run --timeout=5m; \
	fi
	@echo "✅ All quality checks passed!"

# Dependency management
deps: ## Download and tidy dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

deps-update: ## Update all dependencies
	@echo "Updating dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy

# Database commands
db-start: ## Start PostgreSQL database using Docker
	@echo "Starting PostgreSQL database..."
	docker-compose up postgres -d
	@echo "Waiting for database to be ready..."
	@sleep 5

db-stop: ## Stop PostgreSQL database
	@echo "Stopping PostgreSQL database..."
	docker-compose stop postgres

db-reset: ## Reset database (stop, remove, and start fresh)
	@echo "Resetting database..."
	docker-compose down postgres
	docker volume rm shopping-list-api_postgres_data 2>/dev/null || true
	docker-compose up postgres -d
	@echo "Database reset completed"

db-logs: ## Show database logs
	docker-compose logs -f postgres

# Migration commands
migrate-up: ## Run all pending migrations
	@echo "Running database migrations..."
	$(GOCMD) run ./cmd/migrator/main.go -action=up -migrations-path=./cmd/migrator/migrations

migrate-down: ## Rollback one migration
	@echo "Rolling back one migration..."
	$(GOCMD) run ./cmd/migrator/main.go -action=down -migrations-path=./cmd/migrator/migrations

migrate-version: ## Show current migration version
	@echo "Checking migration version..."
	$(GOCMD) run ./cmd/migrator/main.go -action=version -migrations-path=./cmd/migrator/migrations

migrate-force: ## Force migration to specific version (use VERSION=n)
	@echo "Forcing migration to version $(VERSION)..."
	$(GOCMD) run ./cmd/migrator/main.go -action=force -force-version=$(VERSION) -migrations-path=./cmd/migrator/migrations

migrate-drop: ## Drop all database tables (DANGEROUS)
	@echo "WARNING: This will drop all database tables!"
	$(GOCMD) run ./cmd/migrator/main.go -action=drop

migrate-create: ## Create new migration files (use NAME=migration_name)
	@if [ -z "$(NAME)" ]; then echo "Usage: make migrate-create NAME=migration_name"; exit 1; fi
	@echo "Creating migration files for $(NAME)..."
	@TIMESTAMP=$$(date +%s); \
	NEXT_VERSION=$$(printf "%06d" $$((TIMESTAMP % 1000000))); \
	touch cmd/migrator/migrations/$${NEXT_VERSION}_$(NAME).up.sql; \
	touch cmd/migrator/migrations/$${NEXT_VERSION}_$(NAME).down.sql; \
	echo "Created cmd/migrator/migrations/$${NEXT_VERSION}_$(NAME).up.sql"; \
	echo "Created cmd/migrator/migrations/$${NEXT_VERSION}_$(NAME).down.sql"

# Docker commands
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE) .

docker-run: ## Run the application using Docker Compose
	@echo "Starting application with Docker Compose..."
	docker-compose up -d

docker-stop: ## Stop Docker containers
	@echo "Stopping Docker containers..."
	docker-compose down

docker-logs: ## Show application logs
	docker-compose logs -f api

docker-rebuild: ## Rebuild and restart Docker containers
	@echo "Rebuilding Docker containers..."
	docker-compose down
	docker-compose build --no-cache
	docker-compose up -d

# Utility commands
clean: ## Clean build artifacts and test cache
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	$(GOMOD) tidy

install-tools: ## Install development tools
	@echo "Installing development tools..."
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# API testing commands
api-test: ## Test API endpoints (requires the service to be running)
	@echo "Testing API endpoints..."
	@echo "Health check:"
	@curl -s http://localhost:8080/health | jq . || echo "Service not running or jq not installed"
	@echo "\nCreating a test shopping list:"
	@curl -s -X POST http://localhost:8080/api/v1/lists \
		-H "Content-Type: application/json" \
		-d '{"name":"Test List","description":"Test Description"}' | jq . || echo "Failed to create list"

# Development workflow commands
dev-setup: deps db-start ## Set up development environment
	@echo "Waiting for database to be ready..."
	@sleep 5
	@$(MAKE) migrate-up
	@echo "Development environment setup completed!"
	@echo "Run 'make run-dev' to start the application"

dev-reset: clean db-reset deps ## Reset development environment
	@echo "Waiting for database to be ready..."
	@sleep 5
	@$(MAKE) migrate-up
	@echo "Development environment reset completed!"

run-dev: db-start ## Start database, run migrations, and run the application
	@echo "Starting development environment..."
	@sleep 5  # Wait for database to be ready
	@$(MAKE) migrate-up
	$(GOCMD) run ./cmd/server/main.go

# Production commands
build-prod: ## Build production binary
	@echo "Building production binary..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/server

deploy-docker: docker-build docker-run ## Build and deploy using Docker

# Show project status
status: ## Show project status
	@echo "=== Project Status ==="
	@echo "Go version: $$(go version)"
	@echo "Dependencies status:"
	@$(GOMOD) verify
	@echo "Database status:"
	@docker-compose ps postgres 2>/dev/null || echo "Database not running"
	@echo "Application status:"
	@curl -s http://localhost:8080/health >/dev/null && echo "API is running" || echo "API is not running"
