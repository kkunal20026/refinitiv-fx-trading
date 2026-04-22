.PHONY: help build run test clean docker-build docker-run docker-down lint fmt vet

help:
	@echo "Refinitiv FX Trading Service - Available Commands"
	@echo ""
	@echo "Development:"
	@echo "  make build       - Build the application"
	@echo "  make run         - Run the application"
	@echo "  make test        - Run tests"
	@echo "  make test-unit   - Run unit tests only"
	@echo "  make test-int    - Run integration tests"
	@echo ""
	@echo "Code Quality:"
	@echo "  make fmt         - Format code"
	@echo "  make lint        - Run linter"
	@echo "  make vet         - Run go vet"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run Docker containers"
	@echo "  make docker-down  - Stop Docker containers"
	@echo ""
	@echo "Database:"
	@echo "  make migrate-up   - Run database migrations"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean       - Clean build artifacts"

# Build
build:
	@echo "Building application..."
	@go build -o bin/server ./cmd/server
	@echo "Build complete: bin/server"

# Run
run: build
	@echo "Running application..."
	@./bin/server

# Test
test:
	@echo "Running all tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

test-unit:
	@echo "Running unit tests..."
	@go test -v -race -short ./tests/unit/...

test-int:
	@echo "Running integration tests..."
	@go test -v -race ./tests/integration/... -run Integration

# Code Quality
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@gofmt -s -w .

lint:
	@echo "Running linter..."
	@golangci-lint run ./...

vet:
	@echo "Running go vet..."
	@go vet ./...

# Docker
docker-build:
	@echo "Building Docker image..."
	@docker build -f docker/Dockerfile -t refinitiv-fx-trading:latest .
	@echo "Docker image built: refinitiv-fx-trading:latest"

docker-run:
	@echo "Starting Docker containers..."
	@docker-compose up -d
	@echo "Containers started. Access API at http://localhost:8080"

docker-down:
	@echo "Stopping Docker containers..."
	@docker-compose down
	@echo "Containers stopped"

docker-logs:
	@docker-compose logs -f server

# Database
migrate-up:
	@echo "Running database migrations..."
	@migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/refinitiv_db?sslmode=disable" up

migrate-down:
	@echo "Rolling back database migrations..."
	@migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/refinitiv_db?sslmode=disable" down

# Dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies downloaded"

# Cleanup
clean:
	@echo "Cleaning up..."
	@rm -rf bin/
	@rm -f coverage.out
	@docker-compose down --volumes
	@echo "Cleanup complete"

# CI/CD helpers
ci-test: vet lint test

ci-build: ci-test docker-build
