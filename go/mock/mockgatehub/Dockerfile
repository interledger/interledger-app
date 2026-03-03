# Build stage
FROM golang:1.24-alpine AS builder

ARG BUILD_TIME

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Clean Go cache and build the application
RUN go clean -cache
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-X 'main.buildTime=${BUILD_TIME:-unknown}'" -o mockgatehub ./cmd/mockgatehub

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl tzdata

# Create non-root user
RUN addgroup -g 1000 mockgatehub && \
    adduser -D -u 1000 -G mockgatehub mockgatehub

# Set working directory
WORKDIR /app

# Copy binary and web assets from builder
COPY --from=builder --chown=mockgatehub:mockgatehub /app/mockgatehub .
COPY --from=builder --chown=mockgatehub:mockgatehub /app/web ./web

# Switch to non-root user
USER mockgatehub

EXPOSE 8080

# Health check
HEALTHCHECK --interval=10s --timeout=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

CMD ["./mockgatehub"]
