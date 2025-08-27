# github.com/Dmitrii-Khramtsov/orderservice/Makefile
BINARY_NAME := orderservice
GO_LINT := golangci-lint

.PHONY: all build run docker-up docker-down docker-logs lint test clean deps script-up help

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

test:
	@echo "Running tests..."
	@go test -v -race -cover ./...

clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)

deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

help:
	@echo "Available commands:"
	@echo "  make build           - Build the project"
	@echo "  make run             - Run the project locally"
	@echo "  make lint            - Run linter for all Go files"
	@echo "  make test            - Run all tests with race detection"
	@echo "  make clean           - Remove binary file"
	@echo "  make deps            - Install Go dependencies and linter"
	@echo "  make docker-up       - Start services via Docker Compose"
	@echo "  make docker-down     - Stop Docker Compose services"
	@echo "  make docker-logs     - Tail Docker Compose logs"
	@echo "  make script-up       - Run script via Docker Compose"
