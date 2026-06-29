# Multi-stage build for mockgatehub
FROM golang:1.25-alpine AS builder

ARG BUILD_TIME

WORKDIR /workspace

# Copy only Go module files first for better layer caching
COPY go/go.mod go/go.sum ./go/

# Download dependencies
RUN cd go && go mod download

# Copy configa (shared library) and mockgatehub source from monorepo
COPY go/configa ./go/configa/
COPY go/mock/mockgatehub ./go/mock/mockgatehub/

# Build
RUN cd go && CGO_ENABLED=0 GOOS=linux go build -ldflags="-X 'main.buildTime=${BUILD_TIME:-unknown}'" -o /workspace/mockgatehub ./mock/mockgatehub/cmd/mockgatehub

# Final stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates curl tzdata

# Create non-root user
RUN addgroup -g 1000 mockgatehub && \
    adduser -D -u 1000 -G mockgatehub mockgatehub

WORKDIR /app

# Copy binary and web assets from builder
COPY --from=builder --chown=mockgatehub:mockgatehub /workspace/mockgatehub .
COPY --from=builder --chown=mockgatehub:mockgatehub /workspace/go/mock/mockgatehub/web ./web

# Switch to non-root user
USER mockgatehub

EXPOSE 8080

# Health check
HEALTHCHECK --interval=10s --timeout=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

CMD ["./mockgatehub"]
