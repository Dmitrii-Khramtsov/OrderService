BINARY_NAME := l0
SOURCE_DIR := .
GO_LINT := golangci-lint

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BINARY_NAME) ./cmd

run: build
	@echo "Running $(BINARY_NAME)..."
	@./$(BINARY_NAME)

lint:
	@echo "Linting Go files..."
# ./... — специальный синтаксис Go, означающий "все пакеты в текущей директории и её подпапках рекурсивно".
	@$(GO_LINT) run ./...

test:
	@echo "Running tests..."
	@go test -v -race -cover ./...

clean:
	@echo "Cleaning..."
	@rm -f $(BINARY_NAME)

# Установка зависимостей (go mod tidy + линтер)
deps:
	@echo "Installing dependencies..."
	@go mod tidy
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

help:
	@echo "Available commands:"
	@echo "  make build    - Build the project"
	@echo "  make run      - Build and run the project"
	@echo "  make lint     - Run linter for all Go files"
	@echo "  make test     - Run all tests with race detection"
	@echo "  make clean    - Remove binary file"
	@echo "  make deps     - Install Go dependencies and linter"
	@echo "  make help     - Show this help"
