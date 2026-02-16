# Multi-stage build for mockxago
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/mockxago ./cmd/mockxago

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/mockxago /usr/local/bin/mockxago
COPY --from=builder /app/web ./web
EXPOSE 8080
ENV XAGO_MOCK_PORT=8080 \
    XAGO_API_PUBLIC_KEY=test-public-key \
    XAGO_API_SECRET=test-secret
ENTRYPOINT ["/usr/local/bin/mockxago"]
