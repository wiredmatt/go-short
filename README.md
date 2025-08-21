# go-short

A URL shortener service built with Go, featuring a clean architecture with comprehensive test coverage.

## AI Usage

AI was used in this project to create tests, add comments and generate documentation, speeding up scaffolding time.

## Features

- URL shortening with customizable short codes
- URL resolution with redirects
- In-memory & PostgreSQL storage (extensible to other storage backends)
- RESTful API with Go's servemux
- Fully documented API thanks to huma
- Comprehensive test suite with >80% coverage
- Benchmark tests for performance monitoring
- Centralized configuration management

## TODO

- integrate [opentelemetry](https://opentelemetry.io/docs/languages/go/getting-started/) or [prometheus](https://prometheus.io/docs/guides/go-application/) and [graphana](https://grafana.com/docs/grafana-cloud/monitor-infrastructure/integrations/integration-reference/integration-golang/)

## Dependencies

- golang >= 1.24.4
- air >= v1.62.0 (install with `go install github.com/air-verse/air@latest`)
- make

## Configuration

The application uses environment variables for configuration. Create a `.env` file in the root directory:

```env
# Optional (with defaults)

BASE_URL=https://your-domain.com
PORT=3000
HOST=0.0.0.0
ENVIRONMENT=development
LOG_LEVEL=info
SHORT_CODE_LENGTH=6
DB_CONNECTION_STRING=postgres://user:password@localhost:5432/shortener?sslmode=disable
DB_TYPE=memory # memory | postgres

# Server timeouts
READ_TIMEOUT=30s
WRITE_TIMEOUT=30s
IDLE_TIMEOUT=60s
```

### Environment Variables

| Variable            | Required | Default                                                             | Description                               |
| ------------------- | -------- | ------------------------------------------------------------------- | ----------------------------------------- |
| `BASE_URL`          | Yes      | -                                                                   | Base URL for shortened links              |
| `PORT`              | No       | `3000`                                                              | Server port                               |
| `HOST`              | No       | `0.0.0.0`                                                           | Server host                               |
| `ENVIRONMENT`       | No       | `development`                                                       | Environment (development/production/test) |
| `LOG_LEVEL`         | No       | `info`                                                              | Logging level                             |
| `SHORT_CODE_LENGTH` | No       | `6`                                                                 | Length of generated short codes (3-20)    |
| `DB_TYPE`           | No       | `memory`                                                            | Database type (memory/postgres/redis)     |
| `DB_TYPE`           | No       | `postgres://user:password@localhost:5432/shortener?sslmode=disable` | Database Connection String                |
| `READ_TIMEOUT`      | No       | `30s`                                                               | HTTP read timeout                         |
| `WRITE_TIMEOUT`     | No       | `30s`                                                               | HTTP write timeout                        |
| `IDLE_TIMEOUT`      | No       | `60s`                                                               | HTTP idle timeout                         |

## Commands

### Development

```sh
docker-compose up -d # (optional) if you want to run with postgres, otherwise set DB_TYPE=memory
# note that if you run tests without a postgres instance running, some of them may fail.

make dev-api    # requires air to be installed, use `make run-api` otherwise
```

### Building

```sh
make build-api
```

### Running

```sh
make run-api
```

### Testing

```sh
make test              # Run all tests
make test-short        # Run short tests only
make test-coverage     # Run tests with coverage report
make test-benchmark    # Run benchmark tests
```

### Cleanup

```sh
make clean
```

### K8s

See [./k8s/README.md](./k8s/README.md)

## API Docs

API docs are avaiable at http://localhost:3000/docs

## Testing

The codebase includes comprehensive tests covering:

- **Unit Tests**: Individual component testing with mocks
- **Integration Tests**: End-to-end testing with real components
- **Benchmark Tests**: Performance testing for critical operations
- **Concurrency Tests**: Thread safety verification

### Test Coverage

- **API Layer**: 100% coverage
- **Service Layer**: 100% coverage
- **Storage Layer**: 100% coverage
- **Model Layer**: 100% coverage
- **Config Layer**: 100% coverage

### Running Tests

```sh
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run benchmarks
make test-benchmark
```

### Test Configuration

Tests use a separate `.env.test` file to avoid interfering with your development configuration. The test configuration is automatically loaded during test execution.

## Architecture

The application follows a clean architecture pattern:

```
cmd/
├── api/          # API server entry point
└── cli/          # CLI entry point

internal/
├── api/          # HTTP handlers and routing
├── shortener/    # Business logic for URL shortening
├── storage/      # Data persistence layer
├── model/        # Data models
└── config/       # Configuration management
```

## Configuration Management

The configuration system provides:

- **Environment-based configuration** with sensible defaults
- **Validation** of required settings
- **Type-safe access** to configuration values
- **Test isolation** using separate `.env.test` files
- **Graceful error handling** for missing or invalid values
