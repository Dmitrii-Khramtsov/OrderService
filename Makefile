# github.com/Dmitrii-Khramtsov/orderservice/Makefile
BINARY_NAME := orderservice
GO_LINT := golangci-lint

.PHONY: all build run docker-up docker-down docker-logs lint test clean deps script-up help test-unit test-integration test-all test-coverage

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	@go build -ldflags="-s -w" -o $(BINARY_NAME) ./cmd/server

run: build
	@echo "Running $(BINARY_NAME) locally..."
	@./$(BINARY_NAME)

docker-up:
	@echo "Starting services via Docker Compose..."
	@docker compose up -d --build

docker-down:
	@echo "Stopping services via Docker Compose..."
	@docker compose down

docker-logs:
	@echo "Showing Docker Compose logs..."
	@docker compose logs -f

script-up:
	@echo "Running script via Docker Compose..."
	@docker compose up script

lint:
	@echo "Linting Go files..."
	@$(GO_LINT) run ./...

test: test-unit
	@echo "Running unit tests..."

test-unit:
	@echo "Running unit tests..."
	@go test -v -race -short ./...

test-integration:
	@echo "Running integration tests (requires Docker)..."
	@go test -v -race -tags=integration ./internal/infrastructure/database/...

test-all: test-unit test-integration
	@echo "All tests completed"

test-coverage:
	@echo "Running tests with coverage report..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-coverage-unit:
	@echo "Running unit tests with coverage report..."
	@go test -v -race -short -coverprofile=coverage-unit.out ./...
	@go tool cover -html=coverage-unit.out -o coverage-unit.html
	@echo "Unit tests coverage report generated: coverage-unit.html"

test-coverage-integration:
	@echo "Running integration tests with coverage report..."
	@go test -v -race -tags=integration -coverprofile=coverage-integration.out ./internal/infrastructure/database/...
	@go tool cover -html=coverage-integration.out -o coverage-integration.html
	@echo "Integration tests coverage report generated: coverage-integration.html"

clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)
	@rm -f coverage*.out coverage*.html

deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

help:
	@echo "Available commands:"
	@echo "  make build           - Build the project"
	@echo "  make run             - Run the project locally"
	@echo "  make lint            - Run linter for all Go files"
	@echo "  make test            - Run unit tests only"
	@echo "  make test-unit       - Run unit tests only"
	@echo "  make test-integration - Run integration tests (requires Docker)"
	@echo "  make test-all        - Run all tests (unit + integration)"
	@echo "  make test-coverage   - Run tests with coverage report"
	@echo "  make test-coverage-unit - Run unit tests with coverage report"
	@echo "  make test-coverage-integration - Run integration tests with coverage report"
	@echo "  make clean           - Remove binary file and coverage reports"
	@echo "  make deps            - Install Go dependencies and linter"
	@echo "  make docker-up       - Start services via Docker Compose"
	@echo "  make docker-down     - Stop Docker Compose services"
	@echo "  make docker-logs     - Tail Docker Compose logs"
	@echo "  make script-up       - Run script via Docker Compose"