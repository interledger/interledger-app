# cant use alpine as we need glibc for CGO
FROM public.ecr.aws/docker/library/golang:1.23.6 AS base

WORKDIR /build

# Fetch dependencies.
COPY go.mod go.sum ./
RUN go mod download
COPY backend ./backend
COPY env ./env
COPY log ./log
FROM base as builder

RUN go build -o main ./backend/providers/pti/external/server/main.go
ENTRYPOINT ["/build/main"]

FROM gcr.io/distroless/cc:nonroot
WORKDIR /dist

COPY --from=builder /build/main .
ENTRYPOINT ["/dist/main"]
