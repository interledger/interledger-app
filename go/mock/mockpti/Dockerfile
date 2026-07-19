# Multi-stage build for mockpti
FROM golang:1.25-alpine AS builder

WORKDIR /workspace

# Copy only the necessary Go module files first
COPY go/go.mod go/go.sum ./go/

# Download dependencies
RUN cd go && go mod download

# Copy configa (shared library) and mockpti source
COPY go/configa ./go/configa/
COPY go/mock/mockpti ./go/mock/mockpti/

# Build
RUN cd go && CGO_ENABLED=0 GOOS=linux go build -o /workspace/mockpti ./mock/mockpti/cmd/mockpti

FROM alpine:3.19
RUN apk --no-cache add ca-certificates curl

WORKDIR /app
COPY --from=builder /workspace/mockpti /usr/local/bin/mockpti

EXPOSE 8080
ENV MOCKPTI_PORT=8080 \
    MOCKPTI_CLIENT_ID=test-client-id

HEALTHCHECK --interval=5s --timeout=3s --retries=3 --start-period=5s \
  CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["/usr/local/bin/mockpti"]
