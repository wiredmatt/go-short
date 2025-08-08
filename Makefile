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

## Clean
.PHONY: clean
clean:
	@echo "Cleaning binaries..."
	rm -rf $(BIN_DIR)
