# cant use alpine as we need glibc for CGO
FROM public.ecr.aws/docker/library/golang:1.18 AS base

WORKDIR /build

# Fetch dependencies.
COPY go.mod go.sum ./
RUN go mod download
COPY backend ./backend
COPY tigerbeetle_go ./tigerbeetle_go
COPY proto ./proto
ENTRYPOINT ["/build/backend/entrypoint.sh"]

FROM base as builder

RUN go build -o main ./backend/main.go

FROM gcr.io/distroless/cc:nonroot
WORKDIR /dist

COPY --from=builder /build/main .
ENTRYPOINT ["/dist/main"]
