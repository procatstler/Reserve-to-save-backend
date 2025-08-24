# R2S Backend Makefile for Go Microservices

.PHONY: help
help: ## Show this help message
	@echo "R2S Backend - Go Microservices"
	@echo "==============================="
	@echo ""
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development
.PHONY: deps
deps: ## Install all dependencies
	@echo "Installing dependencies..."
	cd api-server && go mod download
	cd auth-server && go mod download
	cd core-server && go mod download
	cd query-server && go mod download
	cd batch-server && go mod download
	cd tx-helper && go mod download
	cd event-receiver && go mod download
	cd pkg && go mod download

.PHONY: build
build: ## Build all services
	@echo "Building all services..."
	go build -o bin/api-server ./api-server
	go build -o bin/auth-server ./auth-server
	go build -o bin/core-server ./core-server
	go build -o bin/query-server ./query-server
	go build -o bin/batch-server ./batch-server
	go build -o bin/tx-helper ./tx-helper
	go build -o bin/event-receiver ./event-receiver

.PHONY: build-api
build-api: ## Build API gateway
	go build -o bin/api-server ./api-server

.PHONY: build-auth
build-auth: ## Build auth server
	go build -o bin/auth-server ./auth-server

.PHONY: build-core
build-core: ## Build core server
	go build -o bin/core-server ./core-server

# Running services
.PHONY: run-api
run-api: ## Run API gateway
	go run api-server/main_new.go api-server/gateway.go

.PHONY: run-auth
run-auth: ## Run auth server
	go run auth-server/main.go

.PHONY: run-core
run-core: ## Run core server
	go run core-server/main.go

.PHONY: run-query
run-query: ## Run query server
	go run query-server/main.go

.PHONY: run-batch
run-batch: ## Run batch server
	go run batch-server/main.go

.PHONY: run-tx
run-tx: ## Run tx-helper
	go run tx-helper/main.go

.PHONY: run-event
run-event: ## Run event receiver
	go run event-receiver/main.go

.PHONY: run-all
run-all: ## Run all services
	@echo "Starting all services..."
	@make run-auth &
	@make run-core &
	@make run-query &
	@make run-batch &
	@make run-tx &
	@make run-event &
	@sleep 2
	@make run-api

# Database
.PHONY: db-migrate
db-migrate: ## Run database migrations
	@echo "Running database migrations..."
	psql -U postgres -d r2s_dev -f pkg/db/init-postgres.sql

.PHONY: db-seed
db-seed: ## Seed database with test data
	@echo "Seeding database..."
	go run pkg/db/seed.go

.PHONY: db-reset
db-reset: ## Reset database
	@echo "Resetting database..."
	psql -U postgres -c "DROP DATABASE IF EXISTS r2s_dev;"
	psql -U postgres -c "CREATE DATABASE r2s_dev;"
	@make db-migrate

# Testing
.PHONY: test
test: ## Run all tests
	@echo "Running tests..."
	go test ./api-server/... -v
	go test ./auth-server/... -v
	go test ./core-server/... -v
	go test ./query-server/... -v
	go test ./batch-server/... -v
	go test ./tx-helper/... -v
	go test ./event-receiver/... -v
	go test ./pkg/... -v

.PHONY: test-auth
test-auth: ## Run auth server tests
	go test ./auth-server/... -v

.PHONY: test-core
test-core: ## Run core server tests
	go test ./core-server/... -v

.PHONY: test-integration
test-integration: ## Run integration tests
	go test -tags=integration ./tests/integration/... -v

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	go test -cover ./...

# Docker
.PHONY: docker-build
docker-build: ## Build Docker images
	docker-compose build

.PHONY: docker-up
docker-up: ## Start services with Docker Compose
	docker-compose up -d

.PHONY: docker-down
docker-down: ## Stop Docker Compose services
	docker-compose down

.PHONY: docker-logs
docker-logs: ## View Docker logs
	docker-compose logs -f

# Utilities
.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting code..."
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

.PHONY: lint
lint: ## Run golangci-lint
	@echo "Running linter..."
	golangci-lint run

.PHONY: proto
proto: ## Generate protobuf files
	@echo "Generating protobuf files..."
	protoc --go_out=. --go-grpc_out=. pkg/proto/query/*.proto

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf tmp/
	go clean -cache

.PHONY: install
install: deps build ## Install dependencies and build

# Development shortcuts
.PHONY: dev
dev: ## Start development environment
	@echo "Starting development environment..."
	@make db-migrate
	@make run-all

.PHONY: dev-reset
dev-reset: ## Reset development environment
	@make db-reset
	@make clean
	@make deps
	@make dev

# Monitoring
.PHONY: health
health: ## Check health of all services
	@echo "Checking service health..."
	@curl -s http://localhost:3001/health | jq '.' || echo "API Gateway: DOWN"
	@curl -s http://localhost:3002/health | jq '.' || echo "Auth Server: DOWN"
	@curl -s http://localhost:3003/health | jq '.' || echo "Core Server: DOWN"
	@curl -s http://localhost:3004/health | jq '.' || echo "Query Server: DOWN"
	@curl -s http://localhost:3005/health | jq '.' || echo "Batch Server: DOWN"
	@curl -s http://localhost:3006/health | jq '.' || echo "TX Helper: DOWN"
	@curl -s http://localhost:3007/health | jq '.' || echo "Event Receiver: DOWN"

.PHONY: logs
logs: ## Tail logs from all services
	tail -f logs/*.log

# Production
.PHONY: prod-build
prod-build: ## Build for production
	@echo "Building for production..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/api-server ./api-server
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/auth-server ./auth-server
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/core-server ./core-server
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/query-server ./query-server
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/batch-server ./batch-server
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/tx-helper ./tx-helper
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/event-receiver ./event-receiver

.PHONY: deploy
deploy: prod-build ## Deploy to production
	@echo "Deploying to production..."
	# Add your deployment commands here
	# kubectl apply -f k8s/
	# or
	# docker push ...

# Default target
.DEFAULT_GOAL := help