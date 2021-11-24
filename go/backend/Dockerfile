# cant use alpine as we need glibc for CGO
FROM golang:1.17 AS base

WORKDIR /build

# Fetch dependencies.
COPY go.mod go.sum ./
RUN go mod download
COPY backend ./backend
ENTRYPOINT ["/build/backend/entrypoint.sh"]

FROM base as builder

RUN go build -o main ./backend/main.go

FROM gcr.io/distroless/cc:nonroot
WORKDIR /dist

COPY --from=builder /build/main .
ENTRYPOINT ["/dist/main"]