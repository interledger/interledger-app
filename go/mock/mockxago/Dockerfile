# Multi-stage build for mockxago
FROM golang:1.25-alpine AS builder

WORKDIR /workspace

# Copy only the necessary Go module files first
COPY go/go.mod go/go.sum ./go/

# Download dependencies
RUN cd go && go mod download

# Now copy the rest of the source code
COPY go/mock/mockxago ./go/mock/mockxago/

# Build
RUN cd go && CGO_ENABLED=0 GOOS=linux go build -o /workspace/mockxago ./mock/mockxago/cmd/mockxago

FROM alpine:3.19
RUN apk --no-cache add ca-certificates curl

WORKDIR /app
COPY --from=builder /workspace/mockxago /usr/local/bin/mockxago

EXPOSE 8080
ENV XAGO_MOCK_PORT=8080 \
    XAGO_API_PUBLIC_KEY=test-public-key \
    XAGO_API_SECRET=test-secret

HEALTHCHECK --interval=5s --timeout=3s --retries=3 --start-period=5s \
  CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["/usr/local/bin/mockxago"]




