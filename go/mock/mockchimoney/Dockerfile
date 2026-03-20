# Multi-stage build for mockchimoney
FROM golang:1.25-alpine AS builder

ARG BUILD_TIME

WORKDIR /workspace

# Copy only Go module files first for better layer caching
COPY go/go.mod go/go.sum ./go/

# Download dependencies
RUN cd go && go mod download

# Copy mockchimoney source from monorepo
COPY go/mock/mockchimoney ./go/mock/mockchimoney/

# Build
RUN cd go && CGO_ENABLED=0 GOOS=linux go build -ldflags="-X 'main.buildTime=${BUILD_TIME:-unknown}'" -o /workspace/mockchimoney ./mock/mockchimoney/cmd/mockchimoney

# Final stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates curl tzdata

# Create non-root user
RUN addgroup -g 1000 mockchimoney && \
    adduser -D -u 1000 -G mockchimoney mockchimoney

WORKDIR /app

# Copy binary from builder
COPY --from=builder --chown=mockchimoney:mockchimoney /workspace/mockchimoney .

# Switch to non-root user
USER mockchimoney

ENV MOCKCHIMONEY_PORT=41800
EXPOSE 41800

# Health check
HEALTHCHECK --interval=10s --timeout=5s --retries=3 \
  CMD curl -f http://localhost:${MOCKCHIMONEY_PORT}/health || exit 1

CMD ["./mockchimoney"]
