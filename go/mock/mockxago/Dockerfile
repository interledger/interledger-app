# Multi-stage build for mockxago
FROM golang:1.25-alpine AS builder

WORKDIR /workspace

# Copy only the necessary Go module files first
COPY go/go.mod go/go.sum ./go/

# Download dependencies
RUN cd go && go mod download

# Copy configa (shared library) and mockxago source
COPY go/configa ./go/configa/
COPY go/mock/mockxago ./go/mock/mockxago/

# Build
RUN cd go && CGO_ENABLED=0 GOOS=linux go build -o /workspace/mockxago ./mock/mockxago/cmd/mockxago

# Copy web assets to a known location for the final stage
RUN cp -r /workspace/go/mock/mockxago/web /workspace/web

FROM alpine:3.19
RUN apk --no-cache add ca-certificates curl

WORKDIR /app
COPY --from=builder /workspace/mockxago /usr/local/bin/mockxago
COPY --from=builder /workspace/web ./web/

EXPOSE 8080
ENV XAGO_MOCK_PORT=8080 \
    XAGO_API_PUBLIC_KEY=test-public-key \
    XAGO_API_SECRET=test-secret

HEALTHCHECK --interval=5s --timeout=3s --retries=3 --start-period=5s \
  CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["/usr/local/bin/mockxago"]




