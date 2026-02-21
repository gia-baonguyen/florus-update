.PHONY: build run dev test clean docker-build docker-up docker-down migrate seed help

# Variables
APP_NAME := florus-api
MAIN_PATH := ./cmd/api
DOCKER_COMPOSE := docker-compose

# Colors for help
CYAN := \033[36m
RESET := \033[0m

help: ## Show this help
	@echo "$(CYAN)Florus Backend - Available Commands$(RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CYAN)%-15s$(RESET) %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building $(APP_NAME)..."
	go build -o bin/$(APP_NAME) $(MAIN_PATH)

run: build ## Build and run the application
	@echo "Running $(APP_NAME)..."
	./bin/$(APP_NAME)

dev: ## Run with hot reload (requires air)
	@echo "Running in development mode..."
	air

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -f florus.db

tidy: ## Tidy and verify dependencies
	@echo "Tidying dependencies..."
	go mod tidy
	go mod verify

lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

# Docker commands
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):latest .

docker-up: ## Start all services with Docker Compose
	@echo "Starting services..."
	$(DOCKER_COMPOSE) up -d

docker-up-sqlite: ## Start API with SQLite (no Oracle)
	@echo "Starting API with SQLite..."
	$(DOCKER_COMPOSE) --profile sqlite up -d api-sqlite

docker-down: ## Stop all services
	@echo "Stopping services..."
	$(DOCKER_COMPOSE) down

docker-logs: ## View logs
	$(DOCKER_COMPOSE) logs -f

docker-logs-api: ## View API logs only
	$(DOCKER_COMPOSE) logs -f api

# Database commands
migrate: ## Run database migrations
	@echo "Running migrations..."
	go run $(MAIN_PATH) migrate

seed: ## Seed database with sample data
	@echo "Seeding database..."
	go run $(MAIN_PATH) seed

# Development setup
setup: ## Setup development environment
	@echo "Setting up development environment..."
	go mod download
	cp .env.example .env
	@echo "Setup complete! Edit .env file and run 'make dev'"

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed!"
