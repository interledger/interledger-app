# Multi-stage build for mockplaid
FROM golang:1.25-alpine AS builder

WORKDIR /workspace

# Copy only the necessary Go module files first
COPY go/go.mod go/go.sum ./go/

# Download dependencies
RUN cd go && go mod download

# Now copy the rest of the source code
COPY go/mock/mockplaid ./go/mock/mockplaid/

# Build
RUN cd go && CGO_ENABLED=0 GOOS=linux go build -o /workspace/mockplaid ./mock/mockplaid/cmd/mockplaid

FROM alpine:3.19
RUN apk --no-cache add ca-certificates curl

WORKDIR /app
COPY --from=builder /workspace/mockplaid /usr/local/bin/mockplaid

EXPOSE 8080
ENV MOCKPLAID_PORT=8080

HEALTHCHECK --interval=5s --timeout=3s --retries=3 --start-period=5s \
  CMD curl -f http://localhost:8080/health || exit 1

ENTRYPOINT ["/usr/local/bin/mockplaid"]
