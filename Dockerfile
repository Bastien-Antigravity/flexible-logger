# === BUILD STAGE ===
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev ca-certificates tzdata

WORKDIR /flexible-logger

# Copy dependency files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /flexible-logger/flexible-logger ./cmd/main

# === RUNTIME STAGE ===
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /flexible-logger

# Copy the binary and assets from the build stage
COPY --from=builder /flexible-logger/flexible-logger /flexible-logger/flexible-logger

# Set the entrypoint
ENTRYPOINT ["/flexible-logger/flexible-logger"]
