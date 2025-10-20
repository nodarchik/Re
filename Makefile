.PHONY: help build up down logs test clean

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build all Docker images
	docker-compose build

up: ## Start all services
	docker-compose up -d
	@echo "Services are starting..."
	@echo "Backend API: http://localhost:8080"
	@echo "Frontend UI: http://localhost:3000"
	@echo "PostgreSQL: localhost:5432"

down: ## Stop all services
	docker-compose down

logs: ## View logs from all services
	docker-compose logs -f

logs-backend: ## View backend logs
	docker-compose logs -f backend

logs-frontend: ## View frontend logs
	docker-compose logs -f frontend

logs-db: ## View database logs
	docker-compose logs -f db

test: ## Run backend tests
	cd backend && go test -v ./...

test-edge: ## Run edge case test
	cd backend && go test -v ./internal/calculator -run TestCalculator_EdgeCase

clean: ## Remove all containers, volumes, and images
	docker-compose down -v
	docker system prune -f

restart: down up ## Restart all services

status: ## Show status of all services
	docker-compose ps

