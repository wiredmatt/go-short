# Project variables
BIN_DIR := bin
API_BIN := $(BIN_DIR)/api
CLI_BIN := $(BIN_DIR)/cli

API_MAIN := ./cmd/api/main.go
CLI_MAIN := ./cmd/cli/main.go

# Default target
.PHONY: all
all: build

## API Development target
.PHONY: dev-api
dev-api:
	air -c .air.api.toml

## CLI Development target
.PHONY: dev-cli
dev-cli:
	air -c .air.cli.toml

## Build targets
.PHONY: build
build: build-api build-cli

.PHONY: build-api
build-api:
	@echo "Building API..."
	@mkdir -p $(BIN_DIR)
	go build -o $(API_BIN) $(API_MAIN)

.PHONY: build-cli
build-cli:
	@echo "Building CLI..."
	@mkdir -p $(BIN_DIR)
	go build -o $(CLI_BIN) $(CLI_MAIN)

## Run targets
.PHONY: run-api
run-api:
	go run $(API_MAIN)

.PHONY: run-cli
run-cli:
	go run $(CLI_MAIN)

## Test targets
.PHONY: test-e2e
test-e2e:
	@echo "Running e2e tests..."
	go test -run Integration ./...
	@echo "All tests completed."

.PHONY: test-unit
test-unit:
	@echo "Running unit tests..."
	go test -v -short ./...
	@echo "All tests completed."

## Combined test target
.PHONY: test
test: test-unit test-e2e
	@echo "All tests completed."

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -short -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-benchmark
test-benchmark:
	@echo "Running benchmarks..."
	go test -v -short -bench=. -benchmem ./...

## Clean
.PHONY: clean
clean:
	@echo "Cleaning binaries and test artifacts..."
	rm -rf $(BIN_DIR)
	rm -f coverage.out coverage.html
