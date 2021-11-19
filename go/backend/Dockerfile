# cant use alpine as we need glibc for CGO
FROM golang:1.17 AS deps

WORKDIR /build

# Fetch dependencies.
COPY go.mod go.sum ./
RUN go mod download

FROM deps as dev

COPY backend ./backend
ENTRYPOINT ["/build/backend/entrypoint.sh"]

FROM dev as builder

RUN go build -o main ./backend/main.go

FROM gcr.io/distroless/cc:nonroot
WORKDIR /dist

COPY --from=builder /build/main .
ENTRYPOINT ["/dist/main"]