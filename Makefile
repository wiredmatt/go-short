# Project variables
BIN_DIR := bin
API_BIN := $(BIN_DIR)/api

API_MAIN := ./cmd/api/main.go

# Default target
.PHONY: all
all: build

.PHONY: run_api
run_api:
	go run cmd/api/main.go

# Run API with just hot reload (dev)
.PHONY: run_dev_api
run_dev_api:
	air -c .air.api.toml

# Run API with hot reload + postgres, prometheus and grafana (dev)
.PHONY: run_compose_dev_api
run_compose_dev_api:
	docker compose -f ./.docker/docker-compose.dev.yaml up

# Run API with postgres, prometheus and grafana (prod)
.PHONY: run_compose_prod_api
run_compose_prod_api:
	docker compose -f .docker/docker-compose.yaml up -d --build
	
## Build targets
.PHONY: build
build: build_api

.PHONY: build_api
build_api:
	@echo "Building API..."
	@mkdir -p $(BIN_DIR)
	go build -o $(API_BIN) $(API_MAIN)

## Test targets
.PHONY: test_e2e
test_e2e:
	@echo "Running e2e tests..."
	go clean -testcache
	go test -v -p 1 -run Integration ./...
	@echo "e2e tests completed."

.PHONY: test_unit
test_unit:
	@echo "Running unit tests..."
	go test -v -short ./...
	@echo "unit tests completed."

## Combined test target
.PHONY: test
test: test_unit test_e2e
	@echo "All tests completed."

.PHONY: test_coverage
test_coverage:
	@echo "Running tests with coverage..."
	go test -v -short -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test_benchmark
test_benchmark:
	@echo "Running benchmarks..."
	go test -v -short -bench=. -benchmem ./...

## Clean
.PHONY: clean
clean:
	@echo "Cleaning binaries and test artifacts..."
	rm -rf $(BIN_DIR)
	rm -f coverage.out coverage.html
