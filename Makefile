.PHONY: help build run test clean docker-build docker-up docker-down migrate seed lint swagger

# Variables
APP_NAME=hr-management-system
VERSION=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO=go
GOFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Colors
GREEN=\033[0;32m
NC=\033[0m

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# ==================== Build ====================

build: ## Build all binaries
	@echo "$(GREEN)Building API...$(NC)"
	$(GO) build $(GOFLAGS) -o bin/api ./cmd/api
	@echo "$(GREEN)Building Worker...$(NC)"
	$(GO) build $(GOFLAGS) -o bin/worker ./cmd/worker
	@echo "$(GREEN)Building Scheduler...$(NC)"
	$(GO) build $(GOFLAGS) -o bin/scheduler ./cmd/scheduler
	@echo "$(GREEN)Build complete!$(NC)"

build-api: ## Build API only
	$(GO) build $(GOFLAGS) -o bin/api ./cmd/api

build-worker: ## Build Worker only
	$(GO) build $(GOFLAGS) -o bin/worker ./cmd/worker

build-scheduler: ## Build Scheduler only
	$(GO) build $(GOFLAGS) -o bin/scheduler ./cmd/scheduler

# ==================== Run ====================

run: ## Run API server
	$(GO) run ./cmd/api

run-worker: ## Run Worker
	$(GO) run ./cmd/worker

run-scheduler: ## Run Scheduler
	$(GO) run ./cmd/scheduler

run-all: ## Run all services (requires tmux)
	tmux new-session -d -s hr 'make run'
	tmux split-window -h 'make run-worker'
	tmux split-window -v 'make run-scheduler'
	tmux attach -t hr

# ==================== Development ====================

dev: ## Run with hot reload (requires air)
	air -c .air.toml

deps: ## Download dependencies
	$(GO) mod download
	$(GO) mod tidy

lint: ## Run linter
	golangci-lint run ./...

fmt: ## Format code
	$(GO) fmt ./...
	goimports -w .

test: ## Run tests
	$(GO) test -v -race -cover ./...

test-coverage: ## Run tests with coverage report
	$(GO) test -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

swagger: ## Generate Swagger documentation
	swag init -g cmd/api/main.go -o docs

# ==================== Database ====================

migrate: ## Run database migrations
	@echo "$(GREEN)Running migrations...$(NC)"
	psql -h localhost -U postgres -d hr_management -f migrations/001_initial_schema.sql

seed: ## Seed database with sample data
	@echo "$(GREEN)Seeding database...$(NC)"
	psql -h localhost -U postgres -d hr_management -f migrations/002_seed_data.sql

migrate-fresh: ## Drop and recreate database
	@echo "$(GREEN)Dropping database...$(NC)"
	psql -h localhost -U postgres -c "DROP DATABASE IF EXISTS hr_management;"
	psql -h localhost -U postgres -c "CREATE DATABASE hr_management;"
	make migrate
	make seed

# ==================== Docker ====================

docker-build: ## Build Docker image
	docker build -t $(APP_NAME):$(VERSION) -t $(APP_NAME):latest .

docker-up: ## Start all services with Docker Compose
	docker-compose up -d

docker-down: ## Stop all services
	docker-compose down

docker-logs: ## View logs
	docker-compose logs -f

docker-clean: ## Remove all containers, volumes, and images
	docker-compose down -v --rmi all

docker-restart: ## Restart all services
	docker-compose restart

docker-ps: ## Show running containers
	docker-compose ps

# ==================== Production ====================

docker-prod: ## Build and push production image
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) registry.example.com/$(APP_NAME):$(VERSION)
	docker push registry.example.com/$(APP_NAME):$(VERSION)

k8s-deploy: ## Deploy to Kubernetes
	kubectl apply -f deployments/k8s/

k8s-delete: ## Delete from Kubernetes
	kubectl delete -f deployments/k8s/

# ==================== Utilities ====================

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf coverage.*
	rm -rf docs/

gen-ssl: ## Generate self-signed SSL certificate
	mkdir -p deployments/nginx/ssl
	openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
		-keyout deployments/nginx/ssl/server.key \
		-out deployments/nginx/ssl/server.crt \
		-subj "/CN=localhost"

gen-jwt-secrets: ## Generate JWT secrets
	@echo "JWT_ACCESS_SECRET=$$(openssl rand -base64 32)"
	@echo "JWT_REFRESH_SECRET=$$(openssl rand -base64 32)"

check-health: ## Check API health
	curl -s http://localhost:8080/health | jq

check-ready: ## Check API readiness
	curl -s http://localhost:8080/ready | jq
