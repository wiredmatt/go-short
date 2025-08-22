# --- Build stage ---
FROM golang:1.24.4 AS builder

WORKDIR /app

# Cache modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build API binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api/main.go

# --- Runtime stage ---
FROM gcr.io/distroless/base-debian12

WORKDIR /app

# Copy binary from builder
COPY --from=builder /api /app/api

EXPOSE 4000

# Run
ENTRYPOINT ["/app/api"]