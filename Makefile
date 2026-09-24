.PHONY: build test vet lint run clean help

APP_NAME := server
BUILD_DIR := bin

## help: Show this help message
help:
	@echo "Broadcast Platform - Development Commands"
	@echo ""
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@grep -E '^##' Makefile | sed 's/## //' | column -t -s ':'

## build: Build the server binary
build: vet
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

## test: Run all tests
test:
	go test -v -count=1 ./...

## test-short: Run tests without verbose output
test-short:
	go test -count=1 ./...

## vet: Run go vet
vet:
	go vet ./...

## lint: Run linting (currently same as vet)
lint: vet
	@echo "Lint passed (go vet)"

## run: Run the server
run:
	go run ./cmd/server

## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f broadcast.db

## swagger: Generate Swagger docs (requires swag CLI)
swagger:
	@echo "Swagger requires swag CLI tool:"
	@echo "  go install github.com/swaggo/swag/cmd/swag@latest"
	@echo ""
	@echo "Then run: swag init -g cmd/server/main.go -o docs/swagger"

## deps: Download dependencies
deps:
	go mod download

## tidy: Tidy go modules
tidy:
	go mod tidy

